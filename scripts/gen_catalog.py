#!/usr/bin/env python3
"""Generate internal/sky/catalog_gen.go: the constellations that are not in
the hand-curated catalog.go, built from the d3-celestial data files.

Usage:
    python3 scripts/gen_catalog.py DATA_DIR > internal/sky/catalog_gen.go

DATA_DIR must hold these files from d3-celestial (BSD-3-Clause, by Olaf
Frohn) at commit D3_COMMIT below, from its data/ directory:

    constellations.json        names and IAU abbreviations
    constellations.lines.json  stick figures (after the IAU constellation charts)
    stars.6.json               stars to magnitude 6 (from XHIP)
    starnames.json             proper names and designations

Each figure vertex is matched to the nearest star in stars.6.json for its
magnitude and name. The best-viewing month is the month the figure's center
crosses the meridian at 21:00 local time; the hemisphere follows the center's
declination (|dec| < 10 is "Equatorial"). The English names are short forms
of the meanings in the IAU list.
"""

import json
import math
import os
import sys

D3_COMMIT = "7e720a3de062059d4c5400a379146a601d9010e0"

# Constellations already in catalog.go, curated by hand.
CURATED = {
    "Ori", "UMa", "UMi", "Cas", "Cyg", "Lyr", "Aql", "Sco", "Sgr", "Leo",
    "Tau", "Gem", "CMa", "Boo", "Cru", "And", "Peg", "Per", "Aur", "CrB",
    "Cep", "Dra", "Cnc", "Ari",
}

# Short English names, from the meanings in the IAU constellation list.
COMMON = {
    "Ant": "The Air Pump", "Aps": "The Bird of Paradise", "Aqr": "The Water Bearer",
    "Ara": "The Altar", "Cae": "The Chisel", "Cam": "The Giraffe",
    "CVn": "The Hunting Dogs", "CMi": "The Lesser Dog", "Cap": "The Sea Goat",
    "Car": "The Keel", "Cen": "The Centaur", "Cet": "The Sea Monster",
    "Cha": "The Chameleon", "Cir": "The Drawing Compasses", "Col": "The Dove",
    "Com": "Berenice's Hair", "CrA": "The Southern Crown", "Crv": "The Crow",
    "Crt": "The Cup", "Del": "The Dolphin", "Dor": "The Dolphinfish",
    "Equ": "The Pony", "Eri": "The River", "For": "The Furnace",
    "Gru": "The Crane", "Her": "Hercules the Hero", "Hor": "The Pendulum Clock",
    "Hya": "The Water Snake", "Hyi": "The Lesser Water Snake", "Ind": "The Indian",
    "Lac": "The Lizard", "LMi": "The Lesser Lion", "Lep": "The Hare",
    "Lib": "The Scales", "Lup": "The Wolf", "Lyn": "The Lynx",
    "Men": "Table Mountain", "Mic": "The Microscope", "Mon": "The Unicorn",
    "Mus": "The Fly", "Nor": "The Carpenter's Level", "Oct": "The Octant",
    "Oph": "The Serpent Bearer", "Pav": "The Peacock", "Phe": "The Phoenix",
    "Pic": "The Painter", "Psc": "The Fishes", "PsA": "The Southern Fish",
    "Pup": "The Poop Deck", "Pyx": "The Mariner's Compass", "Ret": "The Reticle",
    "Sge": "The Arrow", "Scl": "The Sculptor", "Sct": "The Shield",
    "Ser": "The Serpent", "Sex": "The Sextant", "Tel": "The Telescope",
    "Tri": "The Triangle", "TrA": "The Southern Triangle", "Tuc": "The Toucan",
    "Vel": "The Sails", "Vir": "The Maiden", "Vol": "The Flying Fish",
    "Vul": "The Little Fox",
}

# IAU spellings where the data differs, and one name for the two-part Serpens.
LATIN = {"CrA": "Corona Australis", "Ser": "Serpens"}

MONTHS = ["January", "February", "March", "April", "May", "June", "July",
          "August", "September", "October", "November", "December"]

GREEK = {
    "α": "Alpha", "β": "Beta", "γ": "Gamma", "δ": "Delta", "ε": "Epsilon",
    "ζ": "Zeta", "η": "Eta", "θ": "Theta", "ι": "Iota", "κ": "Kappa",
    "λ": "Lambda", "μ": "Mu", "ν": "Nu", "ξ": "Xi", "ο": "Omicron", "π": "Pi",
    "ρ": "Rho", "σ": "Sigma", "τ": "Tau", "υ": "Upsilon", "φ": "Phi",
    "χ": "Chi", "ψ": "Psi", "ω": "Omega",
}


def ra_hours(lon):
    return (lon if lon >= 0 else lon + 360.0) / 15.0


def star_labels(entry, abbr, mag):
    """(proper name, Bayer label like "Alpha Ant" for stars brighter than
    magnitude 4), either possibly ""."""
    if not entry:
        return "", ""
    bayer_label = ""
    bayer = entry.get("bayer", "")
    if bayer and mag < 4.0 and bayer[0] in GREEK:
        bayer_label = f"{GREEK[bayer[0]]}{bayer[1:]} {abbr}"
    return entry.get("name", ""), bayer_label


def best_month(ra_center):
    # At 21:00 local solar time the meridian is at sun RA + 9h. The Sun's RA
    # is 0h at the March equinox (about March 20) and grows ~2h a month.
    sun_ra = (ra_center - 9.0) % 24.0
    day_of_year = 79 + sun_ra / 24.0 * 365.25  # day 79 is March 20
    month = int(((day_of_year % 365.25) / 365.25) * 12.0) % 12
    return MONTHS[month]


def hemisphere(dec):
    if dec >= 10:
        return "Northern"
    if dec <= -10:
        return "Southern"
    return "Equatorial"


def center(points):
    # Mean of unit vectors, so figures that straddle RA 0h average correctly.
    x = y = z = 0.0
    for ra, dec in points:
        a, d = math.radians(ra * 15), math.radians(dec)
        x += math.cos(d) * math.cos(a)
        y += math.cos(d) * math.sin(a)
        z += math.sin(d)
    ra = math.degrees(math.atan2(y, x)) / 15 % 24
    dec = math.degrees(math.atan2(z, math.hypot(x, y)))
    return ra, dec


def main():
    data = sys.argv[1]
    load = lambda f: json.load(open(os.path.join(data, f), encoding="utf-8"))
    names = {f["id"]: f["properties"]["name"] for f in load("constellations.json")["features"]}
    names.update(LATIN)
    lines = load("constellations.lines.json")["features"]
    stars = [(f["geometry"]["coordinates"], f["properties"]["mag"], str(f["id"]))
             for f in load("stars.6.json")["features"]]
    starnames = load("starnames.json")

    figures = {}
    for f in lines:  # Serpens comes as two features (Caput and Cauda)
        figures.setdefault(f["id"], []).extend(f["geometry"]["coordinates"])

    out = []
    for abbr in sorted(figures, key=lambda a: names[a]):
        if abbr in CURATED:
            continue
        verts, index, segs = [], {}, []
        for path in figures[abbr]:
            prev = None
            for lon, lat in path:
                key = (round(lon, 4), round(lat, 4))
                if key not in index:
                    index[key] = len(verts)
                    verts.append(key)
                cur = index[key]
                if prev is not None and prev != cur:
                    seg = (prev, cur)
                    if seg not in segs and (cur, prev) not in segs:
                        segs.append(seg)
                prev = cur
        stars_out = []
        for lon, lat in verts:
            best, best_d = None, 0.05
            for (slon, slat), mag, hip in stars:
                d = math.hypot((slon - lon) * math.cos(math.radians(lat)), slat - lat)
                if d < best_d:
                    best, best_d = (mag, hip), d
            mag = float(best[0]) if best else 6.0
            proper, bayer = star_labels(starnames.get(best[1]) if best else None, abbr, mag)
            stars_out.append(([proper, bayer], ra_hours(lon), lat, mag))
        # A few proper names repeat within a figure in the source data, so
        # they cannot all be right: those stars get their Bayer label instead.
        propers = [st[0][0] for st in stars_out]
        for i, (labels, ra, dec, mag) in enumerate(stars_out):
            proper, bayer = labels
            name = proper if proper and propers.count(proper) == 1 else bayer
            stars_out[i] = (name, ra, dec, mag)
        ra_c, dec_c = center([(s[1], s[2]) for s in stars_out])
        out.append((names[abbr], abbr, COMMON[abbr], hemisphere(dec_c), best_month(ra_c), stars_out, segs))

    w = sys.stdout.write
    w("// Code generated by scripts/gen_catalog.py from d3-celestial data; DO NOT EDIT.\n")
    w(f"// Source: https://github.com/ofrohn/d3-celestial/tree/{D3_COMMIT}/data\n")
    w("// (BSD-3-Clause, Copyright (c) 2015, Olaf Frohn; see THIRD_PARTY_NOTICES.md).\n\n")
    w("package sky\n\n")
    w("// generated holds the constellations that catalog.go does not curate by hand.\n")
    w("var generated = []Constellation{\n")
    for latin, abbr, common, hemi, month, sts, segs in out:
        w("\t{\n")
        w(f"\t\tLatin: {json.dumps(latin)}, Abbr: {json.dumps(abbr)}, Common: {json.dumps(common)},\n")
        w(f"\t\tHemisphere: {json.dumps(hemi)}, BestMonth: {json.dumps(month)},\n")
        w("\t\tStars: []Star{\n")
        for name, ra, dec, mag in sts:
            w(f"\t\t\t{{{json.dumps(name, ensure_ascii=False)}, {ra:.3f}, {dec:.3f}, {mag:.2f}}},\n")
        w("\t\t},\n")
        w("\t\tLines: [][2]int{")
        w(", ".join(f"{{{a}, {b}}}" for a, b in segs))
        w("},\n")
        w("\t},\n")
    w("}\n")

    # Calibration report on stderr: the rules against the curated entries.
    for abbr in sorted(CURATED):
        pts = []
        for path in figures[abbr]:
            pts += [(ra_hours(lon), lat) for lon, lat in path]
        ra_c, dec_c = center(pts)
        sys.stderr.write(f"{abbr}: {hemisphere(dec_c):10} {best_month(ra_c):9} (dec {dec_c:+.1f})\n")


main()

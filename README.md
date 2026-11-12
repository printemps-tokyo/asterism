# asterism

Draw constellations in your terminal.

Type a constellation's name and asterism plots its bright stars (bigger, bolder,
brighter-colored glyphs for brighter stars), connects them with the traditional
stick-figure lines, labels the named stars, and prints a short info panel with
the Latin name, IAU abbreviation, brightest star, and best viewing season. All
88 IAU constellations are included. Give it your latitude and longitude and it
draws the figure as it stands in your sky at that moment, says where to look,
and when it rises and sets, or lists everything that is up.

It is a small, single-binary CLI with no dependencies beyond the Go standard
library. Color is emitted as ANSI escape codes; terminal size is read from the
`COLUMNS`/`LINES` environment variables (with an 80x40 fallback).

## Install

With Go 1.23 or newer:

```bash
go install github.com/printemps-tokyo/asterism@latest
```

Prebuilt binaries for Linux and macOS (amd64 and arm64) are attached to each
[tagged release](https://github.com/printemps-tokyo/asterism/releases).

## Quickstart

```bash
asterism orion            # draw Orion
asterism "ursa major"     # multi-word names need quotes
asterism uma              # or use the IAU abbreviation
asterism --list           # list every available constellation
asterism --all            # render them all
asterism cru --ascii --no-color   # plain ASCII, no color

# From Tokyo, now or at a given moment
asterism orion --lat 35.68 --lon 139.77
asterism orion --lat 35.68 --lon 139.77 --time 2026-01-15T21:00
asterism --visible --lat 35.68 --lon 139.77 --time 2026-01-15T21:00
```

## Example

```text
$ asterism orion

                               ─◦ Meissa
                           ────  ╲
                        ───       ╲╲
           Betelgeuse ★───────      ╲
                       ╲      ───────✦ Bellatrix
                        ╲           ╱
                        ╲           ╱
                         ╲         ╱
                          ╲       ╱
                           ╲      ╱
                           ╲     ● Mintaka
                            ╲  ✦─ ╲
                     Alnitak ✦─    ╲
                            │       ╲
                            │        ╲
                           │          ╲
                           │           ╲
                          │             ╲
                          │              ╲
                         │                ★ Rigel
                         ● Saiph

Orion (Ori) - The Hunter
Brightest star: Rigel (mag +0.18)
Stars plotted:  8
Best viewing:   January, Equatorial Hemisphere
```

```text
$ asterism cygnus

                   ✦ Deneb               ─● Delta Cyg
                    ╲╲                ───
                      ╲            ───
                       ╲       ────
                        ╲╲  ───
                          ●─Sadr
                         ╱ ╲╲
                       ╱╱    ╲╲
                      ╱        ╲
                     ╱          ╲╲
                   ╱╱             ╲
                  ╱                ╲╲
                 ● Gienah            ╲╲
                                       ╲
                                        ╲╲
                                          ╲
                                           ╲╲
                                             ╲╲
                                               ◦ Albireo

Cygnus (Cyg) - The Swan (Northern Cross)
Brightest star: Deneb (mag +1.25)
Stars plotted:  5
Best viewing:   September, Northern Hemisphere
```

```text
$ asterism orion --lat 35.68 --lon 139.77 --time 2026-01-15T21:00:00+09:00
  (the chart, turned to how Orion stands in that sky)

Orion (Ori) - The Hunter
Brightest star: Rigel (mag +0.18)
Stars plotted:  8
Best viewing:   January, Equatorial Hemisphere
Seen from:      35.68N 139.77E at 2026-01-15 21:00 JST
Center:         up, 54° high in the SSE (azimuth 164°)
Stars up:       8 of 8
Rise/set:       rises 15:36, highest 21:36, sets 03:36
(Chart drawn as it stands in the sky: zenith up, the horizon below.)

$ asterism --visible --lat 35.68 --lon 139.77 --time 2026-01-15T21:00:00+09:00
Up from 35.68N 139.77E at 2026-01-15 21:00 JST (35 of 88):

  Aur   Auriga                84°  NE   6/6 stars up
  Tau   Taurus                73°  SSW  8/8 stars up
  Per   Perseus               70°  NW   8/8 stars up
  ...
```

(Star glyphs and colors render best in a terminal with a Unicode font and
ANSI color support. The examples above are shown without color.)

## Flags

| Flag | Description |
| --- | --- |
| `-l`, `--list` | list all available constellations and exit |
| `-a`, `--all` | render every constellation |
| `--width N` | chart width in columns (default: `$COLUMNS` or 80) |
| `--height N` | chart height in rows (default: `$LINES` or 40) |
| `--size WxH` | set width and height together, e.g. `--size 100x50` |
| `--no-color` | disable ANSI color output |
| `--no-lines` | do not draw the figure lines |
| `--no-labels` | do not draw star name labels |
| `--ascii` | use ASCII-only glyphs (`* + .` and `- | / \`) |
| `--lat DEG` | observer latitude, north positive (needs `--lon`) |
| `--lon DEG` | observer longitude, east positive (needs `--lat`) |
| `--time T` | the moment (default: now), from 1900 to 2100: `2026-11-12T21:00` or `2026-11-12 21:00` in local time, or RFC 3339 with a zone such as `2026-11-12T21:00:00+09:00` |
| `--visible` | list the constellations whose center is above the horizon, highest first (needs `--lat`/`--lon`) |
| `-h`, `--help` | show help and exit |
| `-v`, `--version` | show version and exit |

The positional name is case-insensitive and accepts the common name, the Latin
name, or the IAU three-letter abbreviation. Spaces and hyphens are ignored, so
`ursa major`, `Ursa-Major`, and `ursamajor` all resolve to the same figure.

## Included constellations

All 88 IAU constellations (`asterism --list`). 24 of the best known are curated
by hand: Andromeda, Aquila, Aries, Auriga, Bootes, Cancer, Canis Major,
Cassiopeia, Cepheus, Corona Borealis, Crux, Cygnus, Draco, Gemini, Leo, Lyra,
Orion, Pegasus, Perseus, Sagittarius, Scorpius, Taurus, Ursa Major, Ursa Minor.

The other 64 are generated from the data of
[d3-celestial](https://github.com/ofrohn/d3-celestial) (BSD-3-Clause; see
[THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md)): its stick figures follow the
IAU constellation charts, and each figure star takes its magnitude and name
from d3-celestial's star data. Stars with a proper name are labeled with it;
other stars brighter than magnitude 4 get their Bayer designation (such as
`Alpha Ant`). Their best viewing month is the month the figure's center crosses
the meridian at 21:00, and the hemisphere follows its declination (within 10
degrees of the equator is "Equatorial"); those two rules agree with the curated
entries for 23 of the 24.

## Where it is in your sky

With `--lat` and `--lon`, asterism computes each star's altitude and azimuth
for the moment (`--time`, default now) from Greenwich mean sidereal time
(Meeus, Astronomical Algorithms, 12.4, checked against his worked examples) and
draws the figure in horizon coordinates, so it appears turned the way it stands
in the sky, zenith up. The panel gives the center's altitude and compass
direction, how many stars are above the horizon, and when the center rises,
crosses the meridian (highest), and sets, nearest to the given moment. Rising
and setting are for the center at altitude 0. The chart is the sky rotated so
that you face the figure's center with the zenith up, which also keeps figures
whole at the zenith. Precession since J2000, nutation, and refraction are
ignored, which keeps positions within about half a degree today and about 1.5
degrees at the ends of the range `--time` accepts (1900 to 2100): fine for
knowing where to look, not for pointing a telescope.

## How it works

Each star has real celestial coordinates (right ascension and declination).
asterism projects the stars of a constellation onto a tangent plane centered on
their centroid, corrects for terminal cells being about twice as tall as they
are wide, fits the figure into the grid, rasterizes the connecting lines with a
Bresenham line drawer, and plots the stars on top. All of the rendering is a
pure, deterministic function of the catalog data, so the same inputs always
produce the same picture.

## A note on accuracy

Coordinates (J2000 right ascension in hours, declination in degrees) and
apparent magnitudes are approximate catalog values chosen so the projected stick
figures are recognizable, not for astrometric precision. asterism is a
visualization and education tool, not a navigation instrument.

## Development

```bash
gofmt -l .        # should print nothing
go vet ./...
go test ./...
go build ./...
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for details, including how the
constellation catalog in `internal/sky/catalog.go` is structured.

## License

[MIT](LICENSE) (c) 2026 printemps.tokyo

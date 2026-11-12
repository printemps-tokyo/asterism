package sky

import (
	"math"
	"strings"
	"testing"
	"time"
)

// The worked examples below are from Meeus, Astronomical Algorithms, 2nd ed.,
// chapters 12 and 13.

func TestGMSTMeeusExamples(t *testing.T) {
	// Example 12.a: 1987 April 10, 0h UT -> 13h10m46.3668s.
	got := GMST(time.Date(1987, 4, 10, 0, 0, 0, 0, time.UTC))
	want := (13 + 10.0/60 + 46.3668/3600) * 15
	if math.Abs(got-want) > 1e-5 {
		t.Errorf("12.a: got %.6f°, want %.6f°", got, want)
	}
	// Example 12.b: 1987 April 10, 19h21m00s UT -> 8h34m57.0896s.
	got = GMST(time.Date(1987, 4, 10, 19, 21, 0, 0, time.UTC))
	want = (8 + 34.0/60 + 57.0896/3600) * 15
	if math.Abs(got-want) > 1e-4 {
		t.Errorf("12.b: got %.6f°, want %.6f°", got, want)
	}
}

func TestAltAzMeeusExample13b(t *testing.T) {
	// Venus from the US Naval Observatory, 1987 April 10, 19h21m00s UT:
	// alpha 23h09m16.641s, delta -6°43'11.61"; latitude +38°55'17",
	// longitude 77°03'56" W. Meeus gets h = 15.1249° and A = 68.0337°
	// measured from the south, which is 248.0337° from the north. He uses
	// apparent sidereal time; mean sidereal time differs by about 0.004°.
	o := Observer{
		Lat:  38 + 55.0/60 + 17.0/3600,
		Lon:  -(77 + 3.0/60 + 56.0/3600),
		Time: time.Date(1987, 4, 10, 19, 21, 0, 0, time.UTC),
	}
	ra := 23 + 9.0/60 + 16.641/3600
	dec := -(6 + 43.0/60 + 11.61/3600)
	alt, az := AltAz(ra, dec, o)
	if math.Abs(alt-15.1249) > 0.01 || math.Abs(az-248.0337) > 0.01 {
		t.Errorf("got alt %.4f az %.4f, want 15.1249 248.0337", alt, az)
	}
}

func TestPolarisStaysNearThePole(t *testing.T) {
	polaris := Star{RA: 2.530, Dec: 89.264}
	for h := 0; h < 24; h += 3 {
		o := Observer{Lat: 35.68, Lon: 139.77, Time: time.Date(2026, 11, 12, h, 0, 0, 0, time.UTC)}
		alt, az := AltAz(polaris.RA, polaris.Dec, o)
		if math.Abs(alt-35.68) > 1 {
			t.Errorf("%02d:00 UTC: Polaris altitude %.2f, want about the latitude", h, alt)
		}
		if az > 2 && az < 358 {
			t.Errorf("%02d:00 UTC: Polaris azimuth %.2f, want about north", h, az)
		}
	}
}

func TestSeeRiseTransitSet(t *testing.T) {
	orion, _ := Lookup("orion")
	tokyo := time.FixedZone("JST", 9*3600)
	o := Observer{Lat: 35.68, Lon: 139.77, Time: time.Date(2026, 1, 15, 21, 0, 0, 0, tokyo)}
	v := See(orion, o)
	if v.Circumpolar || v.NeverRises {
		t.Fatalf("Orion rises and sets in Tokyo: %+v", v)
	}
	if v.Alt < 30 || v.Az < 135 || v.Az > 225 {
		t.Errorf("mid-January 21:00 in Tokyo, Orion should be high in the south: alt %.1f az %.1f", v.Alt, v.Az)
	}
	if !(v.Rise.Before(v.Transit) && v.Transit.Before(v.Set)) {
		t.Errorf("rise %v, transit %v, set %v out of order", v.Rise, v.Transit, v.Set)
	}
	if d := v.Transit.Sub(o.Time); d < -12*time.Hour || d > 12*time.Hour {
		t.Errorf("transit %v not within 12 hours of %v", v.Transit, o.Time)
	}
	// At transit the center is due south (azimuth 180) for a southern-ish
	// declination seen from the north.
	at := o
	at.Time = v.Transit
	if v2 := See(orion, at); math.Abs(v2.Az-180) > 1 {
		t.Errorf("azimuth at transit %.2f, want 180", v2.Az)
	}
	// At the computed rise the center is on the horizon.
	at.Time = v.Rise
	if v3 := See(orion, at); math.Abs(v3.Alt) > 0.5 {
		t.Errorf("altitude at rise %.2f, want 0", v3.Alt)
	}

	crux, _ := Lookup("crux")
	if v := See(crux, Observer{Lat: 60, Lon: 10, Time: o.Time}); !v.NeverRises {
		t.Errorf("Crux never rises at 60N: %+v", v)
	}
	uma, _ := Lookup("uma")
	if v := See(uma, Observer{Lat: 60, Lon: 10, Time: o.Time}); !v.Circumpolar {
		t.Errorf("Ursa Major is circumpolar at 60N: %+v", v)
	}
}

func TestHorizonViewAndPanel(t *testing.T) {
	orion, _ := Lookup("orion")
	o := Observer{Lat: 35.68, Lon: 139.77, Time: time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)}
	view := HorizonView(orion, o)
	if len(view.Stars) != len(orion.Stars) || view.Stars[0].Name != orion.Stars[0].Name {
		t.Fatal("HorizonView must keep the stars and names")
	}
	if out := Render(view, DefaultOptions(60, 20)).String(false); !strings.Contains(out, "Betelgeuse") {
		t.Errorf("horizon view render lost a label:\n%s", out)
	}
	lines := ObserverLines(orion, o)
	if len(lines) != 4 || !strings.Contains(lines[0], "35.68N 139.77E") || !strings.Contains(lines[3], "rises") {
		t.Errorf("panel: %q", lines)
	}
	for az, want := range map[float64]string{0: "N", 11: "N", 12: "NNE", 90: "E", 180: "S", 248: "WSW", 359: "N"} {
		if got := Compass(az); got != want {
			t.Errorf("Compass(%v) = %s, want %s", az, got, want)
		}
	}
	if got := FormatLatLon(-33.87, -70.65); got != "33.87S 70.65W" {
		t.Errorf("FormatLatLon: %s", got)
	}
}

func TestHorizonViewOrientationAndZenith(t *testing.T) {
	orion, _ := Lookup("orion")
	idx := map[string]int{}
	for i, s := range orion.Stars {
		idx[s.Name] = i
	}
	// From Tokyo at Orion's transit, facing south: Betelgeuse (east and north
	// of Rigel) is up and to the left.
	tokyo := Observer{Lat: 35.68, Lon: 139.77, Time: time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)}
	tokyo.Time = See(orion, tokyo).Transit
	pts := ProjectStars(HorizonView(orion, tokyo).Stars, 80, 40, 3, 2)
	b, r := pts[idx["Betelgeuse"]], pts[idx["Rigel"]]
	if !(b.X < r.X && b.Y < r.Y) {
		t.Errorf("facing south, Betelgeuse %+v should be up-left of Rigel %+v", b, r)
	}

	// At the equator with Orion near the zenith the figure must stay whole:
	// the belt stars stay close together.
	eq := Observer{Lat: 0, Lon: 0, Time: time.Date(2026, 1, 15, 22, 0, 0, 0, time.UTC)}
	if v := See(orion, eq); v.Alt < 80 {
		t.Fatalf("setup: Orion should be near the zenith, alt %.1f", v.Alt)
	}
	pts = ProjectStars(HorizonView(orion, eq).Stars, 80, 40, 3, 2)
	belt := []Point{pts[idx["Mintaka"]], pts[idx["Alnilam"]], pts[idx["Alnitak"]]}
	for i := 1; i < len(belt); i++ {
		if d := math.Hypot(belt[i].X-belt[i-1].X, belt[i].Y-belt[i-1].Y); d > 20 {
			t.Errorf("belt stars %d apart at the zenith: %+v", int(d), belt)
		}
	}
}

func TestJulianDayOutsideUnixNanoRange(t *testing.T) {
	// 1600-01-01 00:00 UTC is JD 2305447.5 (Gregorian), outside UnixNano.
	if got := JulianDay(time.Date(1600, 1, 1, 0, 0, 0, 0, time.UTC)); math.Abs(got-2305447.5) > 1e-6 {
		t.Errorf("JD: %f", got)
	}
}

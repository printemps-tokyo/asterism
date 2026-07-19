package sky

import (
	"strings"
	"testing"
)

// requiredConstellations must all be present in the catalog.
var requiredConstellations = []string{
	"Orion", "Ursa Major", "Ursa Minor", "Cassiopeia", "Cygnus", "Lyra",
	"Aquila", "Scorpius", "Sagittarius", "Leo", "Taurus", "Gemini",
	"Canis Major", "Bootes", "Crux", "Andromeda", "Pegasus", "Perseus",
	"Auriga", "Corona Borealis", "Cepheus", "Draco", "Cancer", "Aries",
}

func TestCatalogHasRequiredConstellations(t *testing.T) {
	if len(catalog) < 24 {
		t.Fatalf("catalog has %d constellations, want at least 24", len(catalog))
	}
	for _, name := range requiredConstellations {
		if _, ok := Lookup(name); !ok {
			t.Errorf("required constellation %q not found in catalog", name)
		}
	}
}

func TestCatalogIntegrity(t *testing.T) {
	for _, c := range catalog {
		if len(c.Stars) == 0 {
			t.Errorf("%s: has no stars", c.Latin)
		}
		if c.Latin == "" || c.Abbr == "" || c.Common == "" {
			t.Errorf("%s (%s): missing a name field", c.Latin, c.Abbr)
		}
		if c.Hemisphere == "" || c.BestMonth == "" {
			t.Errorf("%s: missing hemisphere or best month", c.Latin)
		}
		for i, s := range c.Stars {
			if s.RA < 0 || s.RA >= 24 {
				t.Errorf("%s star %d (%s): RA %g out of [0,24)", c.Latin, i, s.Name, s.RA)
			}
			if s.Dec < -90 || s.Dec > 90 {
				t.Errorf("%s star %d (%s): Dec %g out of [-90,90]", c.Latin, i, s.Name, s.Dec)
			}
			if s.Mag < -2 || s.Mag > 8 {
				t.Errorf("%s star %d (%s): Mag %g out of [-2,8]", c.Latin, i, s.Name, s.Mag)
			}
		}
		for _, seg := range c.Lines {
			for _, idx := range seg {
				if idx < 0 || idx >= len(c.Stars) {
					t.Errorf("%s: line segment references star index %d, only %d stars", c.Latin, idx, len(c.Stars))
				}
			}
			if seg[0] == seg[1] {
				t.Errorf("%s: line segment connects star %d to itself", c.Latin, seg[0])
			}
		}
	}
}

func TestAbbreviationsUnique(t *testing.T) {
	seen := make(map[string]string)
	for _, c := range catalog {
		if prev, ok := seen[c.Abbr]; ok {
			t.Errorf("duplicate abbreviation %q used by %q and %q", c.Abbr, prev, c.Latin)
		}
		seen[c.Abbr] = c.Latin
	}
}

func TestLookupCaseInsensitiveAndAbbrev(t *testing.T) {
	cases := []struct {
		query string
		want  string
	}{
		{"orion", "Orion"},
		{"ORION", "Orion"},
		{"Ori", "Orion"},
		{"ursa major", "Ursa Major"},
		{"URSA-MAJOR", "Ursa Major"},
		{"ursamajor", "Ursa Major"},
		{"uma", "Ursa Major"},
		{"cru", "Crux"},
		{"The Southern Cross", "Crux"},
	}
	for _, tc := range cases {
		c, ok := Lookup(tc.query)
		if !ok {
			t.Errorf("Lookup(%q) not found", tc.query)
			continue
		}
		if c.Latin != tc.want {
			t.Errorf("Lookup(%q) = %q, want %q", tc.query, c.Latin, tc.want)
		}
	}
}

func TestLookupMiss(t *testing.T) {
	if _, ok := Lookup("nonesuch"); ok {
		t.Error("Lookup(nonesuch) unexpectedly found a match")
	}
}

func TestCloseMatches(t *testing.T) {
	m := CloseMatches("ursa")
	if len(m) < 2 {
		t.Errorf("CloseMatches(ursa) = %v, want both Ursa constellations", m)
	}
	joined := strings.Join(m, ",")
	if !strings.Contains(joined, "Ursa Major") || !strings.Contains(joined, "Ursa Minor") {
		t.Errorf("CloseMatches(ursa) = %v, want to include both Ursa Major and Ursa Minor", m)
	}
}

func TestBrightestStar(t *testing.T) {
	ori, _ := Lookup("orion")
	b := ori.BrightestStar()
	// Rigel (0.18) is marginally brighter than Betelgeuse (0.50) in this catalog.
	if b.Name != "Rigel" {
		t.Errorf("Orion brightest = %q (mag %g), want Rigel", b.Name, b.Mag)
	}
}

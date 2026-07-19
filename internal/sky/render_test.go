package sky

import (
	"math"
	"strings"
	"testing"
	"unicode"
)

func TestRenderDeterministic(t *testing.T) {
	ori, _ := Lookup("orion")
	opts := DefaultOptions(80, 30)
	a := Render(ori, opts).String(true)
	b := Render(ori, opts).String(true)
	if a != b {
		t.Error("Render is not deterministic: two identical calls produced different output")
	}
}

func TestRenderNoColorHasNoANSI(t *testing.T) {
	ori, _ := Lookup("orion")
	opts := DefaultOptions(80, 30)
	opts.Color = false
	out := Render(ori, opts).String(false)
	if strings.Contains(out, "\x1b[") {
		t.Error("no-color output contains an ANSI escape sequence")
	}
}

func TestRenderColorHasANSI(t *testing.T) {
	ori, _ := Lookup("orion")
	opts := DefaultOptions(80, 30)
	out := Render(ori, opts).String(true)
	if !strings.Contains(out, "\x1b[") {
		t.Error("color output has no ANSI escape sequence")
	}
}

func TestRenderASCIIOnly(t *testing.T) {
	opts := DefaultOptions(80, 30)
	opts.ASCII = true
	opts.Color = false
	for _, c := range catalog {
		out := Render(c, opts).String(false)
		for _, r := range out {
			if r > unicode.MaxASCII {
				t.Errorf("%s: ASCII mode emitted non-ASCII rune %q", c.Latin, r)
				break
			}
		}
	}
}

func TestRenderNoLines(t *testing.T) {
	ori, _ := Lookup("orion")
	opts := DefaultOptions(60, 24)
	opts.Color = false
	opts.Labels = false
	opts.ASCII = true

	withLines := Render(ori, opts).String(false)
	opts.Lines = false
	withoutLines := Render(ori, opts).String(false)

	countGlyph := func(s string, glyphs string) int {
		n := 0
		for _, r := range s {
			if strings.ContainsRune(glyphs, r) {
				n++
			}
		}
		return n
	}
	if countGlyph(withoutLines, `-|/\`) != 0 {
		t.Error("--no-lines still drew line glyphs")
	}
	if countGlyph(withLines, `-|/\`) == 0 {
		t.Error("expected line glyphs when lines are enabled")
	}
}

func TestRenderLabelsDoNotOverwriteStars(t *testing.T) {
	ori, _ := Lookup("orion")
	opts := DefaultOptions(100, 40)
	opts.Color = false
	buf := Render(ori, opts)

	// Every projected star cell must still hold a star glyph, not a label char.
	pts := ProjectStars(ori.Stars, 100, 40, 3, 2)
	starGlyphs := "★✦●◦·"
	for i, p := range pts {
		col := int(math.Round(p.X))
		row := int(math.Round(p.Y))
		ch := buf.at(col, row)
		if !strings.ContainsRune(starGlyphs, ch) {
			t.Errorf("Orion star %d (%s) cell holds %q, not a star glyph", i, ori.Stars[i].Name, ch)
		}
	}
}

// --- known-shape sanity checks ---

func TestOrionBeltIsCollinear(t *testing.T) {
	ori, _ := Lookup("orion")
	pts := ProjectStars(ori.Stars, 80, 40, 3, 2)
	// Belt stars are catalog indices 3 (Mintaka), 4 (Alnilam), 5 (Alnitak).
	a, mid, c := pts[3], pts[4], pts[5]

	// Near-collinear: the cross product of (mid-a) and (c-a) is small relative
	// to the segment length.
	cross := (mid.X-a.X)*(c.Y-a.Y) - (mid.Y-a.Y)*(c.X-a.X)
	length := math.Hypot(c.X-a.X, c.Y-a.Y)
	perpDist := math.Abs(cross) / length
	if perpDist > 1.5 {
		t.Errorf("Orion belt not collinear: middle star is %g cells off the line", perpDist)
	}

	// The middle star should lie between the outer two.
	dot := (mid.X-a.X)*(c.X-a.X) + (mid.Y-a.Y)*(c.Y-a.Y)
	if dot < 0 || dot > length*length {
		t.Error("Orion belt middle star is not between the outer belt stars")
	}

	// Belt stars are of similar brightness.
	if math.Abs(ori.Stars[3].Mag-ori.Stars[5].Mag) > 1.0 {
		t.Error("Orion belt stars are not of similar magnitude")
	}
}

func TestBigDipperShape(t *testing.T) {
	uma, _ := Lookup("ursa major")
	if len(uma.Stars) != 7 {
		t.Errorf("Big Dipper has %d stars, want 7", len(uma.Stars))
	}
	if len(uma.Lines) != 7 {
		t.Errorf("Big Dipper has %d connecting segments, want 7 (4 bowl + 3 handle)", len(uma.Lines))
	}
}

func TestCruxHasFourMainStars(t *testing.T) {
	cru, _ := Lookup("crux")
	if len(cru.Stars) != 4 {
		t.Errorf("Crux has %d stars, want 4", len(cru.Stars))
	}
}

// --- rasterizer checks ---

func TestDrawLineHorizontal(t *testing.T) {
	b := newBuffer(20, 5)
	noStars := map[[2]int]bool{}
	drawLine(b, 2, 2, 10, 2, '─', "", noStars)
	for c := 2; c <= 10; c++ {
		if b.at(c, 2) != '─' {
			t.Errorf("horizontal line: cell (%d,2) = %q, want ─", c, b.at(c, 2))
		}
	}
	// endpoints must be hit
	if b.at(2, 2) != '─' || b.at(10, 2) != '─' {
		t.Error("horizontal line endpoints not drawn")
	}
}

func TestDrawLineVertical(t *testing.T) {
	b := newBuffer(6, 12)
	noStars := map[[2]int]bool{}
	drawLine(b, 3, 1, 3, 8, '│', "", noStars)
	for r := 1; r <= 8; r++ {
		if b.at(3, r) != '│' {
			t.Errorf("vertical line: cell (3,%d) = %q, want │", r, b.at(3, r))
		}
	}
}

func TestDrawLineDiagonal(t *testing.T) {
	b := newBuffer(12, 12)
	noStars := map[[2]int]bool{}
	drawLine(b, 1, 1, 8, 8, '╲', "", noStars)
	// The diagonal should produce a connected run: each row from 1..8 has
	// exactly one glyph and columns advance by one each step.
	for d := 0; d <= 7; d++ {
		if b.at(1+d, 1+d) != '╲' {
			t.Errorf("diagonal line: cell (%d,%d) = %q, want ╲", 1+d, 1+d, b.at(1+d, 1+d))
		}
	}
}

func TestDrawLineSkipsStarCells(t *testing.T) {
	b := newBuffer(20, 5)
	stars := map[[2]int]bool{{5, 2}: true}
	drawLine(b, 2, 2, 10, 2, '─', "", stars)
	if b.at(5, 2) != ' ' {
		t.Errorf("drawLine overwrote a star cell at (5,2): got %q", b.at(5, 2))
	}
}

func TestLineGlyphSelection(t *testing.T) {
	cases := []struct {
		dc, dr int
		want   rune
	}{
		{10, 0, '─'},
		{0, 10, '│'},
		{10, 10, '╲'},
		{10, -10, '╱'},
		{-10, -10, '╲'},
	}
	for _, tc := range cases {
		if got := lineGlyph(tc.dc, tc.dr, false); got != tc.want {
			t.Errorf("lineGlyph(%d,%d) = %q, want %q", tc.dc, tc.dr, got, tc.want)
		}
	}
	// ASCII fallbacks.
	if lineGlyph(10, 0, true) != '-' || lineGlyph(0, 10, true) != '|' {
		t.Error("ASCII line glyphs incorrect")
	}
}

func TestInfoLines(t *testing.T) {
	ori, _ := Lookup("orion")
	lines := InfoLines(ori)
	joined := strings.Join(lines, "\n")
	for _, want := range []string{"Orion", "Ori", "Rigel", "January", "Equatorial"} {
		if !strings.Contains(joined, want) {
			t.Errorf("info panel missing %q\n%s", want, joined)
		}
	}
}

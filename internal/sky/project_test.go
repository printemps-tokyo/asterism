package sky

import (
	"math"
	"testing"
)

func TestProjectStarsWithinBounds(t *testing.T) {
	const w, h, mx, my = 80, 40, 3, 2
	for _, c := range catalog {
		pts := ProjectStars(c.Stars, w, h, mx, my)
		for i, p := range pts {
			if p.X < 0 || p.X > float64(w) {
				t.Errorf("%s star %d: X=%g outside [0,%d]", c.Latin, i, p.X, w)
			}
			if p.Y < 0 || p.Y > float64(h) {
				t.Errorf("%s star %d: Y=%g outside [0,%d]", c.Latin, i, p.Y, h)
			}
			// Points should respect the margins (with a small rounding slack).
			if p.X < float64(mx)-0.5 || p.X > float64(w-mx)+0.5 {
				t.Errorf("%s star %d: X=%g violates x-margin %d", c.Latin, i, p.X, mx)
			}
			if p.Y < float64(my)-0.5 || p.Y > float64(h-my)+0.5 {
				t.Errorf("%s star %d: Y=%g violates y-margin %d", c.Latin, i, p.Y, my)
			}
		}
	}
}

func TestProjectStarsCentroidCentered(t *testing.T) {
	const w, h, mx, my = 80, 40, 3, 2
	ori, _ := Lookup("orion")
	pts := ProjectStars(ori.Stars, w, h, mx, my)
	var sx, sy float64
	for _, p := range pts {
		sx += p.X
		sy += p.Y
	}
	cx := sx / float64(len(pts))
	cy := sy / float64(len(pts))
	// The projected centroid should sit roughly in the middle of the grid.
	if math.Abs(cx-float64(w)/2) > float64(w)/4 {
		t.Errorf("centroid X=%g not near grid center %d", cx, w/2)
	}
	if math.Abs(cy-float64(h)/2) > float64(h)/4 {
		t.Errorf("centroid Y=%g not near grid center %d", cy, h/2)
	}
}

func TestProjectStarsAspectCorrection(t *testing.T) {
	// Three stars: an origin, one shifted east by ~15 degrees, one shifted
	// north by 15 degrees. Equal angular separations should map to a projected
	// x-distance roughly twice the y-distance, thanks to the cell-aspect
	// correction (terminal cells are ~2x taller than wide).
	stars := []Star{
		{"O", 12.0, 0.0, 1.0},
		{"E", 11.0, 0.0, 1.0}, // 1 hour = 15 degrees of RA
		{"N", 12.0, 15.0, 1.0},
	}
	pts := ProjectStars(stars, 200, 200, 0, 0)
	dxEast := math.Hypot(pts[1].X-pts[0].X, pts[1].Y-pts[0].Y)
	dyNorth := math.Hypot(pts[2].X-pts[0].X, pts[2].Y-pts[0].Y)
	ratio := dxEast / dyNorth
	if ratio < 1.7 || ratio > 2.3 {
		t.Errorf("aspect ratio east/north = %g, want ~2.0", ratio)
	}
}

func TestProjectStarsEastIsLeft(t *testing.T) {
	// A star at larger RA (further east) should project to a smaller column
	// than one at smaller RA, because the sky is drawn as seen looking up.
	stars := []Star{
		{"west", 10.0, 20.0, 1.0},
		{"east", 11.0, 20.0, 1.0},
	}
	pts := ProjectStars(stars, 80, 40, 3, 2)
	if pts[1].X >= pts[0].X {
		t.Errorf("east star X=%g should be left of west star X=%g", pts[1].X, pts[0].X)
	}
}

func TestUnwrapRA(t *testing.T) {
	// Pegasus straddles 0h; unwrapping around a ~23h reference should pull the
	// 0.2h star to a value near 24h so the figure does not tear.
	got := unwrapRA(0.221, 23.0)
	if math.Abs(got-24.221) > 1e-9 {
		t.Errorf("unwrapRA(0.221, 23) = %g, want ~24.221", got)
	}
	got = unwrapRA(23.079, 0.14)
	if math.Abs(got-(-0.921)) > 1e-9 {
		t.Errorf("unwrapRA(23.079, 0.14) = %g, want ~-0.921", got)
	}
}

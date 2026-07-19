package sky

import (
	"math"
	"strings"
)

// Options controls how a constellation is rendered into a character buffer.
// The zero value is not useful; callers should set Width and Height and pick
// the boolean features they want. Rendering is a pure function of a
// Constellation plus these Options.
type Options struct {
	Width  int  // grid width in character cells
	Height int  // grid height in character cells
	Color  bool // emit ANSI color escapes when the buffer is stringified
	Lines  bool // draw the figure line segments
	Labels bool // draw star name labels
	ASCII  bool // use ASCII-only glyphs instead of Unicode
}

// DefaultOptions returns Options with sensible defaults for the given size.
func DefaultOptions(width, height int) Options {
	return Options{
		Width:  width,
		Height: height,
		Color:  true,
		Lines:  true,
		Labels: true,
		ASCII:  false,
	}
}

// Cell is one character in the grid: a rune plus an optional ANSI SGR
// parameter string (for example "97" or "1;97"). An empty Color means the cell
// is rendered without any escape sequence.
type Cell struct {
	Ch    rune
	Color string
}

// Buffer is a 2D grid of Cells, addressed as Cells[row][col].
type Buffer struct {
	W, H  int
	Cells [][]Cell
}

// Point is a projected star position in fractional grid coordinates
// (X = column, Y = row), before rounding to a cell.
type Point struct {
	X, Y float64
}

// cellAspect corrects for terminal cells being roughly twice as tall as they
// are wide: horizontal distances are scaled up by this factor before fitting so
// figures are not vertically squashed.
const cellAspect = 2.0

// newBuffer allocates a blank Buffer filled with spaces.
func newBuffer(w, h int) *Buffer {
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	cells := make([][]Cell, h)
	for r := range cells {
		row := make([]Cell, w)
		for c := range row {
			row[c] = Cell{Ch: ' '}
		}
		cells[r] = row
	}
	return &Buffer{W: w, H: h, Cells: cells}
}

// set writes a cell if (col,row) is inside the grid.
func (b *Buffer) set(col, row int, ch rune, color string) {
	if col < 0 || col >= b.W || row < 0 || row >= b.H {
		return
	}
	b.Cells[row][col] = Cell{Ch: ch, Color: color}
}

// at reports the rune currently at (col,row), or ' ' when out of bounds.
func (b *Buffer) at(col, row int) rune {
	if col < 0 || col >= b.W || row < 0 || row >= b.H {
		return ' '
	}
	return b.Cells[row][col].Ch
}

// String renders the buffer to a multi-line string. When color is true, cells
// with a non-empty Color are wrapped in the matching ANSI SGR escape; when
// false, no escape sequences are emitted at all. Trailing spaces on each line
// are trimmed.
func (b *Buffer) String(color bool) string {
	var sb strings.Builder
	for r := 0; r < b.H; r++ {
		last := -1
		for c := 0; c < b.W; c++ {
			if b.Cells[r][c].Ch != ' ' {
				last = c
			}
		}
		for c := 0; c <= last; c++ {
			cell := b.Cells[r][c]
			if color && cell.Color != "" && cell.Ch != ' ' {
				sb.WriteString("\x1b[")
				sb.WriteString(cell.Color)
				sb.WriteByte('m')
				sb.WriteRune(cell.Ch)
				sb.WriteString("\x1b[0m")
			} else {
				sb.WriteRune(cell.Ch)
			}
		}
		if r < b.H-1 {
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}

// unwrapRA shifts an RA (hours) to be within 12 hours of ref, so a
// constellation that straddles the 0h/24h boundary projects without a seam.
func unwrapRA(ra, ref float64) float64 {
	for ra-ref > 12 {
		ra -= 24
	}
	for ra-ref < -12 {
		ra += 24
	}
	return ra
}

// ProjectStars projects a set of stars onto the character grid using a
// locally-linear tangent-plane projection centered on the star centroid, then
// scales the result to fit within the grid minus the given margins while
// preserving aspect ratio (with the terminal cell-aspect correction applied).
// North is up: smaller row numbers are higher declinations. East is left, to
// match the sky as seen looking up.
func ProjectStars(stars []Star, width, height, marginX, marginY int) []Point {
	pts := make([]Point, len(stars))
	if len(stars) == 0 {
		return pts
	}

	// Unwrap RA around the first star and compute the centroid.
	ref := stars[0].RA
	ras := make([]float64, len(stars))
	var sumRA, sumDec float64
	for i, s := range stars {
		ras[i] = unwrapRA(s.RA, ref)
		sumRA += ras[i]
		sumDec += s.Dec
	}
	ra0 := sumRA / float64(len(stars))
	dec0 := sumDec / float64(len(stars))
	cosDec0 := math.Cos(dec0 * math.Pi / 180)

	// Tangent-plane coordinates. RA is converted from hours to degrees (x15).
	// x is negated so east falls to the left; then scaled by the cell aspect.
	raw := make([]Point, len(stars))
	for i := range stars {
		x := -(ras[i] - ra0) * 15 * cosDec0 * cellAspect
		y := stars[i].Dec - dec0
		raw[i] = Point{X: x, Y: y}
	}

	minX, maxX := raw[0].X, raw[0].X
	minY, maxY := raw[0].Y, raw[0].Y
	for _, p := range raw[1:] {
		minX = math.Min(minX, p.X)
		maxX = math.Max(maxX, p.X)
		minY = math.Min(minY, p.Y)
		maxY = math.Max(maxY, p.Y)
	}
	spanX := maxX - minX
	spanY := maxY - minY
	if spanX == 0 {
		spanX = 1
	}
	if spanY == 0 {
		spanY = 1
	}

	availW := float64(width - 2*marginX)
	availH := float64(height - 2*marginY)
	if availW < 1 {
		availW = 1
	}
	if availH < 1 {
		availH = 1
	}
	scale := math.Min(availW/spanX, availH/spanY)

	offX := float64(marginX) + (availW-spanX*scale)/2
	offY := float64(marginY) + (availH-spanY*scale)/2

	for i, p := range raw {
		gx := offX + (p.X-minX)*scale
		// Flip Y so higher declination is nearer the top of the grid.
		gy := offY + (maxY-p.Y)*scale
		pts[i] = Point{X: gx, Y: gy}
	}
	return pts
}

// starGlyph returns the plotting glyph for a star of the given magnitude.
func starGlyph(mag float64, ascii bool) rune {
	if ascii {
		switch {
		case mag < 1.5:
			return '*'
		case mag < 3.0:
			return '+'
		default:
			return '.'
		}
	}
	switch {
	case mag < 1.0:
		return '★' // solid star
	case mag < 2.0:
		return '✦' // black four pointed star
	case mag < 3.0:
		return '●' // large filled circle
	case mag < 4.0:
		return '◦' // small hollow circle
	default:
		return '·' // middle dot
	}
}

// starColor returns the ANSI SGR parameters for a star of the given magnitude,
// running from bold blue-white for the brightest to gray for the faintest.
func starColor(mag float64) string {
	switch {
	case mag < 1.0:
		return "1;97" // bold bright white
	case mag < 2.0:
		return "97" // bright white
	case mag < 3.0:
		return "96" // bright cyan (blue-white feel)
	case mag < 4.0:
		return "37" // white
	default:
		return "90" // gray
	}
}

// lineColor is the dim color used for figure line segments.
const lineColor = "90"

// labelColor is the color used for star name labels.
const labelColor = "37"

// lineGlyph picks a line-drawing glyph for a segment with the given screen-space
// deltas (dc = columns, dr = rows). Row numbers increase downward.
func lineGlyph(dc, dr int, ascii bool) rune {
	adc, adr := abs(dc), abs(dr)
	switch {
	case adr*2 <= adc:
		if ascii {
			return '-'
		}
		return '─' // horizontal
	case adc*2 <= adr:
		if ascii {
			return '|'
		}
		return '│' // vertical
	default:
		// Same sign on both deltas means down-right / up-left: a backslash.
		if (dc > 0) == (dr > 0) {
			if ascii {
				return '\\'
			}
			return '╲' // box drawings light diagonal upper-left to lower-right
		}
		if ascii {
			return '/'
		}
		return '╱' // box drawings light diagonal upper-right to lower-left
	}
}

// drawLine rasterizes a straight segment between two cells using Bresenham's
// algorithm, writing glyph into each cell along the path. Existing star cells
// are left untouched so lines never paint over a plotted star.
func drawLine(b *Buffer, c0, r0, c1, r1 int, glyph rune, color string, stars map[[2]int]bool) {
	dc := abs(c1 - c0)
	dr := -abs(r1 - r0)
	sc := 1
	if c0 > c1 {
		sc = -1
	}
	sr := 1
	if r0 > r1 {
		sr = -1
	}
	err := dc + dr
	c, r := c0, r0
	for {
		if !stars[[2]int{c, r}] {
			b.set(c, r, glyph, color)
		}
		if c == c1 && r == r1 {
			break
		}
		e2 := 2 * err
		if e2 >= dr {
			err += dr
			c += sc
		}
		if e2 <= dc {
			err += dc
			r += sr
		}
	}
}

// Render projects a constellation and rasterizes its figure lines, star glyphs,
// and (optionally) labels into a fresh Buffer. It is a pure, deterministic
// function of c and opts: identical inputs always yield an identical Buffer.
func Render(c Constellation, opts Options) *Buffer {
	w, h := opts.Width, opts.Height
	if w < 10 {
		w = 10
	}
	if h < 6 {
		h = 6
	}
	b := newBuffer(w, h)

	const marginX, marginY = 3, 2
	pts := ProjectStars(c.Stars, w, h, marginX, marginY)

	// Rounded star cell positions, and a set for collision checks.
	type cell struct{ c, r int }
	starCells := make([]cell, len(pts))
	starSet := make(map[[2]int]bool, len(pts))
	for i, p := range pts {
		cc := int(math.Round(p.X))
		rr := int(math.Round(p.Y))
		starCells[i] = cell{cc, rr}
		starSet[[2]int{cc, rr}] = true
	}

	// 1. Figure lines, underneath the stars.
	if opts.Lines {
		for _, seg := range c.Lines {
			a, z := starCells[seg[0]], starCells[seg[1]]
			glyph := lineGlyph(z.c-a.c, z.r-a.r, opts.ASCII)
			color := ""
			if opts.Color {
				color = lineColor
			}
			drawLine(b, a.c, a.r, z.c, z.r, glyph, color, starSet)
		}
	}

	// 2. Star glyphs, on top of the lines.
	for i, sc := range starCells {
		glyph := starGlyph(c.Stars[i].Mag, opts.ASCII)
		color := ""
		if opts.Color {
			color = starColor(c.Stars[i].Mag)
		}
		b.set(sc.c, sc.r, glyph, color)
	}

	// 3. Labels for named stars, placed adjacent without overwriting.
	if opts.Labels {
		occupied := make(map[[2]int]bool)
		for _, sc := range starCells {
			occupied[[2]int{sc.c, sc.r}] = true
		}
		for i, sc := range starCells {
			name := c.Stars[i].Name
			if name == "" {
				continue
			}
			placeLabel(b, sc.c, sc.r, name, opts.Color, occupied)
		}
	}

	return b
}

// placeLabel tries to write name next to a star at (col,row) without colliding
// with any occupied cell or another label. It prefers the right side, then the
// left, and skips the label entirely if neither fits on the grid.
func placeLabel(b *Buffer, col, row int, name string, color bool, occupied map[[2]int]bool) {
	runes := []rune(name)
	clr := ""
	if color {
		clr = labelColor
	}

	// Candidate start columns: two cells to the right, then to the left.
	tryRight := col + 2
	tryLeft := col - 1 - len(runes)

	for _, start := range []int{tryRight, tryLeft} {
		if start < 0 || start+len(runes) > b.W {
			continue
		}
		if row < 0 || row >= b.H {
			continue
		}
		fits := true
		for i := range runes {
			key := [2]int{start + i, row}
			if occupied[key] || b.at(start+i, row) != ' ' {
				fits = false
				break
			}
		}
		if !fits {
			continue
		}
		for i, ch := range runes {
			b.set(start+i, row, ch, clr)
			occupied[[2]int{start + i, row}] = true
		}
		return
	}
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

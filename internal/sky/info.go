package sky

import (
	"fmt"
	"strconv"
)

// InfoLines returns the text of the info panel for a constellation: its Latin
// and common names, IAU abbreviation, brightest star and magnitude, plotted
// star count, and best-viewing hemisphere and month. The lines are plain text
// with no ANSI escapes so callers can color or box them as they wish.
func InfoLines(c Constellation) []string {
	brightest := c.BrightestStar()
	name := brightest.Name
	if name == "" {
		name = "(unnamed)"
	}
	return []string{
		fmt.Sprintf("%s (%s) - %s", c.Latin, c.Abbr, c.Common),
		fmt.Sprintf("Brightest star: %s (mag %s)", name, formatMag(brightest.Mag)),
		fmt.Sprintf("Stars plotted:  %d", len(c.Stars)),
		fmt.Sprintf("Best viewing:   %s, %s Hemisphere", c.BestMonth, c.Hemisphere),
	}
}

// formatMag renders a magnitude with a leading sign and two decimals, so the
// panel lines up and negative magnitudes (very bright stars) read clearly.
func formatMag(mag float64) string {
	s := strconv.FormatFloat(mag, 'f', 2, 64)
	if mag >= 0 {
		return "+" + s
	}
	return s
}

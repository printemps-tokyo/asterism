// Command asterism renders constellations and their stick-figure lines in the
// terminal. Give it a constellation name (or IAU abbreviation) and it plots the
// bright stars, connects them with the figure lines, labels the named stars,
// and prints a short info panel.
//
// asterism has no dependencies beyond the Go standard library. Terminal color
// is emitted as ANSI escape codes; terminal size is read from the COLUMNS and
// LINES environment variables (with an 80x40 fallback) and can be overridden
// with --width/--height or --size.
package main

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/printemps-tokyo/asterism/internal/sky"
)

// version is overridden at release build time via -ldflags "-X main.version=...".
var version = "dev"

const (
	defaultWidth  = 80
	defaultHeight = 40
	// panelReserve is the number of rows left below the chart for the blank
	// separator line and the info panel.
	panelReserve = 6
)

// config holds the parsed command-line state.
type config struct {
	name    string
	list    bool
	all     bool
	width   int
	height  int
	color   bool
	lines   bool
	labels  bool
	ascii   bool
	help    bool
	version bool

	// Observer: set together by --lat and --lon; --time defaults to now.
	observe bool
	lat     float64
	lon     float64
	at      time.Time
	visible bool
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr *os.File) int {
	cfg, err := parseArgs(args)
	if err != nil {
		fmt.Fprintln(stderr, "asterism:", err)
		fmt.Fprintln(stderr, "Try 'asterism --help'.")
		return 2
	}

	switch {
	case cfg.help:
		fmt.Fprint(stdout, usage())
		return 0
	case cfg.version:
		fmt.Fprintf(stdout, "asterism %s\n", version)
		return 0
	case cfg.list:
		printList(stdout)
		return 0
	case cfg.visible:
		printVisible(stdout, cfg.observer())
		return 0
	}

	opts := sky.Options{
		Width:  cfg.width,
		Height: cfg.height,
		Color:  cfg.color,
		Lines:  cfg.lines,
		Labels: cfg.labels,
		ASCII:  cfg.ascii,
	}

	if cfg.all {
		catalog := sky.Catalog()
		for i, c := range catalog {
			if i > 0 {
				fmt.Fprint(stdout, "\n"+strings.Repeat("─", 40)+"\n\n")
			}
			renderOne(stdout, c, opts, cfg)
		}
		return 0
	}

	if cfg.name == "" {
		fmt.Fprintln(stderr, "asterism: no constellation given")
		fmt.Fprintln(stderr, "Try 'asterism orion', 'asterism --list', or 'asterism --help'.")
		return 2
	}

	c, ok := sky.Lookup(cfg.name)
	if !ok {
		fmt.Fprintf(stderr, "asterism: unknown constellation %q\n", cfg.name)
		if matches := sky.CloseMatches(cfg.name); len(matches) > 0 {
			fmt.Fprintf(stderr, "Did you mean: %s\n", strings.Join(matches, ", "))
		}
		fmt.Fprintln(stderr, "Run 'asterism --list' to see all available constellations.")
		return 1
	}

	renderOne(stdout, c, opts, cfg)
	return 0
}

func (cfg config) observer() sky.Observer {
	return sky.Observer{Lat: cfg.lat, Lon: cfg.lon, Time: cfg.at}
}

// renderOne renders a single constellation chart followed by its info panel.
// With an observer the chart shows the figure as it stands in that sky
// (zenith up) and the panel adds where to look.
func renderOne(w *os.File, c sky.Constellation, opts sky.Options, cfg config) {
	shown := c
	if cfg.observe {
		shown = sky.HorizonView(c, cfg.observer())
	}
	buf := sky.Render(shown, opts)
	fmt.Fprintln(w, buf.String(opts.Color))
	for _, line := range sky.InfoLines(c) {
		fmt.Fprintln(w, line)
	}
	if cfg.observe {
		for _, line := range sky.ObserverLines(c, cfg.observer()) {
			fmt.Fprintln(w, line)
		}
		fmt.Fprintln(w, "(Chart drawn as it stands in the sky: zenith up, the horizon below.)")
	}
}

// printVisible lists the constellations whose center is above the horizon,
// highest first.
func printVisible(w *os.File, o sky.Observer) {
	type row struct {
		c sky.Constellation
		v sky.Visibility
	}
	var rows []row
	for _, c := range sky.Catalog() {
		if v := sky.See(c, o); v.Alt > 0 {
			rows = append(rows, row{c, v})
		}
	}
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].v.Alt > rows[j].v.Alt })
	fmt.Fprintf(w, "Up from %s at %s (%d of %d):\n\n",
		sky.FormatLatLon(o.Lat, o.Lon), o.Time.Format("2006-01-02 15:04 MST"), len(rows), len(sky.Catalog()))
	for _, r := range rows {
		note := ""
		if r.v.Alt < 15 {
			note = "  low"
		}
		fmt.Fprintf(w, "  %-4s  %-20s %3.0f°  %-3s  %d/%d stars up%s\n",
			r.c.Abbr, r.c.Latin, r.v.Alt, sky.Compass(r.v.Az), r.v.Above, r.v.Total, note)
	}
}

// printList prints every available constellation with its abbreviation.
func printList(w *os.File) {
	fmt.Fprintf(w, "Available constellations (%d):\n\n", len(sky.Catalog()))
	for _, c := range sky.Catalog() {
		fmt.Fprintf(w, "  %-4s  %-18s  %s\n", c.Abbr, c.Latin, c.Common)
	}
}

// parseArgs parses argv into a config. Flags may appear before or after the
// positional constellation name. Color, lines, and labels default to on.
func parseArgs(args []string) (config, error) {
	cfg := config{
		width:  0, // filled from terminal size below
		height: 0,
		color:  true,
		lines:  true,
		labels: true,
	}
	var wFlag, hFlag int
	var sizeSet bool
	var latSet, lonSet bool

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--help" || arg == "-h":
			cfg.help = true
		case arg == "--version" || arg == "-v":
			cfg.version = true
		case arg == "--list" || arg == "-l":
			cfg.list = true
		case arg == "--all" || arg == "-a":
			cfg.all = true
		case arg == "--no-color":
			cfg.color = false
		case arg == "--no-lines":
			cfg.lines = false
		case arg == "--no-labels":
			cfg.labels = false
		case arg == "--ascii":
			cfg.ascii = true
		case arg == "--visible":
			cfg.visible = true
		case arg == "--lat" || arg == "--lon" || arg == "--time" ||
			strings.HasPrefix(arg, "--lat=") || strings.HasPrefix(arg, "--lon=") || strings.HasPrefix(arg, "--time="):
			name, val, ok := strings.Cut(arg, "=")
			if !ok {
				if i+1 >= len(args) {
					return cfg, fmt.Errorf("%s requires a value", arg)
				}
				i++
				val = args[i]
			}
			if err := applyObserver(name, val, &cfg, &latSet, &lonSet); err != nil {
				return cfg, err
			}
		case arg == "--width" || arg == "--height" || arg == "--size":
			if i+1 >= len(args) {
				return cfg, fmt.Errorf("%s requires a value", arg)
			}
			i++
			val := args[i]
			if err := applyDimension(arg, val, &wFlag, &hFlag, &sizeSet); err != nil {
				return cfg, err
			}
		case strings.HasPrefix(arg, "--width="):
			if err := applyDimension("--width", strings.TrimPrefix(arg, "--width="), &wFlag, &hFlag, &sizeSet); err != nil {
				return cfg, err
			}
		case strings.HasPrefix(arg, "--height="):
			if err := applyDimension("--height", strings.TrimPrefix(arg, "--height="), &wFlag, &hFlag, &sizeSet); err != nil {
				return cfg, err
			}
		case strings.HasPrefix(arg, "--size="):
			if err := applyDimension("--size", strings.TrimPrefix(arg, "--size="), &wFlag, &hFlag, &sizeSet); err != nil {
				return cfg, err
			}
		case strings.HasPrefix(arg, "-") && arg != "-":
			return cfg, fmt.Errorf("unknown flag %q", arg)
		default:
			// Positional constellation name; join multi-word names.
			if cfg.name == "" {
				cfg.name = arg
			} else {
				cfg.name += " " + arg
			}
		}
	}

	if latSet != lonSet {
		return cfg, fmt.Errorf("--lat and --lon go together")
	}
	cfg.observe = latSet
	if !cfg.observe && (cfg.visible || !cfg.at.IsZero()) {
		return cfg, fmt.Errorf("--visible and --time need --lat and --lon")
	}
	if cfg.observe && cfg.at.IsZero() {
		cfg.at = time.Now()
	}

	// Resolve terminal size: explicit flags win, else environment, else default.
	tw, th := terminalSize()
	if wFlag > 0 {
		tw = wFlag
	}
	if hFlag > 0 {
		th = hFlag
	}
	cfg.width = tw
	// Reserve space below the chart for the info panel unless height was set
	// explicitly (an explicit --height means "chart is this tall").
	if hFlag > 0 || sizeSet {
		cfg.height = th
	} else {
		cfg.height = th - panelReserve
	}
	if cfg.width < 10 {
		cfg.width = 10
	}
	if cfg.height < 6 {
		cfg.height = 6
	}
	return cfg, nil
}

// timeLayouts are the accepted --time forms. Forms without a zone are read in
// the local time zone.
var timeLayouts = []string{
	time.RFC3339,
	"2006-01-02T15:04Z07:00",
	"2006-01-02T15:04:05",
	"2006-01-02T15:04",
	"2006-01-02 15:04:05",
	"2006-01-02 15:04",
}

// applyObserver parses --lat, --lon, and --time.
func applyObserver(flag, val string, cfg *config, latSet, lonSet *bool) error {
	switch flag {
	case "--lat", "--lon":
		n, err := strconv.ParseFloat(strings.TrimSpace(val), 64)
		limit := 90.0
		if flag == "--lon" {
			limit = 180
		}
		if err != nil || math.IsNaN(n) || n < -limit || n > limit {
			return fmt.Errorf("%s: expected degrees between -%g and %g, got %q", flag, limit, limit, val)
		}
		if flag == "--lat" {
			cfg.lat, *latSet = n, true
		} else {
			cfg.lon, *lonSet = n, true
		}
	case "--time":
		for _, layout := range timeLayouts {
			if t, err := time.ParseInLocation(layout, strings.TrimSpace(val), time.Local); err == nil {
				if y := t.Year(); y < sky.MinYear || y > sky.MaxYear {
					return fmt.Errorf("--time: %d is outside %d-%d (positions ignore precession)", y, sky.MinYear, sky.MaxYear)
				}
				cfg.at = t
				return nil
			}
		}
		return fmt.Errorf("--time: expected e.g. 2026-11-12T21:00 or 2026-11-12T21:00:00+09:00, got %q", val)
	}
	return nil
}

// applyDimension parses a --width / --height / --size value into wFlag/hFlag.
func applyDimension(flag, val string, wFlag, hFlag *int, sizeSet *bool) error {
	switch flag {
	case "--width":
		n, err := parsePositive(val)
		if err != nil {
			return fmt.Errorf("--width: %w", err)
		}
		*wFlag = n
	case "--height":
		n, err := parsePositive(val)
		if err != nil {
			return fmt.Errorf("--height: %w", err)
		}
		*hFlag = n
	case "--size":
		parts := strings.FieldsFunc(val, func(r rune) bool { return r == 'x' || r == 'X' || r == ',' })
		if len(parts) != 2 {
			return fmt.Errorf("--size: expected WxH, got %q", val)
		}
		w, err := parsePositive(parts[0])
		if err != nil {
			return fmt.Errorf("--size width: %w", err)
		}
		h, err := parsePositive(parts[1])
		if err != nil {
			return fmt.Errorf("--size height: %w", err)
		}
		*wFlag, *hFlag = w, h
		*sizeSet = true
	}
	return nil
}

func parsePositive(s string) (int, error) {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0, fmt.Errorf("%q is not a number", s)
	}
	if n <= 0 {
		return 0, fmt.Errorf("must be positive, got %d", n)
	}
	return n, nil
}

// terminalSize reads COLUMNS and LINES from the environment, falling back to a
// sensible default. It deliberately avoids any external dependency for size
// detection, as required by the project's stdlib-only constraint.
func terminalSize() (int, int) {
	w := envInt("COLUMNS", defaultWidth)
	h := envInt("LINES", defaultHeight)
	return w, h
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil && n > 0 {
			return n
		}
	}
	return fallback
}

func usage() string {
	return `asterism - draw constellations in your terminal

USAGE:
  asterism [flags] NAME
  asterism --list
  asterism --all
  asterism --lat 35.68 --lon 139.77 [--time 2026-11-12T21:00] NAME
  asterism --lat 35.68 --lon 139.77 --visible

NAME is a constellation's common/Latin name or IAU abbreviation
(case-insensitive), for example: orion, "ursa major", uma, cru.

FLAGS:
  -l, --list        list all available constellations and exit
  -a, --all         render every constellation
      --width N     chart width in columns (default: $COLUMNS or 80)
      --height N    chart height in rows (default: $LINES or 40)
      --size WxH    set width and height together, e.g. --size 100x50
      --no-color    disable ANSI color output
      --no-lines    do not draw the figure lines
      --no-labels   do not draw star name labels
      --ascii       use ASCII-only glyphs (for terminals without good Unicode)
      --lat DEG     observer latitude (north positive); with --lon, the chart
                    is drawn as the figure stands in that sky (zenith up) and
                    the panel says where to look and when it rises and sets
      --lon DEG     observer longitude (east positive)
      --time T      moment to use (default: now), e.g. 2026-11-12T21:00 (local
                    time) or 2026-11-12T21:00:00+09:00
      --visible     list the constellations above the horizon, highest first
  -h, --help        show this help and exit
  -v, --version     show version and exit

Coordinates and magnitudes are approximate catalog values chosen for
recognizable rendering. Positions in the sky ignore precession, nutation, and
refraction (about half a degree off now; --time accepts 1900-2100). asterism is
a visualization / education tool.
`
}

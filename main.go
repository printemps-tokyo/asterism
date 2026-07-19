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
	"os"
	"strconv"
	"strings"

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
			renderOne(stdout, c, opts)
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

	renderOne(stdout, c, opts)
	return 0
}

// renderOne renders a single constellation chart followed by its info panel.
func renderOne(w *os.File, c sky.Constellation, opts sky.Options) {
	buf := sky.Render(c, opts)
	fmt.Fprintln(w, buf.String(opts.Color))
	for _, line := range sky.InfoLines(c) {
		fmt.Fprintln(w, line)
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
  -h, --help        show this help and exit
  -v, --version     show version and exit

Coordinates and magnitudes are approximate catalog values chosen for
recognizable rendering. asterism is a visualization / education tool.
`
}

# asterism

Draw constellations in your terminal.

Type a constellation's name and asterism plots its bright stars (bigger, bolder,
brighter-colored glyphs for brighter stars), connects them with the traditional
stick-figure lines, labels the named stars, and prints a short info panel with
the Latin name, IAU abbreviation, brightest star, and best viewing season.

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
| `-h`, `--help` | show help and exit |
| `-v`, `--version` | show version and exit |

The positional name is case-insensitive and accepts the common name, the Latin
name, or the IAU three-letter abbreviation. Spaces and hyphens are ignored, so
`ursa major`, `Ursa-Major`, and `ursamajor` all resolve to the same figure.

## Included constellations

24 of the most recognizable constellations across both hemispheres:

Andromeda, Aquila, Aries, Auriga, Bootes, Cancer, Canis Major, Cassiopeia,
Cepheus, Corona Borealis, Crux, Cygnus, Draco, Gemini, Leo, Lyra, Orion,
Pegasus, Perseus, Sagittarius, Scorpius, Taurus, Ursa Major, Ursa Minor.

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

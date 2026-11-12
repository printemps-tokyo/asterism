# Contributing

Thanks for your interest in contributing to a printemps-tokyo project.

## Development

1. Fork / clone the repository
2. Make sure you have Go 1.23 or newer installed
3. Create a branch: `git switch -c feat/your-change`
4. Make your change and verify locally:
   ```bash
   gofmt -l .
   go vet ./...
   go test ./...
   go build ./...
   ```
5. Commit and open a pull request

## Commit messages

[Conventional Commits](https://www.conventionalcommits.org/) are preferred
(`feat:`, `fix:`, `docs:`, `chore:`, `refactor:`, `test:`).

## Pull requests

- Keep one PR focused on a single purpose
- CI (gofmt / vet / test / build) must be green
- Update the README and tests when behavior changes

## Adding or correcting constellation data

The catalog has two parts. `internal/sky/catalog.go` holds 24 constellations
curated by hand; edit those directly. `internal/sky/catalog_gen.go` holds the
other 64 and is generated, so do not edit it: change `scripts/gen_catalog.py`
and regenerate. The script reads the data files of d3-celestial at a pinned
commit (see the script's docstring for the four files and where to get them):

```bash
python3 scripts/gen_catalog.py path/to/d3-celestial/data > internal/sky/catalog_gen.go
gofmt -l .
go test ./...
```

Coordinates are J2000 right ascension (hours) and declination (degrees);
magnitudes are apparent visual magnitudes. Values are approximate catalog
figures chosen for recognizable rendering, not for astrometric precision. Keep
every IAU abbreviation unique and make sure every figure line segment
references stars that exist (the tests check both).

## Bug reports and requests

Please open an [issue](../../issues) using the templates.

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

The catalog lives in `internal/sky/catalog.go`. Coordinates are J2000
right ascension (hours) and declination (degrees); magnitudes are apparent
visual magnitudes. Values are approximate catalog figures chosen for
recognizable rendering, not for astrometric precision. When adding a
constellation, keep the IAU abbreviation unique and make sure every figure
line segment references stars that exist.

## Bug reports and requests

Please open an [issue](../../issues) using the templates.

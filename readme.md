# The Anti-Mage Project

A permanent, non-profit Go project holding browser-fingerprint reference data
and the analysis that reads it. Every reference value carries its own
provenance, enforced by the type system rather than by convention: a table
whose values have not been observed on a real system of the configuration it
describes cannot be read as evidence by the code that consumes it.

## Packages

- `reference` — constant tables of browser and platform signals
  (font lists, native-function `toString` forms, trusted error names, CSS
  system font keywords) verified against primary sources: a specification
  clause, a vendor's own documentation, or a measurement this project took of
  a named engine build. Every exported table is a `Table`, carrying a
  `Source` (origin and the date it was checked) and a `Verified` flag.

- `osfont` — reads the font tables in `reference` and answers two questions
  about a set of resolved font families. `EvaluateWindows` gives a per-release
  verdict: `Present`, `Absent`, `Inconclusive`, or `Unverified` — never a score
  or a probability. `ReleaseFloor` gives the oldest release the observation is
  compatible with, counting a release as reached when any family it introduced
  resolved. Neither treats absence from a table as suspicious.

## Usage

```go
import "github.com/N4darae/anti-mage/osfont"

res := osfont.EvaluateWindows(resolvedFamilies)
res.AtLeast("10") // present, absent, inconclusive, or unverified
res.Skipped       // families excluded as independent of the OS

f := osfont.ReleaseFloor(resolvedFamilies)
f.Release         // oldest release the observation supports, "" if none
f.AboveGap        // releases reported above a gap; they do not narrow
```

Font detection by advance width is unreliable per family rather than per
release: a substituting font stack answers for a family the machine does not
have, an icon font carries no glyph for an ASCII probe string, and a
script-supplemental package arrives only once its language is enabled. So
`ReleaseFloor` reads the presence of a release and never the shape of what is
missing, and it has no way to report a release as impossible.

## Provenance model

Every reference value names where it came from and when it was last checked:
a spec section, a vendor page URL, or the JavaScript engine build this project
measured.
A table whose `Verified` field is `false` records values that have not been
observed on a real system of the configuration it describes. Any code reading
it gets back `Unverified` rather than a positive or negative verdict, so such a
table cannot be treated as evidence.

## Building and testing

Requires Go 1.24. The module has zero external dependencies.

```
make check
```

runs `gofmt`, `go vet`, and `go test` across the module. Run the pieces
individually with `make fmt`, `make vet`, or `make test`.

## License

MIT. See `LICENSE` and `NOTICE`.

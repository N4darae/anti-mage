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

- `osfont` — reads the font tables in `reference` and turns a set of resolved
  font families into a verdict on the Windows release that produced them. It
  reports `Present`, `Absent`, `Inconclusive`, or `Unverified` — never a score
  or a probability — and treats anything absent from a table as unknown rather
  than suspicious.

## Usage

```go
import "github.com/N4darae/anti-mage/osfont"

res := osfont.EvaluateWindows(resolvedFamilies)
res.AtLeast("10") // present, absent, inconclusive, or unverified
res.Skipped       // families excluded as independent of the OS
```

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

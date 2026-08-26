# The Anti-Mage Project

A permanent, non-profit Go project holding browser-fingerprint reference data
and the analysis that reads it. Every reference value carries its own
provenance, enforced by the type system rather than by convention: a table
whose values have not been observed on a real system of the configuration it
describes cannot be read as evidence by the code that consumes it.

## Running it

```
git clone https://github.com/N4darae/anti-mage && cd anti-mage && go run .
```

prints one URL on loopback:

```
open http://127.0.0.1:8787/ in the browser you want to examine
```

Open it in that browser and press the button. The page measures the browser it
is loaded in, sends what it measured to the local server that served it, and
shows one assessment of the whole environment: a determination, a score, and a
sentence saying what they mean. Below the assessment it lists everything the
browser reported, so a reader can see what the assessment was computed from.

Requires Go 1.24 and nothing else. The module has zero external dependencies,
and the page and its collector are compiled into the binary with `go:embed`, so
a clone needs no build step, no configuration and no network beyond the loopback
address it binds. `-addr` changes the address, which must be loopback;
`-web <dir>` serves the page from disk instead of the copy compiled in.

Two endpoints are there if you would rather drive it yourself than press the
button. `GET /api/bootstrap` returns the inputs the server chose for one scan —
a nonce, invented font-family names, and the instants at which to sample a UTC
offset. `POST /api/scan` takes the observations and returns the assessment:

```json
{"v": 1,
 "determination": "coherent",
 "score": 0,
 "statement": "Everything that could be read agrees with the platform this environment claims.",
 "supplied": ["auto.residue", "geom.css", "geom.screen", "…"]}
```

Send the nonce back as a top-level `nonce` field so the server can confirm the
scan ran with the inputs it issued.

## Usage

`assess` is the way in. You hand over whatever you observed, however you
observed it, and get back one assessment computed across all of it.

```go
import "github.com/N4darae/anti-mage/assess"

env := assess.Environment{
    Observations: map[string]assess.Observation{
        "scope.main":      {Status: assess.StatusOK, Value: mainThreadFacts},
        "native.tostring": {Status: assess.StatusOK, Value: accessorSources},
        "font.resolved":   {Status: assess.StatusUnsupported},
    },
    // The questions you chose for this scan, and the clock you kept. These
    // must be yours: a question the examined environment picked is not one it
    // could not have prepared an answer for.
    FontControls: controls,
    OffsetDates:  dates,
    ElapsedMS:    int(time.Since(issued) / time.Millisecond),
}

a := assess.Evaluate(env)

a.Determination // coherent, discrepant, instrumented, insufficient, not-evaluated
a.Score         // 0 to 100, in steps of ten
a.Statement     // one sentence to show a reader
a.Supplied      // the observation ids you handed over, echoed back
```

The score is there to be built on. It is monotone — evidence that disagrees can
only raise it, evidence that agrees can only lower it — and it is quantised, so
a threshold set against it stays where you put it. The determination is ordered,
so a policy is a comparison:

```go
if a.Determination.AtLeast(assess.Discrepant) && a.Score >= 30 {
    // your policy here
}
if !a.Determination.Established() {
    // too little was read to characterise this environment either way;
    // that is its own case, not a quiet pass
}
```

Evidence the library did not collect goes in beside the evidence it did:

```go
env.Findings = append(env.Findings,
    assess.Finding{Name: "audio pipeline", Verdict: assess.Contradiction},
    assess.Finding{Name: "codec matrix", Verdict: assess.Unverified},
)
```

A finding weighs exactly what one of the library's own bodies of evidence
weighs, and `Unverified` keeps a reading out of the arithmetic entirely rather
than letting it count as a failure to establish anything.

If the observations arrive as JSON from the environment being examined, decode
them with `assess.Decode`. It reads the observations and the echoed nonce and
nothing else: the questions, the clock and any findings are yours to set, and a
payload that carries them is decoded as though it had not.

### One narrow reading

`osfont` answers a single question about a single input — which Windows release
a set of resolved font families is compatible with — and answers it as a floor
rather than as a verdict:

```go
import "github.com/N4darae/anti-mage/osfont"

f := osfont.ReleaseFloor(resolvedFamilies)
f.Release  // oldest release the observation supports, "" if none
f.AboveGap // releases reported above a gap; they do not narrow the floor
```

Font detection by advance width is unreliable per family rather than per
release: a substituting font stack answers for a family the machine does not
have, an icon font carries no glyph for an ASCII probe string, and a
script-supplemental package arrives only once its language is enabled. So
`ReleaseFloor` reads the presence of a release and never the shape of what is
missing, and it has no way to report a release as impossible. It is one input
among several, and `assess.Evaluate` is what weighs it against the rest.

## What the assessment says

One determination and one score, for the whole environment. The score is
computed from counts of independent bodies of evidence rather than from counts
of individual checks, and quantised to steps of ten, so that one body of
evidence cannot be separated from another by watching the last digit. A caller
sees the observations it supplied and the assessment; nothing in the assessment
says which reading moved the number, because a value that said so would be a
tuning table for anyone shaping an environment against it. The observations are
a different matter: every one of them was computed by the environment being
examined, so echoing them back tells that environment nothing it did not
already know.

Absence reads as inconclusive. An observation that could not be made, a value
that did not parse, a scope the browser refused to create and a reference table
this project has not confirmed all leave the score where it is, and a browser
that lacks a feature is never scored for lacking it. The price of that rule is
that too little evidence has to be uncertain in both directions rather than
reassuring, which is what `Established()` is for: `insufficient` means nothing
disagreed and not enough was read to call that agreement.

The strongest statement an assessment makes is that an environment appears
modified. Privacy, accessibility and content-blocking tools modify the same
surfaces, in large numbers, so that statement describes the environment and not
the person using it, and it names no vendor, product or tool as the cause.

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

- `assess` — the composite. `Evaluate` is a pure function of its argument: no
  globals, no clock, no filesystem, no network, and total over hostile input.
  It reads every body of evidence the observations reach and returns one
  determination and one score.

- `server` — the HTTP surface, bound to loopback. It chooses each scan's
  server-side inputs, serves the page, keeps the clock the scan is timed
  against, and holds the process's only mutable state: the set of inputs it has
  issued recently, which it passes to `assess.Evaluate` as part of that
  function's one argument rather than letting the assessment reach for it.

- `internal/scan` — the engine `assess` runs. It turns one payload into a
  reading per body of evidence, each carrying its own primary source and its own
  rule for when to abstain, and combines those readings into the summary
  `assess` reports. A copy of the IANA zone database is compiled in through
  `time/tzdata` so a zone lookup succeeds with no zone files installed; it is a
  fallback behind `ZONEINFO` and the host's own zone directories, which the
  command accounts for by clearing `ZONEINFO` before serving.

## Provenance model

Every reference value names where it came from and when it was last checked:
a spec section, a vendor page URL, or the JavaScript engine build this project
measured.
A table whose `Verified` field is `false` records values that have not been
observed on a real system of the configuration it describes. Any code reading
it gets back `Unverified` rather than a positive or negative verdict, so such a
table cannot be treated as evidence.

## Building and testing

```
make check
```

runs `gofmt`, `go vet`, and `go test` across the module. Run the pieces
individually with `make fmt`, `make vet`, or `make test`.

## License

MIT. See `LICENSE` and `NOTICE`.

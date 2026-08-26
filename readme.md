# The Anti-Mage Project

A permanent, non-profit Go project that scores how coherent a browser
environment is, together with the reference data that scoring reads. A client
asks for a score; it gets back one determination, one estimated score, and one
sentence saying what the two mean.

Every reference value carries its own provenance, enforced by the type system:
a table whose values have not been observed on a real system of the
configuration it describes cannot be read as evidence by the code that consumes
it.

Requires Go 1.24 and nothing else. Zero external dependencies; the page and its
collector are compiled in with `go:embed`.

## Scoring a client

Two calls. The client asks for the inputs of one scan, measures itself, and
posts what it measured.

**`GET /api/bootstrap`** returns the inputs this server chose for one scan: a
nonce, and the questions the scan asks. The questions are the server's, because
a question the examined environment picked is not one it could not have prepared
an answer for. A client merges the bootstrap over its own defaults and keeps the
nonce. Also answers `POST`, and is aliased at `/bootstrap.json`.

**`POST /api/scan`** takes the observations and returns the assessment. Send the
nonce back as a top-level `nonce` field, or as `X-Anti-Mage-Nonce`, or as a
`nonce` query parameter.

```json
{"v": 1, "nonce": "…",
 "probes": {"<id>": {"status": "ok", "value": …},
            "<id>": {"status": "unsupported", "value": {"reason": "…"}}}}
```

`ok` is the only status read as evidence. `unsupported` and `error` say
something could not be measured, which is a fact about the client's browser and
never a mark against it. An unknown observation id is ignored, so a newer client
still scores against an older server.

```json
{"v": 1,
 "determination": "coherent",
 "score": 0,
 "statement": "Everything that could be read agrees with the platform this environment claims.",
 "supplied": ["…", "…"]}
```

`determination` is one of `coherent`, `discrepant`, `instrumented`,
`insufficient`, `not-evaluated`. `score` is an estimate from 0 to 100 in steps
of ten. `supplied` echoes the ids the client sent, which tells it nothing it did
not compute itself.

The server keeps the clock each scan is timed against, so a client cannot price
its own elapsed time, and remembers the inputs it issued recently so a payload
can be tied to a scan that actually ran.

## The page

```
git clone https://github.com/N4darae/anti-mage && cd anti-mage && go run .
```

prints one loopback URL. Open it in the browser you want to examine and press
the button: the page is a client of the two endpoints above, and shows the
determination, the score, the sentence, and everything the browser reported. It
reads a bootstrap injected into the page first and falls back to
`GET /api/bootstrap`.

`-addr` changes the address, which must be loopback; `-web <dir>` serves the
page from disk instead of the copy compiled in.

## In-process

`assess` is the way in for a caller that already has the observations.

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

`Evaluate` is a pure function of its argument: no globals, no clock, no
filesystem, no network, and total over hostile input. The score is monotone —
evidence that disagrees can only raise it, evidence that agrees can only lower
it — and quantised, so a threshold stays where you put it. The determination is
ordered, so a policy is a comparison:

```go
if a.Determination.AtLeast(assess.Discrepant) && a.Score >= 30 {
    // your policy here
}
if !a.Determination.Established() {
    // too little was read to characterise this environment either way
}
```

Evidence the library did not collect goes in through `env.Findings`, where a
finding weighs what one of the library's own bodies of evidence weighs and a
verdict of `Unverified` keeps a reading out of the arithmetic entirely. If the
observations arrive as JSON from the environment being examined, decode them
with `assess.Decode`: it reads the observations and the echoed nonce and nothing
else, so the questions, the clock and any findings stay the caller's.

## What the score means

The score counts independent bodies of evidence rather than individual checks,
and is quantised to steps of ten, so one body of evidence cannot be separated
from another by watching the last digit. Nothing in the assessment says which
reading moved the number: a value that said so would be a tuning table for
anyone shaping an environment against it.

It is an estimate, and the arithmetic errs in the direction that costs an honest
visitor least. Absence reads as inconclusive — an observation that could not be
made, a value that did not parse, a scope the browser refused to create, a
reference table this project has not confirmed — and a browser is never scored
for lacking a feature. The price is that too little evidence is uncertain in
both directions rather than reassuring, which is what `Established()` is for. A
score of zero says nothing disagreed; it never says an environment is
unmodified.

The strongest statement an assessment makes is that an environment appears
modified. Privacy, accessibility and content-blocking tools modify the same
surfaces, in large numbers, so that statement describes the environment and not
the person using it, and it names no vendor, product or tool as the cause.

## Packages

- `assess` — the way in. `Evaluate` scores a whole environment; `Decode` reads a
  client payload without letting it set anything the caller owns.
- `server` — the loopback HTTP surface. It chooses each scan's inputs, serves the
  page, keeps the clock, and holds the only mutable state: the inputs issued
  recently.
- `reference` — constant tables verified against primary sources, each carrying a
  `Source` and a `Verified` flag. A table whose `Verified` is false yields
  `Unverified` rather than a verdict, so it cannot be treated as evidence.
- `osfont` — one narrow reading: the oldest Windows release a set of resolved
  font families is compatible with, as a floor rather than a verdict. Font
  detection is unreliable per family rather than per release, so it reads the
  presence of a release and never the shape of what is missing, and it cannot
  call a release impossible.
- `internal/scan` — the engine `assess` runs: one reading per body of evidence,
  each with its own primary source and its own rule for when to abstain. The
  IANA zone database is compiled in through `time/tzdata`, behind `ZONEINFO` and
  the host's own zone directories.

## Building and testing

`make check` runs `gofmt`, `go vet` and `go test` across the module. `make fmt`,
`make vet` and `make test` run the pieces.

## License

MIT. See `LICENSE` and `NOTICE`.

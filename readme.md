# The Anti-Mage Project

A permanent, non-profit Go project that scores how coherent a browser
environment is, together with the reference data that scoring reads. A client
asks for a score; it gets back one determination, one estimated score, and one
sentence saying what the two mean.

Every reference value carries its own provenance, enforced by the type system
rather than by convention: a table whose values have not been observed on a real
system of the configuration it describes cannot be read as evidence by the code
that consumes it.

## Scoring a client

Two calls. The client asks for the inputs of one scan, measures itself, and
posts what it measured. The answer is the score.

**`GET /api/bootstrap`** returns the inputs this server chose for one scan: a
nonce, and the questions the scan asks. It answers `POST` identically, and the
same object is served at `/bootstrap.json` for a client that would rather not
use the `/api` prefix. The questions are the server's and not
the client's, because a question the examined environment picked is not one it
could not have prepared an answer for. A client merges the bootstrap over its
own defaults rather than replacing them, and keeps the nonce.

**`POST /api/scan`** takes the observations and returns the assessment. Send the
nonce back as a top-level `nonce` field so the server can confirm the scan ran
with the inputs it issued; it is also accepted as `X-Anti-Mage-Nonce` or a
`nonce` query parameter.

```json
{"v": 1,
 "nonce": "…",
 "probes": {"<id>": {"status": "ok", "value": …},
            "<id>": {"status": "unsupported", "value": {"reason": "…"}}}}
```

Each observation carries its own status. `ok` is the only one read as evidence:
`unsupported` and `error` say that something could not be measured, which is a
fact about the client's browser and never a mark against it. An observation id
this server does not know is ignored, so a newer client still scores against an
older server.

The response is the whole answer:

```json
{"v": 1,
 "determination": "coherent",
 "score": 0,
 "statement": "Everything that could be read agrees with the platform this environment claims.",
 "supplied": ["…", "…"]}
```

`determination` is one of `coherent`, `discrepant`, `instrumented`,
`insufficient` and `not-evaluated`. `score` is an estimate on a 0-100 scale in
steps of ten. `statement` is one sentence fit to show a reader. `supplied` echoes
the observation ids the client sent, which tells the client nothing it did not
already know, since it computed every one of them itself.

The server keeps the clock each scan is timed against, so a client cannot price
its own elapsed time, and it holds the inputs it has issued recently so that a
payload can be tied to a scan it actually ran.

## The page

```
git clone https://github.com/N4darae/anti-mage && cd anti-mage && go run .
```

prints one URL on loopback:

```
open http://127.0.0.1:8787/ in the browser you want to examine
```

Open it in that browser and press the button. The page is a client of the same
two endpoints: the server hands it this scan's inputs in the page itself, and
falls back to `GET /api/bootstrap` if the page did not carry them. It measures
the browser it is loaded in, posts what it measured to `POST /api/scan`, and
shows the determination, the score and the sentence. Below them it
lists everything the browser reported, so a reader can see what the score was
computed from.

Requires Go 1.24 and nothing else. The module has zero external dependencies,
and the page and its collector are compiled into the binary with `go:embed`, so
a clone needs no build step, no configuration and no network beyond the loopback
address it binds. `-addr` changes the address, which must be loopback;
`-web <dir>` serves the page from disk instead of the copy compiled in.

## In-process

`assess` is the way in for a caller that already has the observations and wants
the same score without an HTTP round trip. You hand over whatever you observed,
however you observed it, and get back one assessment computed across all of it.

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
filesystem, no network, and total over hostile input.

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

## What the score means

One determination and one score, for the whole environment. The score is
computed from counts of independent bodies of evidence rather than from counts
of individual checks, and quantised to steps of ten, so that one body of
evidence cannot be separated from another by watching the last digit. A caller
sees the observation ids it supplied and the assessment; nothing in the
assessment says which reading moved the number, because a value that said so
would be a tuning table for anyone shaping an environment against it.

It is an estimate, and the arithmetic is built so that it errs in the direction
that costs an honest visitor least. Absence reads as inconclusive: an
observation that could not be made, a value that did not parse, a scope the
browser refused to create and a reference table this project has not confirmed
all leave the score where it is, and a browser that lacks a feature is never
scored for lacking it. The price of that rule is that too little evidence has to
be uncertain in both directions rather than reassuring, which is what
`Established()` is for: `insufficient` means nothing disagreed and not enough was
read to call that agreement. The bottom of the scale carries the same rule the
other way round: a score of zero says that nothing this project looked at
disagreed, never that the environment is unmodified, because no set of checks it
can run establishes that.

The strongest statement an assessment makes is that an environment appears
modified. Privacy, accessibility and content-blocking tools modify the same
surfaces, in large numbers, so that statement describes the environment and not
the person using it, and it names no vendor, product or tool as the cause.

## Packages

- `assess` — the way in. `Evaluate` returns one determination and one score for
  a whole environment; `Decode` reads a client payload without letting that
  payload set anything the caller owns.

- `server` — the HTTP surface, bound to loopback. It chooses each scan's
  server-side inputs, serves the page, keeps the clock the scan is timed
  against, and holds the process's only mutable state: the set of inputs it has
  issued recently, which it passes to `assess.Evaluate` as part of that
  function's one argument rather than letting the assessment reach for it.

- `reference` — constant tables of browser and platform signals verified
  against primary sources: a specification clause, a vendor's own
  documentation, or a measurement this project took of a named engine build.
  Every exported table is a `Table`, carrying a `Source` (origin and the date it
  was checked) and a `Verified` flag.

- `osfont` — one narrow reading over one input: which Windows release a set of
  resolved font families is compatible with, answered as a floor rather than as
  a verdict. `ReleaseFloor` gives the oldest release the observation supports;
  `EvaluateWindows` gives a per-release `Present` / `Absent` / `Inconclusive` /
  `Unverified` verdict. Font detection is unreliable per family rather than per
  release — a substituting stack answers for a family the machine does not have,
  an icon font carries no glyph for an ASCII probe string, and a
  script-supplemental package arrives only once its language is enabled — so it
  reads the presence of a release and never the shape of what is missing, and it
  has no way to report a release as impossible. It is one input among several,
  and `assess.Evaluate` is what weighs it against the rest.

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
measured. A table whose `Verified` field is `false` records values that have not
been observed on a real system of the configuration it describes. Any code
reading it gets back `Unverified` rather than a positive or negative verdict, so
such a table cannot be treated as evidence.

## Building and testing

```
make check
```

runs `gofmt`, `go vet`, and `go test` across the module. Run the pieces
individually with `make fmt`, `make vet`, or `make test`.

## License

MIT. See `LICENSE` and `NOTICE`.

# The Anti-Mage Project 

Tired of anti-detect browsers? So am I. That is why I am publishing some of the
techniques that catch them.

There are many more of them, against the large commercial browsers and against
the open-source ones that call themselves anti-detect alike. I know this will get
reverse-engineered. That is fine and it is a starting shot, to show how much room
is left in detecting a bot on consistency alone. I hope it gives you somewhere to
begin, and that you take it further.

A permanent, non-profit Go project that scores how coherent a browser
environment is, together with the reference data that scoring reads. A client
asks for a score; it gets back one determination, one estimated score, and one
sentence saying what the two mean.

## What it catches

Consistency, mostly. A browser's own surfaces are checked against each other
and against what the browser claims about itself. Example: a screen against its own
viewport, a version against what it can do, one graphics interface against
another, a device against its own decoders, a font set against the platform
it should imply. None of these need a signature or a known tool; they only
need two things a browser said to disagree.



## Results

100 samples per browser on Windows. A score is always a multiple of ten, so
range and median land on that step exactly; the mean is given to one decimal and
the last column rounds it back onto the step. Every figure in the table is
rounded, not raw. 0 means nothing disagreed.

| browser | n | range | median | mean | rounded |
| --- | ---: | ---: | ---: | ---: | ---: |
| Chrome, stock | 100 | 0 | 0 | 0.0 | **0** |
| Firefox, stock | 100 | 0 | 0 | 0.0 | **0** |
| Edge, stock | 100 | 0 | 0 | 0.0 | **0** |
| Brave | 100 | 10–30 | 10 | 13.4 | **10** |
| AdsPower | 100 | 50–70 | 60 | 60.0 | **60** |
| CloakBrowser | 100 | 40–80 | 70 | 67.0 | **70** |
| NSTBrowser | 100 | 40–90 | 80 | 73.0 | **70** |
| Camoufox | 100 | 60–90 | 80 | 80.0 | **80** |

Every anti-detect browser tested came back modified, none with a median below
60, and every stock browser came back at 0 across all 100 samples, no false
positives to trade for it. The scores are not a ranking of those tools; a higher
number only means more of the environment contradicted itself.

Brave's median sits at 10 because it modifies the surfaces it says it modifies,
which is the lightest thing this project records and not a claim against it.

<img width="1186" height="679" alt="image" src="https://github.com/user-attachments/assets/a7a5ebf2-3843-469d-b806-5b5ca0e26fd6" />


## Requirements

* **Go 1.24+**
* **Zero external dependencies** (the collector UI and reference datasets are embedded via `go:embed`).


## Scoring a client

Two calls: the client asks for the inputs of one scan, measures itself, and posts
what it measured.

**`GET /api/bootstrap`** (also answers `POST`, aliased at `/bootstrap.json`)
returns the inputs this server chose for one scan: a nonce, and the questions the
scan asks. The questions are the server's, so the environment cannot have
prepared its answers. A client merges the bootstrap over its own defaults and
keeps the nonce.

**`POST /api/scan`** takes the observations and returns the assessment. Send the
nonce back as a top-level `nonce` field, as `X-Anti-Mage-Nonce`, or as a `nonce`
query parameter.

```json
{"v": 1, "nonce": "…",
 "probes": {"<id>": {"status": "ok", "value": …},
            "<id>": {"status": "unsupported", "value": {"reason": "…"}}}}
```

`ok` is the only status read as evidence, `observations` is accepted in place of
`probes`, and an unknown id is ignored, so a newer client still scores against an
older server.

```json
{"v": 1,
 "determination": "coherent",
 "score": 0,
 "statement": "Everything that could be read agrees with the platform this environment claims.",
 "supplied": ["…", "…"]}
```

`determination` is one of `coherent`, `discrepant`, `instrumented`,
`insufficient`, `not-evaluated`. `score` is an estimate from 0 to 100 in steps of
ten. `supplied` echoes the ids the client sent, which tells it nothing it did not
compute itself.

The server keeps the clock and remembers the inputs it issued recently, so a
client cannot price its own elapsed time and a payload can be tied to a scan that
actually ran.

## The page

```
git clone https://github.com/N4darae/anti-mage && cd anti-mage && go run .
```

prints one loopback URL. Open it in the browser you want to examine and press
the button: the page is a client of the two endpoints above, and shows the
determination, the score and the sentence. It reads a bootstrap injected into
the page first and falls back to `GET /api/bootstrap`.

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
    FontControls: controls,
    OffsetDates:  dates,
    ElapsedMS:    int(time.Since(issued) / time.Millisecond),
}

a := assess.Evaluate(env)
```

The controls, the dates and the clock must be yours, chosen before the
environment was asked anything. `a` carries the determination, the score, one
sentence to show a reader, and the ids you handed over.

`Evaluate` is a pure function of its argument: no globals, no clock, no
filesystem, no network, and total over hostile input. The score is monotone —
only evidence raises it, no evidence lowers it — and quantised, so a threshold
stays where you put it. The determination is ordered, so a policy is a
comparison:

```go
if a.Determination.AtLeast(assess.Discrepant) && a.Score >= 30 {
    reject()
}
if !a.Determination.Established() {
    tooLittleWasRead()
}
```

Evidence the library did not collect goes in through `env.Findings`. If the
observations arrive as JSON from the environment being examined, decode them with
`assess.Decode`: it reads the observations and the echoed nonce and nothing else,
so the questions, the clock and any findings stay yours.

## What the score means

The score weighs independent bodies of evidence rather than individual checks,
in steps of ten, so one body of evidence cannot be separated from another by
watching the last digit. Nothing in the assessment says which reading moved the
number, and nothing here describes what any reading looks at: either would be a
tuning table for anyone shaping an environment against it.
<img width="1152" height="832" alt="image" src="https://github.com/user-attachments/assets/524e94cf-b7db-410f-8e23-f94e38983aa4" />


The score is an estimate, and errs in the direction that costs an honest visitor
least. A browser is never scored for lacking a feature, and too little evidence
is uncertain in both directions rather than reassuring, which is what
`Established()` is for. A score of zero says nothing disagreed; it never says an
environment is unmodified.

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
  `Source` and a `Verified` flag. A table whose `Verified` is false cannot be
  treated as evidence.
- `osfont` — one narrow reading over resolved font families, as a floor rather
  than a verdict.
- `internal/scan` — the engine `assess` runs. The IANA zone database is compiled
  in through `time/tzdata`, behind `ZONEINFO` and the host's own zone
  directories.

## Building and testing

`make check` runs `gofmt`, `go vet` and `go test` across the module. `make fmt`,
`make vet` and `make test` run the pieces.

## License

MIT. See `LICENSE` and `NOTICE`.

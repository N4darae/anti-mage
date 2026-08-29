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

<img width="1186" height="679" alt="image" src="https://github.com/user-attachments/assets/a7a5ebf2-3843-469d-b806-5b5ca0e26fd6" />


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

The score weighs independent bodies of evidence rather than individual checks,
in steps of ten, so one body of evidence cannot be separated from another by
watching the last digit. Nothing in the assessment says which reading moved the
number: a value that said so would be a tuning table for anyone shaping an
environment against it.
<img width="1152" height="832" alt="image" src="https://github.com/user-attachments/assets/524e94cf-b7db-410f-8e23-f94e38983aa4" />


Scores read on one Windows machine:

| environment | score |
| --- | --- |
| Chrome, stock | 0 |
| Firefox, stock | 0 |
| Edge, stock | 0 |
| Brave | 10 |
| CloakBrowser | 60 |
| AdsPower | 70 |
| NSTBrowser | 70 |
| Camoufox | 90 |

Brave sits at ten because it modifies the surfaces it says it modifies, which is
the lightest thing this project records and not a claim against it.

The score is an estimate, and errs in the direction that costs an honest visitor
least. Absence reads as inconclusive — an observation that could not be made, a
value that did not parse, a scope the browser refused to create, a reference
table this project has not confirmed — and a browser is never scored for lacking
a feature. The price is that too little evidence is uncertain in both directions
rather than reassuring, which is what `Established()` is for. A score of zero
says nothing disagreed; it never says an environment is unmodified.

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

## A viewport against its own screen

The simplest reading in the project needs no reference table at all.

A page is given an area to draw in, and it is told the size of the device it is
drawn on. Both are in CSS pixels and both scale together, so a viewport larger
than its own screen is arithmetic rather than a matter of degree or of any
vendor's behaviour.

The reading declines wherever the arithmetic stops being decisive: only the
drawing area is compared, never the window box or the work area, because a window
may legitimately overlap a taskbar or hang off an edge. A negative offset from the
origin puts the window on a display this reading was never given the size of. A
screen or viewport of no size settles nothing. And a payload missing any of the
four numbers carries no weight either way.

## Capabilities against the version claimed

A browser names a version, and a version implies a set of capabilities. The
reading compares the two, in one direction only: a capability that shipped
*later* than the version claimed, found present, is disagreeing evidence. A
capability the claimed version should have and does not is never read as
anything.

The asymmetry is the whole design. A capability can be missing on an honest
machine for reasons that belong to a policy, a build or a platform rather than to
the engine's age, so absence establishes nothing. An engine that answers for
something introduced after the version it claims cannot be explained that way,
and that is where the common failure lies: a tool keeping a newer engine and
rewriting the version string leaves exactly this trace.

The table records which major version each capability shipped in, sourced to the
vendor's release notes, and only rows observed on a real system of that version
are verified. The reading also declines when no version parses, when no
capability was reported, and when no verified row is later than the version
claimed -- the ordinary case on the newest release.

The checks live in the collector, as code rather than as expressions sent from
here. The capabilities are public knowledge, so choosing them here would buy
nothing, and a collector that evaluated strings from the server would no longer
be one whose behaviour can be read from its own source.

## Readings that record without scoring

Two readings here reach no verdict at all, by construction.

A surface can be worth watching before this project knows what an honest browser
does on it. Candidate gathering produces a state sequence and a set of candidate
kinds; two serialisation paths for one drawing surface produce two byte streams.
Environments differ on both, and on neither has the range an unmodified browser
produces across builds, drivers and configurations been established. A verdict
drawn from that would be a guess wearing the clothes of evidence.

So they collect, they show what they read, and they carry no weight in any
direction; their tests assert that adding either moves neither the band nor the
confidence. When someone establishes the honest range, the rows they have been
recording are what the question gets settled against.

## Two graphics interfaces, one device

A machine has one graphics device, and a browser offers two interfaces onto it.
An environment where the older interface names a hardware device by vendor and
model, while the newer interface grants no adapter at all, is reporting two
different machines through two windows onto the same one.

That much is read as a modification rather than as a false claim, and carries the
lightest weight there is, because an honest machine can reach it: a policy may
disable the newer interface, a driver may be excluded from it, and a device may
support one backend without the other. What the reading establishes is that
something between the two interfaces has been changed, which is a fact about the
environment and not about the person using it.

The request for a software fallback is read last, and only once the three
requests that ask for a device have come back empty. An environment granted that
fallback was never short of a device to grant: the backend it received is proof
the interface reaches the machine, so the refusals before it were a choice about
what to disclose. That pairing carries the heaviest weight here, because no
policy, driver or device configuration produces it — a disabled interface grants
nothing, and a working one does not refuse its own hardware and then answer.

The reading abstains outside a secure context, where the newer interface is
gated; when the browser does not expose it at all; when no device was named for
it to disagree with; and when the device named draws in software.

## A device against its own decoders

One reading is worth describing on its own, because it shows what the `Verified`
flag is for.

A graphics device names a generation, and a generation carries a documented set
of hardware video decoders. An environment that names a device, demonstrates a
working hardware decoder for one codec, and reports none for another codec its
own generation carries has stated three things that cannot all describe one
machine. The control codec is what makes this a reading rather than a complaint
about a missing feature: an environment with no hardware decoder at all abstains.

It abstains in every other direction too — no device named, a device outside the
table, a codec this build cannot decode by any path, a generation the table does
not record as carrying the decoder, a reading the collector did not report.

And it abstains on any generation whose entry is not `Verified`. Documented
capability is not observed capability: hardware decoding can be unavailable on a
real machine for reasons belonging to the driver or the platform rather than the
device. An entry becomes evidence only once this project has watched a real
system of that configuration report the decoder, and records which system that
was. Today one generation meets that bar; the others are tabulated, sourced, and
read as `Unverified`.

## Building and testing

`make check` runs `gofmt`, `go vet` and `go test` across the module. `make fmt`,
`make vet` and `make test` run the pieces.

## License

MIT. See `LICENSE` and `NOTICE`.

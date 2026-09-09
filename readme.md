# The Anti-Mage project

Tired of anti-detect browsers? So am I. That is why I am publishing some of the techniques that catch them.

A non-profit Go project that scores how coherent a browser environment is, together with the reference data the scoring reads. A client asks for a score and gets back one determination, one number, and one sentence saying what the two mean. There is no signature list, no known-tool database, and nothing to keep updated. The whole thing runs on making a browser disagree with itself.

I know this will get reverse engineered. That is fine, and it is a starting shot. Everything here is MIT and the source is the specification.

## What it catches

23 independent readings, each looking at one place where a browser's own surfaces have to agree with each other and with the platform the browser claims to be:

| | |
|---|---|
| Platform claim | Installed fonts against the claimed platform |
| Agreement across execution scopes | Native accessor integrity |
| Numeric built-in behaviour | Exception types against the specification |
| Screen geometry against CSS | Viewport against the screen it claims |
| Layout and text metric identities | Agreement between the CSS and script paths |
| Time zone against measured offsets | Audio buffer coherence |
| Automation residue | Permission state coherence |
| ICE gathering against its own reported state | Candidate gathering against the state machine that runs it |
| Two serialisations of one drawing surface | Reported media capabilities |
| Hardware decoders against the device named | Both graphics interfaces against one device |
| The device named against the generation reported for it | Capabilities against the version claimed |
| What this browser reported it could not do | |

The browser-side collector takes 34 measurements to feed those readings. Native accessor integrity alone is four separate readings of every accessor a spoofer has to touch: how the function serialises, whether three different enumerators of its own keys agree, whether the property still sits on the interface prototype object where the interface definition puts it, and whether calling it with the wrong receiver throws the TypeError the specification mandates.

None of that needs to know what tool produced the environment. It only needs two things the browser said to disagree.

<img width="1186" height="679" alt="image" src="https://github.com/user-attachments/assets/a7a5ebf2-3843-469d-b806-5b5ca0e26fd6" />

## How the score is built

Each reading returns one of five determinations. `consistent` and `contradiction` are what they sound like. `instrumented` means the surface reports its own modification. `inconclusive` means the reading ran and decided nothing. `unverified` means the reading declined to weigh in, and that last one matters most, because it is the difference between a detector and a random number generator.

A reading abstains when the probe was not collected, when the browser does not expose the feature, when the page was not delivered to a secure context and the feature is gated there, when nothing was named to compare against, or when the reference table it would have read has not been verified against a system I have actually observed. A browser is never scored for lacking a feature.

Findings that survive that are weighed by class rather than counted. A surface that reports its own modification weighs least. A plain disagreement weighs more. A disagreement that nothing but a deliberate change produces weighs most. Those combine so that each further body of evidence adds less than the one before it, and the result is rounded onto a step of ten, so no single reading can be isolated by watching the last digit. The constants are in `internal/scan/band.go` if you want them.

Two properties hold by construction and are covered by tests. Only evidence raises the score, so withdrawing a reading can never raise it. And a self-declared hook explains its own downstream damage: when an accessor reports that it has been modified and a later reading contradicts itself in a way that accessor accounts for, the finding is downgraded from a lie to a declared modification. Suppression cannot buy confidence, and a privacy tool that is honest about what it patches is not treated as a liar for it.

The score tops out at 90. Nothing here will ever tell you an environment is certainly modified.

## Results

100 samples per browser on Windows. Scores land on multiples of ten, so range and median sit on that step exactly. The mean is given to one decimal and the last column rounds it back onto the step. 0 means nothing disagreed.

| Browser | n | Range | Median | Mean | Rounded |
|---------|---|-------|--------|------|---------|
| Chrome, stock | 100 | 0 | 0 | 0.0 | **0** |
| Firefox, stock | 100 | 0 | 0 | 0.0 | **0** |
| Edge, stock | 100 | 0 | 0 | 0.0 | **0** |
| Brave | 100 | 10-30 | 10 | 13.4 | **10** |
| AdsPower | 100 | 50-70 | 60 | 60.0 | **60** |
| CloakBrowser | 100 | 40-80 | 70 | 67.0 | **70** |
| NSTBrowser | 100 | 40-90 | 80 | 73.0 | **70** |
| Camoufox | 100 | 60-90 | 80 | 80.0 | **80** |

Every anti-detect browser tested came back modified, none with a median below 60, and every stock browser came back at 0 across all 100 samples, with no false positives traded for it.

Those numbers are a floor rather than a best case. 7 of the 22 reference entries are not yet marked verified, which means the readings that would have used them abstained on all 100 samples. Verifying them can only push the anti-detect column up.

The scores are not a ranking of those tools. A higher number only means more of the environment contradicted itself. Brave's median sits at 10 because Brave modifies the surfaces it says it modifies, which is the lightest thing this project records and not a claim against it.

## Running it

```bash
git clone https://github.com/N4darae/anti-mage && cd anti-mage && go run .
```

Go 1.24 or newer, and no external dependencies at all. It prints one loopback URL, `127.0.0.1:8787` by default. Open that in the browser you want to examine and press the button.

- `-addr` changes the address, which must be loopback. Both the address you ask for and the socket that actually gets bound are checked, so this cannot be talked onto a public interface.
- `-web <dir>` serves the page from disk instead of the copy compiled in.
- `-dump <dir>` writes every payload posted to it, for replay through `tools/debugscan` later.

The server binds loopback only, caps a scan body at 1 MiB, sets header, read, write and idle timeouts, recovers from a panic in a handler, and sends `nosniff`, `no-referrer` and `no-store` on everything it serves.

## The HTTP surface

Two calls. The client asks for the inputs of one scan, measures itself, and posts what it measured.

`GET /api/bootstrap` (also answers `POST`, aliased at `/bootstrap.json`) returns the inputs this server chose for one scan. That is a 128-bit nonce from `crypto/rand`, six font family names invented for this scan alone, and eight randomised dates to measure a time zone offset against. The invented names are the point: a font probe cannot pre-compute an answer for a family that was generated a millisecond ago and exists nowhere. A client merges the bootstrap over its own defaults and keeps the nonce.

`POST /api/scan` takes the observations and returns the assessment. Send the nonce back as a top-level `nonce` field, as `X-Anti-Mage-Nonce`, or as a `nonce` query parameter.

```json
{"v": 1, "nonce": "…",
 "probes": {"<id>": {"status": "ok", "value": …},
            "<id>": {"status": "unsupported", "value": {"reason": "…"}}}}
```

`ok` is the only status read as evidence, `observations` is accepted in place of `probes`, and an unknown id is ignored, so a newer client still scores against an older server.

```json
{"v": 1,
 "determination": "coherent",
 "score": 0,
 "statement": "Everything that could be read agrees with the platform this environment claims.",
 "supplied": […]}
```

`determination` is one of `coherent`, `discrepant`, `instrumented`, `insufficient`, `not-evaluated`. `score` is an estimate in steps of ten. `supplied` echoes the ids the client sent, which tells it nothing it did not compute itself.

The server keeps the clock and remembers the inputs it issued for the last 30 minutes, up to 4096 live at a time. A client cannot price its own elapsed time, and a payload can be tied to a scan that actually ran.

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

The controls, the dates and the clock must be yours, chosen before the environment was asked anything. `a` carries the determination, the score, one sentence to show a reader, and the ids you handed over.

`Evaluate` is a pure function of its argument: no globals, no clock, no filesystem, no network, and total over hostile input. The determination is ordered, so a policy is a comparison:

```go
if a.Determination.AtLeast(assess.Discrepant) && a.Score >= 30 {
    reject()
}
if !a.Determination.Established() {
    tooLittleWasRead()
}
```

Evidence the library did not collect goes in through `env.Findings`, with a verdict of `consistent`, `contradiction`, `inconclusive`, `unverified` or `modified`, and it is weighed exactly like a built-in reading. If the observations arrive as JSON from the environment being examined, decode them with `assess.Decode`. It reads the observations and the echoed nonce and nothing else, so the questions, the clock and any findings stay yours.

## Reference data

Every constant the scoring compares against carries its provenance in the source. A reference entry has the primary source it came from, the date that source was checked, and a `Verified` flag that is only true when I have observed the value myself on a system of the configuration described, recorded in an `Observed` line beside it:

```go
Observed: "GeForce RTX 3060 Ti (0x00002489), Windows 10 22H2, Chrome 151.0.7922.174: ..."
```

An entry whose `Verified` is false cannot be treated as evidence, and that is enforced at every site that reads a table rather than left to convention. Vendor documentation on its own is a citation, not a measurement, which is why the results above are a floor.

## Layout

- `assess` is the way in. `Evaluate` scores a whole environment, `Decode` reads a client payload without letting it set anything the caller owns.
- `server` is the loopback HTTP surface. It chooses each scan's inputs, serves the page, keeps the clock, and holds the only mutable state in the project: the inputs issued recently.
- `reference` holds the constant tables and their provenance.
- `osfont` is one narrow reading over resolved font families, deliberately a floor rather than a verdict.
- `internal/scan` is the engine, one file per reading.
- `web` is the browser-side collector, embedded with `go:embed`.
- `tools/debugscan` replays a dumped payload and prints every reading that fired.

6,400 lines of Go carry 7,600 lines of tests over them: 403 test functions, 100% statement coverage of `reference`, 98% of `assess`, 97% of `osfont`, 92% of the engine. Some of those tests are calibration rather than examples, asserting that withdrawing a reading cannot raise a score, that a declared modification scores below a contradiction, and that a probe reporting itself unsupported cannot move the number at all.

`make check` runs `gofmt`, `go vet` and `go test` across the module.

## What it will not tell you

Nothing in an assessment says which reading moved the number. That is deliberate, and it is the one place this readme stays quiet on purpose, because a per-reading breakdown is a tuning table for anyone shaping an environment against it. The source is public and MIT, so a determined reader will find the mechanism anyway. Being handed it in a response, one scan at a time, is a different thing.

The score is an estimate, and it errs in the direction that costs an honest visitor least. Too little evidence comes out uncertain in both directions rather than reassuring, which is what `Established()` is for. A score of zero says nothing disagreed. It never says an environment is unmodified.

The strongest statement an assessment makes is that an environment appears modified. Privacy, accessibility and content-blocking tools modify the same surfaces, in large numbers, so that statement describes the environment and not the person using it, and it names no vendor, product or tool as the cause.

There is a lot of room left in catching a bot on consistency alone. I hope this gives you somewhere to begin, and that you take it further.

## License

MIT. See `LICENSE` and `NOTICE`.

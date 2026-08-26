// Package assess reads whatever a caller observed about a browser environment

package assess

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/N4darae/anti-mage/internal/scan"
)

const (
	StatusOK = scan.StatusOK

	StatusUnsupported = scan.StatusUnsupported

	StatusError = scan.StatusError
)

type Observation struct {
	Status string          `json:"status"`
	Value  json.RawMessage `json:"value,omitempty"`
}

type Verdict string

const (
	Consistent Verdict = "consistent"

	Contradiction Verdict = "contradiction"

	Inconclusive Verdict = "inconclusive"

	Unverified Verdict = "unverified"

	Modified Verdict = "modified"
)

type Finding struct {
	Name string `json:"name"`

	Verdict Verdict `json:"verdict"`
}

type Environment struct {
	Observations map[string]Observation `json:"observations,omitempty"`

	Findings []Finding `json:"findings,omitempty"`

	Nonce string `json:"nonce,omitempty"`

	OffsetDates  []string `json:"offsetDates,omitempty"`
	FontControls []string `json:"fontControls,omitempty"`

	ElapsedMS int `json:"elapsedMs,omitempty"`
}

type Determination string

const (
	NotEvaluated Determination = "not-evaluated"

	Insufficient Determination = "insufficient"

	Coherent Determination = "coherent"

	Discrepant Determination = "discrepant"

	Instrumented Determination = "instrumented"
)

type Assessment struct {
	V int `json:"v"`

	Determination Determination `json:"determination"`

	Score int `json:"score"`

	Statement string `json:"statement"`

	Supplied []string `json:"supplied"`
}

func (d Determination) rank() int {
	switch d {
	case NotEvaluated:
		return 0
	case Insufficient:
		return 1
	case Coherent:
		return 2
	case Discrepant:
		return 3
	case Instrumented:
		return 4
	}
	return 0
}

func (d Determination) AtLeast(other Determination) bool {
	return d.rank() >= other.rank()
}

func (d Determination) Established() bool {
	switch d {
	case Coherent, Discrepant, Instrumented:
		return true
	}
	return false
}

func (d Determination) String() string { return string(d) }

func Evaluate(env Environment) Assessment {
	req := scan.Request{
		V:      1,
		Mode:   "public",
		Nonce:  env.Nonce,
		Probes: probesOf(env),
	}
	in := scan.Inputs{
		Nonce:        env.Nonce,
		OffsetDates:  env.OffsetDates,
		FontControls: env.FontControls,
		ElapsedMS:    elapsedOf(env),
	}
	rep := scan.AnalyzeWith(req, in, suppliedSections(env.Findings))

	a := Assessment{
		V:        1,
		Score:    clampScore(rep.Summary.BotLikeness),
		Supplied: suppliedIDs(env.Observations),
	}
	a.Determination = determinationOf(rep.Summary.Band)
	if a.Determination == NotEvaluated {

		a.Score = 0
	}
	a.Statement = statements[a.Determination]
	return a
}

var statements = map[Determination]string{
	NotEvaluated: "Nothing that was supplied reached a reading, so this says nothing about this environment.",
	Insufficient: "Nothing disagreed, and too little was established to characterise this environment either way.",
	Coherent:     "Everything that could be read agrees with the platform this environment claims.",
	Discrepant:   "One body of evidence reports two facts that cannot both be true. Everything else that could be read agrees.",
	Instrumented: "This environment appears modified. Privacy, accessibility and content-blocking tools modify the same surfaces, in large numbers, so this describes the environment and not the person using it.",
}

func (e Environment) NamesReported(id string) []string {
	return scan.Request{Probes: probesOf(e)}.ReportedNames(id)
}

func probesOf(env Environment) map[string]scan.Probe {
	out := make(map[string]scan.Probe, len(env.Observations))
	for id, o := range env.Observations {
		if id == "" {
			continue
		}
		out[id] = scan.Probe{Status: normaliseStatus(o.Status), Value: o.Value}
	}
	return out
}

func normaliseStatus(s string) string {
	switch t := strings.ToLower(strings.TrimSpace(s)); t {
	case StatusOK, StatusUnsupported, StatusError:
		return t
	default:
		return s
	}
}

func elapsedOf(env Environment) int {
	if env.ElapsedMS < 0 {
		return 0
	}
	return env.ElapsedMS
}

func suppliedSections(fs []Finding) []scan.Section {
	if len(fs) == 0 {
		return nil
	}
	byName := make(map[string]Verdict, len(fs))
	for _, f := range fs {
		name := strings.TrimSpace(f.Name)
		if name == "" {
			continue
		}
		byName[name] = f.Verdict
	}
	names := make([]string, 0, len(byName))
	for n := range byName {
		names = append(names, n)
	}
	sort.Strings(names)

	out := make([]scan.Section, 0, len(names))
	for _, n := range names {
		out = append(out, scan.Section{
			ID:            "supplied:" + n,
			Title:         n,
			Determination: determinationFor(byName[n]),
		})
	}
	return out
}

func determinationFor(v Verdict) scan.Determination {
	switch v {
	case Consistent:
		return scan.Consistent
	case Contradiction:
		return scan.Contradiction
	case Unverified:
		return scan.Unverified
	case Modified:
		return scan.Instrumented
	}
	return scan.Inconclusive
}

func determinationOf(band string) Determination {
	switch band {
	case scan.BandInsufficient:
		return Insufficient
	case scan.BandCoherent:
		return Coherent
	case scan.BandDiscrepant:
		return Discrepant
	case scan.BandInstrumented:
		return Instrumented
	}
	return NotEvaluated
}

func clampScore(n int) int {
	if n < 0 {
		return 0
	}
	if n > 100 {
		return 100
	}
	return n
}

func suppliedIDs(obs map[string]Observation) []string {
	out := make([]string, 0, len(obs))
	for id := range obs {
		if id == "" {
			continue
		}
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}

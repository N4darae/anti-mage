package assess

import (
	"encoding/json"

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

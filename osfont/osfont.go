// Package osfont turns resolved font families into a determination about a
// claimed Windows version, using the tables in reference.
package osfont

import (
	"sort"

	"github.com/N4darae/anti-mage/reference"
)

// Verdict is the result of one table lookup.
type Verdict int

const (
	Inconclusive Verdict = iota // zero value: the observation answers nothing
	Present                     // every required family resolved
	Absent                      // no required family resolved
	Unverified                  // table's values are unobserved
)

func (v Verdict) String() string {
	switch v {
	case Present:
		return "present"
	case Absent:
		return "absent"
	case Unverified:
		return "unverified"
	default:
		return "inconclusive"
	}
}

// Result is what the tables say about one observation.
type Result struct {
	// Versions maps a release to its verdict.
	Versions map[string]Verdict
	// BaseFonts is the verdict for families on every supported release.
	BaseFonts Verdict
	// Skipped lists families excluded because they ship independent of the OS.
	Skipped []string
}

// AtLeast reports the verdict recorded for release. Present means that
// release or a later one, since a later release keeps the earlier fonts.
func (r Result) AtLeast(release string) Verdict { return r.Versions[release] }

// EvaluateWindows reads resolved families against the Windows tables. An
// empty input yields Inconclusive throughout: a probe that did not run.
func EvaluateWindows(resolved []string) Result {
	res := Result{Versions: map[string]Verdict{}}
	have := make(map[string]bool, len(resolved))
	for _, f := range resolved {
		have[f] = true
	}
	for _, np := range reference.WindowsNonProbativeMarkers.Values {
		if have[np] {
			res.Skipped = append(res.Skipped, np)
			delete(have, np)
		}
	}
	sort.Strings(res.Skipped)
	if len(have) == 0 {
		for ver := range reference.WindowsVersionMarkerFonts {
			res.Versions[ver] = Inconclusive
		}
		return res
	}
	for ver, table := range reference.WindowsVersionMarkerFonts {
		res.Versions[ver] = tally(table, have)
	}
	res.BaseFonts = tally(reference.WindowsBaseFonts, have)
	return res
}

// tally is Present only for a complete match, Absent only for none,
// Unverified if t is unobserved, else Inconclusive.
func tally(t reference.Table, have map[string]bool) Verdict {
	if len(t.Values) == 0 {
		return Inconclusive
	}
	if !t.Verified {
		return Unverified
	}
	n := 0
	for _, f := range t.Values {
		if have[f] {
			n++
		}
	}
	switch n {
	case len(t.Values):
		return Present
	case 0:
		return Absent
	default:
		return Inconclusive
	}
}

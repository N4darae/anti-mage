// Package osfont turns resolved font families into a determination about a

package osfont

import (
	"sort"

	"github.com/N4darae/anti-mage/reference"
)

type Verdict int

const (
	Inconclusive Verdict = iota
	Present
	Absent
	Unverified
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

type Result struct {
	Versions map[string]Verdict

	BaseFonts Verdict

	Skipped []string
}

func (r Result) AtLeast(release string) Verdict { return r.Versions[release] }

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

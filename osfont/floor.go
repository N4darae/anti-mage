package osfont

import "github.com/N4darae/anti-mage/reference"

// Floor is the oldest Windows release an observation is compatible with.
//
// Release is empty when the observation supports no release at all. AboveGap
// names releases whose families resolved while a release below them did not:
// they are reported and never used to narrow, because a gap means a family was
// not measured rather than that the machine is not what it claims.
type Floor struct {
	Release  string
	AboveGap []string
	Skipped  []string
}

// releaseOrder is the order the vendor introduced the tables in, oldest first.
// The lists are cumulative: a release ships everything its predecessors shipped.
var releaseOrder = []string{"10", "11"}

// ReleaseFloor reports the oldest Windows release the resolved families are
// compatible with, and never reports a release as impossible.
//
// A release counts as reached when ANY family it introduced resolved, not when
// all of them did. That is the whole difference between this function and
// EvaluateWindows, and it exists because font detection by advance width is
// unreliable per family rather than per release. A substituting font stack
// answers for a family the machine does not have; an icon font carries no glyph
// for an ASCII probe string to measure; a script-supplemental package arrives
// only once its language is enabled; OEM images differ. Each of those removes a
// family from the observation on a machine that is exactly what it claims to be.
//
// So the shape of what is missing carries no information here. Requiring the
// complete set per release, or treating a later release resolving while an
// earlier one did not as impossible, reads an unreliable probe as a lying
// machine — a fully-installed desktop is the ordinary case for that mistake, not
// a corner case.
//
// Callers wanting the stricter per-release verdicts, including the Unverified
// state for a table with no observation behind it, should use EvaluateWindows.
// This function answers a narrower question and answers it without accusing.
func ReleaseFloor(resolved []string) Floor {
	have := make(map[string]bool, len(resolved))
	for _, f := range resolved {
		have[f] = true
	}

	var f Floor
	for _, np := range reference.WindowsNonProbativeMarkers.Values {
		if have[np] {
			f.Skipped = append(f.Skipped, np)
			delete(have, np)
		}
	}

	reached := make([]bool, len(releaseOrder))
	for i, rel := range releaseOrder {
		t, ok := reference.WindowsVersionMarkerFonts[rel]
		if !ok {
			continue
		}
		for _, fam := range t.Values {
			if have[fam] {
				reached[i] = true
				break
			}
		}
	}

	// The base families predate every release in the table, so they are the
	// floor below the floor: without them no release is supported at all.
	base := false
	for _, fam := range reference.WindowsBaseFonts.Values {
		if have[fam] {
			base = true
			break
		}
	}
	if !base && !anyTrue(reached) {
		return f
	}

	last := -1
	for i := range releaseOrder {
		if !reached[i] {
			break
		}
		last = i
	}
	if last >= 0 {
		f.Release = releaseOrder[last]
	}
	for i := last + 2; i < len(releaseOrder); i++ {
		if reached[i] {
			f.AboveGap = append(f.AboveGap, releaseOrder[i])
		}
	}
	return f
}

func anyTrue(bs []bool) bool {
	for _, b := range bs {
		if b {
			return true
		}
	}
	return false
}

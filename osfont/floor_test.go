package osfont

import (
	"testing"

	"github.com/N4darae/anti-mage/reference"
)

func allOf(rel string) []string { return reference.WindowsVersionMarkerFonts[rel].Values }

// The mistake this function exists to avoid: a machine with everything the
// vendor publishes installed must not read as impossible.
func TestFullyInstalledReachesTheNewestRelease(t *testing.T) {
	var all []string
	all = append(all, reference.WindowsBaseFonts.Values...)
	for _, rel := range releaseOrder {
		all = append(all, allOf(rel)...)
	}
	f := ReleaseFloor(all)
	if f.Release != releaseOrder[len(releaseOrder)-1] {
		t.Errorf("Release = %q, want %q", f.Release, releaseOrder[len(releaseOrder)-1])
	}
	if len(f.AboveGap) != 0 {
		t.Errorf("AboveGap = %v on a complete set, want none", f.AboveGap)
	}
}

// One family per release is enough. Requiring the complete set is what turns an
// unreliable probe into a false accusation.
func TestOneFamilyPerReleaseIsEnough(t *testing.T) {
	obs := []string{reference.WindowsBaseFonts.Values[0]}
	for _, rel := range releaseOrder {
		obs = append(obs, allOf(rel)[0])
	}
	if f := ReleaseFloor(obs); f.Release != releaseOrder[len(releaseOrder)-1] {
		t.Errorf("Release = %q with one family from each release, want %q",
			f.Release, releaseOrder[len(releaseOrder)-1])
	}
}

// A gap is reported and never narrows, and never becomes a verdict of its own.
func TestGapIsReportedAndDoesNotNarrow(t *testing.T) {
	newest := releaseOrder[len(releaseOrder)-1]
	obs := append([]string{reference.WindowsBaseFonts.Values[0]}, allOf(newest)...)
	f := ReleaseFloor(obs)
	if f.Release != "" {
		t.Errorf("Release = %q; the release below it did not resolve, so nothing is narrowed", f.Release)
	}
	if len(f.AboveGap) != 1 || f.AboveGap[0] != newest {
		t.Errorf("AboveGap = %v, want [%s]", f.AboveGap, newest)
	}
}

// Nothing observed supports nothing. An empty observation is a probe that did
// not run, not a machine with no fonts.
func TestEmptyObservationSupportsNothing(t *testing.T) {
	for _, obs := range [][]string{nil, {}, {"DejaVu Sans", "Liberation Sans"}} {
		f := ReleaseFloor(obs)
		if f.Release != "" || len(f.AboveGap) != 0 {
			t.Errorf("ReleaseFloor(%v) = %+v, want the zero Floor", obs, f)
		}
	}
}

// Families that ship independently of the operating system are set aside and
// reported, never counted toward a release.
func TestNonProbativeFamiliesAreSetAside(t *testing.T) {
	np := reference.WindowsNonProbativeMarkers.Values
	f := ReleaseFloor(np)
	if f.Release != "" {
		t.Errorf("Release = %q from families that ship independently of the OS", f.Release)
	}
	if len(f.Skipped) != len(np) {
		t.Errorf("Skipped = %v, want all of %v", f.Skipped, np)
	}
}

// No input may make this function name a release as impossible: it has no way
// to say so, and that is the point. This pins the absence of that vocabulary.
func TestNoObservationProducesAnAccusation(t *testing.T) {
	var every []string
	every = append(every, reference.WindowsBaseFonts.Values...)
	every = append(every, reference.WindowsNonProbativeMarkers.Values...)
	for _, rel := range releaseOrder {
		every = append(every, allOf(rel)...)
	}
	// Every subset of a small sample, plus the whole set, must return a Floor
	// and never panic.
	for i := range every {
		for j := i; j < len(every) && j < i+6; j++ {
			f := ReleaseFloor(every[i : j+1])
			if f.Release != "" && f.Release != "10" && f.Release != "11" {
				t.Fatalf("Release = %q, which is not a release in the table", f.Release)
			}
		}
	}
}

package osfont

import (
	"testing"

	"github.com/N4darae/anti-mage/reference"
)

func allOf(rel string) []string { return reference.WindowsVersionMarkerFonts[rel].Values }

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

func TestEmptyObservationSupportsNothing(t *testing.T) {
	for _, obs := range [][]string{nil, {}, {"DejaVu Sans", "Liberation Sans"}} {
		f := ReleaseFloor(obs)
		if f.Release != "" || len(f.AboveGap) != 0 {
			t.Errorf("ReleaseFloor(%v) = %+v, want the zero Floor", obs, f)
		}
	}
}

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

func TestNoObservationProducesAnAccusation(t *testing.T) {
	var every []string
	every = append(every, reference.WindowsBaseFonts.Values...)
	every = append(every, reference.WindowsNonProbativeMarkers.Values...)
	for _, rel := range releaseOrder {
		every = append(every, allOf(rel)...)
	}

	for i := range every {
		for j := i; j < len(every) && j < i+6; j++ {
			window := every[i : j+1]
			f := ReleaseFloor(window)
			if f.Release != "" && !anyFamilyOfRelease(f.Release, window) {
				t.Fatalf("Release = %q from %v, where no family of that release was observed", f.Release, window)
			}
			for _, rel := range f.AboveGap {
				if !anyFamilyOfRelease(rel, window) {
					t.Fatalf("AboveGap names %q from %v, where no family of that release was observed", rel, window)
				}
			}
		}
	}
}

func anyFamilyOfRelease(rel string, observed []string) bool {
	for _, fam := range reference.WindowsVersionMarkerFonts[rel].Values {
		for _, o := range observed {
			if o == fam {
				return true
			}
		}
	}
	return false
}

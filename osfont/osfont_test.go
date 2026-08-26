package osfont

import (
	"testing"

	"github.com/N4darae/anti-mage/reference"
)

func TestEmptyInputIsInconclusive(t *testing.T) {
	got := EvaluateWindows(nil)
	if got.BaseFonts != Inconclusive {
		t.Errorf("base fonts on empty input = %v, want inconclusive", got.BaseFonts)
	}
	for ver, v := range got.Versions {
		if v != Inconclusive {
			t.Errorf("Windows %s on empty input = %v, want inconclusive", ver, v)
		}
	}
	if len(got.Versions) != len(reference.WindowsVersionMarkerFonts) {
		t.Errorf("got %d version rows, want %d", len(got.Versions), len(reference.WindowsVersionMarkerFonts))
	}
}

func TestZeroValueIsInconclusive(t *testing.T) {
	var r Result
	if r.BaseFonts != Inconclusive {
		t.Errorf("zero Result.BaseFonts = %v, want inconclusive", r.BaseFonts)
	}
	if Verdict(0) != Inconclusive {
		t.Error("Inconclusive must be the zero Verdict")
	}
}

func TestUnverifiedTablePropagatesThroughTally(t *testing.T) {
	tbl := reference.Table{
		Values:   []string{"Example Family"},
		Source:   reference.Source{Origin: "test fixture", Checked: "2026-08-25"},
		Verified: false,
	}
	have := map[string]bool{"Example Family": true}
	if got := tally(tbl, have); got != Unverified {
		t.Errorf("fully-matched unverified table = %v, want unverified", got)
	}
}

func TestUnverifiedTableNeverPresentEvenWhenFullyMatched(t *testing.T) {
	eleven := reference.WindowsVersionMarkerFonts["11"]
	if eleven.Verified {
		t.Fatal("this test assumes the Windows 11 row is unverified")
	}
	got := EvaluateWindows(eleven.Values)
	if got.Versions["11"] != Unverified {
		t.Errorf("complete Windows 11 marker set = %v, want unverified", got.Versions["11"])
	}
}

func TestWindows10ObservationLeavesElevenUnverified(t *testing.T) {
	ten := reference.WindowsVersionMarkerFonts["10"]
	eleven := reference.WindowsVersionMarkerFonts["11"]
	if !ten.Verified {
		t.Fatal("this test assumes the Windows 10 row is verified")
	}
	if eleven.Verified {
		t.Fatal("this test assumes the Windows 11 row is unverified")
	}
	got := EvaluateWindows(ten.Values)
	if got.Versions["10"] != Present {
		t.Errorf("Windows 10 markers = %v, want present", got.Versions["10"])
	}
	if got.Versions["11"] != Unverified {
		t.Errorf("Windows 11 verdict on a Windows 10 observation = %v, want unverified", got.Versions["11"])
	}
}

func TestCompleteVerifiedMarkerSetIsPresent(t *testing.T) {
	got := EvaluateWindows(reference.WindowsVersionMarkerFonts["10"].Values)
	if got.Versions["10"] != Present {
		t.Errorf("complete Windows 10 marker set = %v, want present", got.Versions["10"])
	}
}

func TestPartialMarkerSetIsInconclusive(t *testing.T) {
	full := reference.WindowsVersionMarkerFonts["10"].Values
	if len(full) < 2 {
		t.Skip("Windows 10 row too small to have a partial case")
	}
	got := EvaluateWindows(full[:1])
	if got.Versions["10"] != Inconclusive {
		t.Errorf("partial Windows 10 marker set = %v, want inconclusive", got.Versions["10"])
	}
}

func TestNoMarkersIsAbsent(t *testing.T) {
	got := EvaluateWindows([]string{"Helvetica Neue", "Ubuntu"})
	if got.Versions["10"] != Absent {
		t.Errorf("no Windows 10 markers = %v, want absent", got.Versions["10"])
	}
	if got.BaseFonts != Absent {
		t.Errorf("no base fonts = %v, want absent", got.BaseFonts)
	}
	if got.Versions["11"] != Unverified {
		t.Errorf("Windows 11 verdict with no markers observed = %v, want unverified", got.Versions["11"])
	}
}

func TestNonProbativeFamiliesAreSkipped(t *testing.T) {
	markers := reference.WindowsNonProbativeMarkers.Values
	if len(markers) == 0 {
		t.Fatal("the non-probative table has no values")
	}
	got := EvaluateWindows(markers)
	for _, s := range got.Skipped {
		found := false
		for _, np := range markers {
			if s == np {
				found = true
			}
		}
		if !found {
			t.Errorf("reported %q as skipped but it is not non-probative", s)
		}
	}
	if len(got.Skipped) != len(markers) {
		t.Errorf("skipped %d families, want %d", len(got.Skipped), len(markers))
	}
	for ver, v := range got.Versions {
		if v != Inconclusive {
			t.Errorf("Windows %s from non-probative families alone = %v, want inconclusive", ver, v)
		}
	}
}

func TestVersionVerdictMeansThatReleaseOrLater(t *testing.T) {
	var all []string
	all = append(all, reference.WindowsVersionMarkerFonts["10"].Values...)
	all = append(all, reference.WindowsVersionMarkerFonts["11"].Values...)
	got := EvaluateWindows(all)
	if got.Versions["10"] != Present {
		t.Errorf("Windows 10 markers alongside other observations = %v, want present", got.Versions["10"])
	}
}

func TestAtLeastReadsTheReleaseVerdict(t *testing.T) {
	got := EvaluateWindows(reference.WindowsVersionMarkerFonts["10"].Values)
	if got.AtLeast("10") != Present {
		t.Errorf("AtLeast(10) = %v, want present", got.AtLeast("10"))
	}
	if _, present := got.Versions["99"]; present {
		t.Error("a release this project holds no table for was given a row of its own")
	}
}

func TestVerdictString(t *testing.T) {
	for v, want := range map[Verdict]string{
		Inconclusive: "inconclusive",
		Present:      "present",
		Absent:       "absent",
		Unverified:   "unverified",
	} {
		if got := v.String(); got != want {
			t.Errorf("Verdict(%d).String() = %q, want %q", int(v), got, want)
		}
	}
}

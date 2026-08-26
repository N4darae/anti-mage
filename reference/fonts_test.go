package reference

import (
	"strings"
	"testing"
)

func TestFeatureOnDemandFontsStayOut(t *testing.T) {
	fod := []string{
		"Aldhabi", "Andalus", "Arabic Typesetting", "Sakkal Majalla",
		"Nyala", "Shruti", "Raavi", "Latha", "Gautami", "Kalinga",
		"Kartika", "Tunga", "Vrinda", "Shonar Bangla", "DokChampa",
		"Euphemia", "Plantagenet Cherokee", "Estrangelo Edessa",
		"Iskoola Pota", "Leelawadee", "Mangal", "Meiryo",
	}
	for _, f := range fod {
		for ver, tbl := range WindowsVersionMarkerFonts {
			for _, got := range tbl.Values {
				if got == f {
					t.Errorf("%q is a Feature On Demand font but appears as a Windows %s marker", f, ver)
				}
			}
		}
		for _, got := range WindowsBaseFonts.Values {
			if got == f {
				t.Errorf("%q is a Feature On Demand font but appears in WindowsBaseFonts", f)
			}
		}
	}
}

func TestCascadiaIsNotAVersionMarker(t *testing.T) {
	for ver, tbl := range WindowsVersionMarkerFonts {
		for _, f := range tbl.Values {
			if strings.HasPrefix(f, "Cascadia") {
				t.Errorf("%q appears as a Windows %s marker but ships independently of the OS", f, ver)
			}
		}
	}
	if len(WindowsNonProbativeMarkers.Values) == 0 {
		t.Error("the non-probative table has no values")
	}
}

func TestWindowsMarkersMatchVendorPages(t *testing.T) {
	want := map[string][]string{
		"10": {"Bahnschrift", "HoloLens MDL2 Assets", "Ink Free", "Segoe MDL2 Assets", "Segoe UI Historic"},
		"11": {"Segoe Fluent Icons", "Segoe UI Variable"},
	}
	for ver, exp := range want {
		tbl, ok := WindowsVersionMarkerFonts[ver]
		if !ok {
			t.Errorf("Windows %s row is missing", ver)
			continue
		}
		if strings.Join(tbl.Values, "|") != strings.Join(exp, "|") {
			t.Errorf("Windows %s markers drifted from the vendor page: got %v, want %v", ver, tbl.Values, exp)
		}
	}
	for _, f := range WindowsVersionMarkerFonts["11"].Values {
		if f == "Segoe MDL2 Assets" || f == "Segoe UI Historic" {
			t.Errorf("%q was added in Windows 10 and ships in both; it is not an 11 marker", f)
		}
	}
	for _, f := range WindowsVersionMarkerFonts["10"].Values {
		if f == "Segoe UI Variable" || f == "Segoe Fluent Icons" {
			t.Errorf("%q is Windows 11 only and must not appear as a 10 marker", f)
		}
	}
}

func TestFontMeasurementBases(t *testing.T) {
	generic := map[string]bool{
		"serif": true, "sans-serif": true, "cursive": true, "fantasy": true,
		"monospace": true, "system-ui": true, "math": true, "ui-serif": true,
		"ui-sans-serif": true, "ui-monospace": true, "ui-rounded": true,
	}
	if len(FontMeasurementBases.Values) < 2 {
		t.Fatalf("a width comparison needs more than one baseline, got %v", FontMeasurementBases.Values)
	}
	seen := map[string]bool{}
	for _, b := range FontMeasurementBases.Values {
		if !generic[b] {
			t.Errorf("baseline %q is not a CSS generic family, so a system could install a font by that name and the comparison would measure it", b)
		}
		if seen[b] {
			t.Errorf("baseline %q is listed twice, which adds no comparison", b)
		}
		seen[b] = true
	}
}

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
	if len(FontMeasurementBases.Values) != 3 {
		t.Fatalf("expected the three CSS generic families, got %v", FontMeasurementBases.Values)
	}
}

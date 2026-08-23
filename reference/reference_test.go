package reference

import (
	"testing"
	"time"
)

func allTables() map[string]Table {
	forms := NativeToStringForms("x")
	tables := map[string]Table{
		"WindowsNonProbativeMarkers":      WindowsNonProbativeMarkers,
		"WindowsBaseFonts":                WindowsBaseFonts,
		"FontMeasurementBases":            FontMeasurementBases,
		"CSSSystemFontKeywords":           CSSSystemFontKeywords,
		"TrustedErrorNames":               TrustedErrorNames,
		"TrustedDOMExceptionNames":        TrustedDOMExceptionNames,
		"NativeFunctionOwnKeys":           NativeFunctionOwnKeys,
		"BraveNativeToString":             BraveNativeToString,
		"NativeToStringForms.Measured":    forms.Measured,
		"NativeToStringForms.OtherEngine": forms.OtherEngine,
	}
	for ver, t := range WindowsVersionMarkerFonts {
		tables["WindowsVersionMarkerFonts["+ver+"]"] = t
	}
	return tables
}

func TestEveryTableHasProvenance(t *testing.T) {
	for name, tbl := range allTables() {
		if tbl.Source.Origin == "" {
			t.Errorf("%s: Source.Origin is empty", name)
		}
		if _, err := time.Parse("2006-01-02", tbl.Source.Checked); err != nil {
			t.Errorf("%s: Source.Checked = %q, want YYYY-MM-DD: %v", name, tbl.Source.Checked, err)
		}
	}
}

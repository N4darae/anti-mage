package reference

var WindowsVersionMarkerFonts = map[string]Table{
	"10": {
		Values:   []string{"Bahnschrift", "HoloLens MDL2 Assets", "Ink Free", "Segoe MDL2 Assets", "Segoe UI Historic"},
		Source:   Source{Origin: `learn.microsoft.com/en-us/typography/fonts/windows_10_font_list, entries marked "* Added in Windows 10"`, Checked: "2026-08-25"},
		Verified: true,
	},
	"11": {
		Values:   []string{"Segoe Fluent Icons", "Segoe UI Variable"},
		Source:   Source{Origin: `learn.microsoft.com/en-us/typography/fonts/windows_11_font_list, entries marked "* Added in Windows 11"`, Checked: "2026-08-25"},
		Verified: false,
	},
}

var WindowsNonProbativeMarkers = Table{
	Values:   []string{"Cascadia Code", "Cascadia Mono"},
	Source:   Source{Origin: "learn.microsoft.com/en-us/typography/fonts/windows_11_font_list; github.com/microsoft/cascadia-code", Checked: "2026-08-25"},
	Verified: true,
}

var WindowsBaseFonts = Table{
	Values: []string{
		"Cambria Math", "Gadugi", "Javanese Text", "Leelawadee UI", "Lucida Console",
		"Marlett", "Myanmar Text", "Nirmala UI", "Segoe UI Emoji", "Segoe UI Symbol", "Sylfaen",
	},
	Source:   Source{Origin: "learn.microsoft.com/en-us/typography/fonts/windows_10_font_list, unmarked entries", Checked: "2026-08-25"},
	Verified: true,
}

var FontMeasurementBases = Table{
	Values:   []string{"monospace", "sans-serif", "serif"},
	Source:   Source{Origin: "CSS Fonts Module Level 4, generic-family value (w3.org/TR/css-fonts-4/#generic-family-value)", Checked: "2026-08-25"},
	Verified: true,
}

const FontMeasurementString = "mmmmmmmmmmlliWQ@0O"

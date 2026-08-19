package reference

// WindowsVersionMarkerFonts maps a Windows release to fonts introduced in
// that release: present from that release onward, absent before it.
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

// WindowsNonProbativeMarkers are families the vendor documents as introduced
// in a release but which also ship independently of the OS.
var WindowsNonProbativeMarkers = Table{
	Values:   []string{"Cascadia Code", "Cascadia Mono"},
	Source:   Source{Origin: "learn.microsoft.com/en-us/typography/fonts/windows_11_font_list; github.com/microsoft/cascadia-code", Checked: "2026-08-25"},
	Verified: true,
}

// WindowsBaseFonts are families present in the Windows 10 base image with no
// "added in" annotation, meaning they predate Windows 10.
var WindowsBaseFonts = Table{
	Values: []string{
		"Cambria Math", "Gadugi", "Javanese Text", "Leelawadee UI", "Lucida Console",
		"Marlett", "Myanmar Text", "Nirmala UI", "Segoe UI Emoji", "Segoe UI Symbol", "Sylfaen",
	},
	Source:   Source{Origin: "learn.microsoft.com/en-us/typography/fonts/windows_10_font_list, unmarked entries", Checked: "2026-08-25"},
	Verified: true,
}

// FontMeasurementBases are the fallback families a candidate font's measured
// metrics are compared against to decide whether it resolved.
var FontMeasurementBases = Table{
	Values:   []string{"monospace", "sans-serif", "serif"},
	Source:   Source{Origin: "CSS Fonts Module Level 4, generic-family value (w3.org/TR/css-fonts-4/#generic-family-value)", Checked: "2026-08-25"},
	Verified: true,
}

// FontMeasurementString is the probe string measured against
// FontMeasurementBases: wide glyphs amplify per-family advance-width
// differences, narrow and irregular glyphs catch families differing in the
// tail. It measures advance width, not glyph coverage.
const FontMeasurementString = "mmmmmmmmmmlliWQ@0O"

// FontMeasurementStringSource is FontMeasurementString's provenance.
var FontMeasurementStringSource = Source{Origin: "probe input of this project's own design; glyph composition verifiable by inspection", Checked: "2026-08-25"}

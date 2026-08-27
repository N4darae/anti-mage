// Package scan turns one collector payload into a report.

package scan

import (
	_ "time/tzdata"
)

type Determination string

const (
	Consistent Determination = "consistent"

	Contradiction Determination = "contradiction"

	Inconclusive Determination = "inconclusive"

	Unverified Determination = "unverified"

	Instrumented Determination = "instrumented"
)

type Row struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Note  string `json:"note"`
}

type Section struct {
	ID            string        `json:"id"`
	Title         string        `json:"title"`
	Determination Determination `json:"determination"`
	Rows          []Row         `json:"rows"`
}

type Summary struct {
	Band            string `json:"band"`
	HumanConfidence int    `json:"humanConfidence"`
	BotLikeness     int    `json:"botLikeness"`
	Headline        string `json:"headline"`
}

type Report struct {
	V        int       `json:"v"`
	Sections []Section `json:"sections"`
	Summary  Summary   `json:"summary"`
}

type sectionFunc func(Request, Inputs, claim) Section

var order = []struct {
	id    string
	title string
	build sectionFunc
}{
	{"platform", "Platform claim", sectionPlatform},
	{"fonts", "Installed fonts against the claimed platform", sectionFonts},
	{"scopes", "Agreement across execution scopes", sectionScopes},
	{"natives", "Native accessor integrity", sectionNatives},
	{"numerics", "Numeric built-in behaviour", sectionMath},
	{"throws", "Exception types against the specification", sectionThrows},
	{"geometry", "Screen geometry against CSS", sectionGeometry},
	{"viewport", "Viewport against the screen it claims", sectionViewport},
	{"rects", "Layout and text metric identities", sectionRects},
	{"mediapaths", "Agreement between the CSS and script paths", sectionMediaPaths},
	{"time", "Time zone against measured offsets", sectionTime},
	{"audio", "Audio buffer coherence", sectionAudioBuf},
	{"automation", "Automation residue", sectionAutomation},
	{"permissions", "Permission state coherence", sectionPermissions},
	{"capabilities", "Reported media capabilities", sectionCapabilities},
	{"hwdecode", "Hardware decoders against the device named", sectionHWDecode},
	{"webgpu", "Both graphics interfaces against one device", sectionWebGPU},
	{"claims", "What this browser reported it could not do", sectionClaims},
}

func Analyze(r Request, in Inputs) Report {
	return AnalyzeWith(r, in, nil)
}

func AnalyzeWith(r Request, in Inputs, supplied []Section) Report {
	if r.Probes == nil {
		r.Probes = map[string]Probe{}
	}
	c := readClaim(r)

	sections := make([]Section, 0, len(order)+len(supplied))
	for _, s := range order {
		sec := s.build(r, in, c)

		sec.ID, sec.Title = s.id, s.title
		sections = append(sections, normalise(sec))
	}
	for _, sec := range supplied {
		sections = append(sections, normalise(sec))
	}
	return Report{V: 1, Sections: sections, Summary: summarise(sections)}
}

func normalise(sec Section) Section {
	if !validDetermination(sec.Determination) {
		sec.Determination = Inconclusive
	}
	if sec.Rows == nil {
		sec.Rows = []Row{}
	}
	return sec
}

func validDetermination(d Determination) bool {
	switch d {
	case Consistent, Contradiction, Inconclusive, Unverified, Instrumented:
		return true
	}
	return false
}

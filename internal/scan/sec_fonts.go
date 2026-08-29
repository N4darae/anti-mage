package scan

import (
	"strconv"

	"github.com/N4darae/anti-mage/osfont"
	"github.com/N4darae/anti-mage/reference"
)

type fontTier struct {
	label string
	table reference.Table
}

func windowsTiers() []fontTier {
	return []fontTier{
		{"families that predate Windows 10", reference.WindowsBaseFonts},
		{"families added in Windows 10", reference.WindowsVersionMarkerFonts["10"]},
		{"families added in Windows 11", reference.WindowsVersionMarkerFonts["11"]},
	}
}

func sectionFonts(r Request, in Inputs, c claim) Section {
	s := Section{Determination: Inconclusive}

	raw, ok := r.value("font.resolved")
	if !ok {
		s.Rows = append(s.Rows, Row{Label: "families resolved", Value: "not collected", Note: anomalyNote})
		return s
	}
	resolved, ok := nameSet(raw, "ascii", "resolved", "present", "width")
	if !ok {
		s.Rows = append(s.Rows, Row{Label: "families resolved", Value: "not readable", Note: anomalyNote})
		return s
	}

	controlNames, controlsResolved, haveControls := readFontControls(r)
	issuedSeen := 0
	for _, n := range in.FontControls {
		if contains(controlNames, n) || contains(controlsResolved, n) {
			issuedSeen++
		}
	}
	if haveControls {
		s.Rows = append(s.Rows, Row{
			Label: "invented control families that resolved",
			Value: strconv.Itoa(len(controlsResolved)) + " of " + strconv.Itoa(len(controlNames)),
			Note:  anomalyNote,
		})
	}

	dropped := coverageFailures(r, resolved)

	res := osfont.EvaluateWindows(resolved)
	tiers := windowsTiers()
	counts := make([]int, len(tiers))
	for i, t := range tiers {
		for _, f := range t.table.Values {
			if contains(resolved, f) {
				counts[i]++
			}
		}
	}

	s.Rows = append(s.Rows, Row{
		Label: "families resolved by advance width",
		Value: strconv.Itoa(len(resolved)),
		Note:  anomalyNote,
	})
	if len(dropped) > 0 {
		s.Rows = append(s.Rows, Row{
			Label: "resolved but own script did not render",
			Value: joinLimit(dropped, 6),
			Note:  anomalyNote,
		})
	}
	for i, t := range tiers {
		v := Inconclusive
		if !t.table.Verified {
			v = Unverified
		}
		note := t.table.Source.Origin
		if v == Unverified {
			note = "this project has not observed these values on a system of the configuration described, so they carry no weight; " + note
		}
		s.Rows = append(s.Rows, Row{
			Label: t.label,
			Value: strconv.Itoa(counts[i]) + " of " + strconv.Itoa(len(t.table.Values)),
			Note:  anomalyNote,
		})
	}
	if len(res.Skipped) > 0 {
		s.Rows = append(s.Rows, Row{
			Label: "resolved but not read as evidence",
			Value: joinLimit(res.Skipped, 6),
			Note:  anomalyNote,
		})
	}
	s.Rows = append(s.Rows, Row{
		Label: "operating system claimed",
		Value: c.Family.String(),
		Note:  anomalyNote,
	})

	if !haveControls {
		s.Determination = Unverified
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "the width probe carried no invented-name controls",
			Note:  anomalyNote,
		})
		return s
	}
	if len(controlsResolved) > 0 {
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "the width probe resolved a family that does not exist",
			Note:  anomalyNote,
		})
		return s
	}
	if !c.Agreed || c.Family != osWindows {
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "no comparison was made",
			Note:  anomalyNote,
		})
		return s
	}

	fl := osfont.ReleaseFloor(resolved)

	anyVerifiedPresent, anyVerified := false, false
	for i, t := range tiers {
		if !t.table.Verified {
			continue
		}
		anyVerified = true
		if counts[i] > 0 {
			anyVerifiedPresent = true
		}
	}
	if anyVerified && !anyVerifiedPresent {
		s.Determination = Contradiction
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "Windows is claimed and no family the vendor publishes for Windows resolved",
			Note:  anomalyNote,
		})
		return s
	}

	if len(fl.AboveGap) > 0 {
		s.Rows = append(s.Rows, Row{
			Label: "reported above a gap",
			Value: joinLimit(fl.AboveGap, 3),
			Note:  anomalyNote,
		})
	}
	s.Determination = Consistent
	if fl.Release == "" {
		s.Rows = append(s.Rows, Row{
			Label: "release floor",
			Value: "no release narrowed",
			Note:  anomalyNote,
		})
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "the resolved families are compatible with a genuine Windows release",
			Note:  anomalyNote,
		})
		return s
	}
	s.Rows = append(s.Rows, Row{
		Label: "release floor",
		Value: "Windows " + fl.Release + " or later",
		Note:  anomalyNote,
	})
	s.Rows = append(s.Rows, Row{
		Label: "conclusion",
		Value: "the resolved families are compatible with a genuine Windows release",
		Note:  anomalyNote,
	})
	return s
}

func readFontControls(r Request) (names, resolvedControls []string, ok bool) {
	raw, have := r.value("font.controls")
	if !have {
		return nil, nil, false
	}
	resolvedControls, _ = nameSet(raw, "ascii", "resolved", "present", "width")
	if m, isMap := raw.(map[string]any); isMap {
		names = keys(m)
	} else {

		names = resolvedControls
	}
	if len(names) == 0 && len(resolvedControls) == 0 {
		return nil, nil, false
	}
	return names, resolvedControls, true
}

func coverageFailures(r Request, resolved []string) []string {
	raw, ok := r.value("font.coverage")
	if !ok {
		return nil
	}
	m, ok := raw.(map[string]any)
	if !ok {
		return nil
	}
	var out []string
	for _, f := range keys(m) {
		if !contains(resolved, f) {
			continue
		}
		covers, have := boolean(m[f], "covers")
		if !have {
			if b, isBool := m[f].(bool); isBool {
				covers, have = b, true
			}
		}
		if have && !covers {
			out = append(out, f)
		}
	}
	return out
}

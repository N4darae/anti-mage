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
		s.Rows = append(s.Rows, Row{Label: "families resolved", Value: "not collected", Note: "the collector did not report a usable font probe"})
		return s
	}
	resolved, ok := nameSet(raw, "ascii", "resolved", "present", "width")
	if !ok {
		s.Rows = append(s.Rows, Row{Label: "families resolved", Value: "not readable", Note: "the reported value was not a family list this engine recognises"})
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
		note := "the control families were not the ones this server issued for this scan, so they were not verified as unpredictable"
		if len(in.FontControls) == 0 {
			note = "this server issued no control families for this scan"
		} else if issuedSeen == len(in.FontControls) {
			note = "every control family this server issued for this scan was reported on"
		} else if issuedSeen > 0 {
			note = strconv.Itoa(issuedSeen) + " of " + strconv.Itoa(len(in.FontControls)) + " control families this server issued were reported on"
		}
		s.Rows = append(s.Rows, Row{
			Label: "invented control families that resolved",
			Value: strconv.Itoa(len(controlsResolved)) + " of " + strconv.Itoa(len(controlNames)),
			Note:  note,
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
		Note:  "measured against " + joinLimit(reference.FontMeasurementBases.Values, 4),
	})
	if len(dropped) > 0 {
		s.Rows = append(s.Rows, Row{
			Label: "resolved but own script did not render",
			Value: joinLimit(dropped, 6),
			Note:  "reported for the reader; the comparison below reads the advance-width result only",
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
			Note:  clip(note, 300),
		})
	}
	if len(res.Skipped) > 0 {
		s.Rows = append(s.Rows, Row{
			Label: "resolved but not read as evidence",
			Value: joinLimit(res.Skipped, 6),
			Note:  "the vendor documents these as introduced in a release but they also ship independently of the operating system",
		})
	}
	s.Rows = append(s.Rows, Row{
		Label: "operating system claimed",
		Value: c.Family.String(),
		Note:  "the release tables this project has verified describe Windows only",
	})

	if !haveControls {
		s.Determination = Unverified
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "the width probe carried no invented-name controls",
			Note:  "without them a probe that resolves every requested family cannot be told from one that resolves the installed ones",
		})
		return s
	}
	if len(controlsResolved) > 0 {
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "the width probe resolved a family that does not exist",
			Note:  "it is not measuring font presence in this environment, so nothing was compared: " + joinLimit(controlsResolved, 4),
		})
		return s
	}
	if !c.Agreed || c.Family != osWindows {
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "no comparison was made",
			Note:  "the release tables this project has verified describe Windows, and this browser does not claim Windows on two agreeing surfaces",
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
			Note:  "not one family from the verified Windows tables resolved by advance width",
		})
		return s
	}

	if len(fl.AboveGap) > 0 {
		s.Rows = append(s.Rows, Row{
			Label: "reported above a gap",
			Value: joinLimit(fl.AboveGap, 3),
			Note:  "a tier resolved while a tier below it did not, so it does not narrow the release; a family the probe cannot measure is the ordinary reason for a gap",
		})
	}
	s.Determination = Consistent
	if fl.Release == "" {
		s.Rows = append(s.Rows, Row{
			Label: "release floor",
			Value: "no release narrowed",
			Note:  "families the vendor publishes for Windows resolved, but not the set that would place a release; nothing is concluded from which ones are missing",
		})
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "the resolved families are compatible with a genuine Windows release",
			Note:  "the shape of what is missing is not read as evidence",
		})
		return s
	}
	s.Rows = append(s.Rows, Row{
		Label: "release floor",
		Value: "Windows " + fl.Release + " or later",
		Note:  "a later release keeps an earlier release's families, so this is a floor and never an exact release",
	})
	s.Rows = append(s.Rows, Row{
		Label: "conclusion",
		Value: "the resolved families are compatible with a genuine Windows release",
		Note:  "the shape of what is missing is not read as evidence",
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

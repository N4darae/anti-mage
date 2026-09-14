package scan

import (
	_ "time/tzdata"

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

type fontTierCounts struct {
	tiers  []fontTier
	counts []int
}

func countWindowsTiers(resolved []string) fontTierCounts {
	tiers := windowsTiers()
	counts := make([]int, len(tiers))
	for i, t := range tiers {
		for _, f := range t.table.Values {
			if contains(resolved, f) {
				counts[i]++
			}
		}
	}
	return fontTierCounts{tiers, counts}
}

func (fc fontTierCounts) verifiedTiersAllEmpty() bool {
	anyVerified, anyPresent := false, false
	for i, t := range fc.tiers {
		if !t.table.Verified {
			continue
		}
		anyVerified = true
		if fc.counts[i] > 0 {
			anyPresent = true
		}
	}
	return anyVerified && !anyPresent
}

func readResolvedFonts(r Request) ([]string, string) {
	raw, ok := r.value("font.resolved")
	if !ok {
		return nil, "not collected"
	}
	resolved, ok := nameSet(raw, "ascii", "resolved", "present", "width")
	if !ok {
		return nil, "not readable"
	}
	return resolved, ""
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

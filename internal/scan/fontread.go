package scan

import (
	_ "time/tzdata"
)

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

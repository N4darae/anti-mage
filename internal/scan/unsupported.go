package scan

import (
	"strconv"
	"strings"
	_ "time/tzdata"
)

const clockMarginMS = 250

func assertsAbsence(reason string) bool {
	low := strings.ToLower(reason)
	for _, incomplete := range []string{
		"timed out", "timeout", "abort", "cancel", "interrupt",
		"detached", "navigated away", "closed before", "took too long",
	} {
		if strings.Contains(low, incomplete) {
			return false
		}
	}
	return true
}

func claimedWaitMS(reason string) (int, bool) {
	low := strings.ToLower(reason)
	best, found := 0, false
	for _, unit := range []struct {
		suffix string
		mult   int
	}{{" ms", 1}, {"ms", 1}, {" s", 1000}, {" seconds", 1000}, {" second", 1000}} {
		idx := 0
		for {
			at := strings.Index(low[idx:], unit.suffix)
			if at < 0 {
				break
			}
			at += idx

			end := at
			start := end
			for start > 0 && low[start-1] >= '0' && low[start-1] <= '9' {
				start--
			}
			if start < end {
				if n, err := strconv.Atoi(low[start:end]); err == nil {
					if v := n * unit.mult; v > best {
						best, found = v, true
					}
				}
			}
			idx = at + len(unit.suffix)
			if idx >= len(low) {
				break
			}
		}
	}
	return best, found
}

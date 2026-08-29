package scan

import (
	"encoding/json"
	"strconv"
	"strings"
)

func sectionClaims(r Request, in Inputs, _ claim) Section {
	s := Section{Determination: Inconclusive}

	type pair struct {
		suppressed string
		label      string
		proof      string
		proofNote  string
		proven     func(Request) bool
	}
	pairs := []pair{
		{
			suppressed: "font.resolved",
			label:      "the width probe",
			proof:      "native.tostring",
			proofNote:  "measureText reported as a working native",
			proven: func(r Request) bool {
				v, ok := r.value("native.tostring")
				if !ok {
					return false
				}
				m, isMap := v.(map[string]any)
				if !isMap {
					return false
				}
				for k := range m {
					if strings.Contains(strings.ToLower(k), "measuretext") {
						return true
					}
				}
				return false
			},
		},
		{
			suppressed: "scope.worker",
			label:      "the worker reading",
			proof:      "scope.availability",
			proofNote:  "same payload reports a worker was created",
			proven: func(r Request) bool {
				v, ok := r.value("scope.availability")
				if !ok {
					return false
				}
				m, isMap := v.(map[string]any)
				if !isMap {
					return false
				}
				for k, raw := range m {
					if !strings.Contains(strings.ToLower(k), "worker") {
						continue
					}
					if b, have := boolean(raw, "created"); have && b {
						return true
					}
					if b, isBool := raw.(bool); isBool && b {
						return true
					}
				}
				return false
			},
		},
		{
			suppressed: "time.offsets",
			label:      "the offset samples",
			proof:      "time.zone",
			proofNote:  "same payload names a zone",
			proven: func(r Request) bool {
				v, ok := r.value("time.zone")
				if !ok {
					return false
				}
				if z, have := str(v, "timeZone"); have && z != "" {
					return true
				}
				z, have := str(v, "zone")
				return have && z != ""
			},
		},
	}

	found := 0
	checked := 0
	for _, p := range pairs {
		if !r.unsupported(p.suppressed) {
			continue
		}
		if reason, have := unsupportedReason(r, p.suppressed); have && !assertsAbsence(reason) {

			if waited, named := claimedWaitMS(reason); named && in.ElapsedMS > 0 && waited > in.ElapsedMS+clockMarginMS {
				checked++
				found++
				s.Rows = append(s.Rows, Row{
					Label: p.label,
					Value: "claimed a wait longer than the scan lasted",
					Note:  anomalyNote,
				})
				continue
			}
			s.Rows = append(s.Rows, Row{
				Label: p.label,
				Value: "reported unsupported",
				Note:  anomalyNote,
			})
			checked++
			continue
		}
		if !p.proven(r) {
			s.Rows = append(s.Rows, Row{
				Label: p.label,
				Value: "reported unsupported",
				Note:  anomalyNote,
			})
			checked++
			continue
		}
		checked++
		found++
		s.Rows = append(s.Rows, Row{
			Label: p.label,
			Value: "reported unsupported while the facility is present",
			Note:  anomalyNote,
		})
	}

	if checked == 0 {
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "nothing to compare",
			Note:  anomalyNote,
		})
		return s
	}
	if in.ElapsedMS > 0 {
		s.Rows = append(s.Rows, Row{
			Label: "scan duration, measured by this server",
			Value: strconv.Itoa(in.ElapsedMS) + " ms",
			Note:  anomalyNote,
		})
	}
	if found == 0 {
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "nothing in this payload either confirms or contradicts what this browser reported it could not do",
			Note:  anomalyNote,
		})
		return s
	}
	s.Determination = Contradiction
	s.Rows = append(s.Rows, Row{
		Label: "conclusion",
		Value: "a probe reported unsupported while this payload shows the facility is present",
		Note:  anomalyNote,
	})
	return s
}

func unsupportedReason(r Request, id string) (string, bool) {
	p, ok := r.Probes[id]
	if !ok || len(p.Value) == 0 {
		return "", false
	}
	var raw any
	if err := json.Unmarshal(p.Value, &raw); err != nil {
		return "", false
	}
	if str, isStr := raw.(string); isStr {
		return str, true
	}
	for _, k := range []string{"reason", "message", "why"} {
		if v, have := str(raw, k); have && v != "" {
			return v, true
		}
	}
	return "", false
}

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

const clockMarginMS = 250

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

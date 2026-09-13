package scan

import (
	"strconv"
	"strings"
	_ "time/tzdata"
)

type suppressionPair struct {
	suppressed string
	label      string
	proof      string
	proofNote  string
	proven     func(Request) bool
}

var suppressionPairs = []suppressionPair{
	{
		suppressed: "font.resolved",
		label:      "the width probe",
		proof:      "native.tostring",
		proofNote:  "measureText reported as a working native",
		proven: func(r Request) bool {
			return namedProbeKey(r, "native.tostring", "measuretext", nil)
		},
	},
	{
		suppressed: "scope.worker",
		label:      "the worker reading",
		proof:      "scope.availability",
		proofNote:  "same payload reports a worker was created",
		proven: func(r Request) bool {
			return namedProbeKey(r, "scope.availability", "worker", func(raw any) bool {
				if b, have := boolean(raw, "created"); have && b {
					return true
				}
				b, isBool := raw.(bool)
				return isBool && b
			})
		},
	},
	{
		suppressed: "time.offsets",
		label:      "the offset samples",
		proof:      "time.zone",
		proofNote:  "same payload names a zone",
		proven: func(r Request) bool {
			v, ok := r.value("time.zone")
			return ok && firstString(v, []string{"timeZone"}, []string{"zone"}) != ""
		},
	},
}

func namedProbeKey(r Request, id, substring string, accept func(any) bool) bool {
	v, ok := r.value(id)
	if !ok {
		return false
	}
	m, isMap := v.(map[string]any)
	if !isMap {
		return false
	}
	for k, raw := range m {
		if !strings.Contains(strings.ToLower(k), substring) {
			continue
		}
		if accept == nil || accept(raw) {
			return true
		}
	}
	return false
}

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

func (p suppressionPair) excused(r Request) bool {
	reason, have := unsupportedReason(r, p.suppressed)
	return have && !assertsAbsence(reason)
}

func (p suppressionPair) overstatesItsWait(r Request, in Inputs) bool {
	if !p.excused(r) || in.ElapsedMS <= 0 {
		return false
	}
	reason, _ := unsupportedReason(r, p.suppressed)
	waited, named := claimedWaitMS(reason)
	return named && waited > in.ElapsedMS+clockMarginMS
}

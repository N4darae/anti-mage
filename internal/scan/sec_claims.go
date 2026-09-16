package scan

import (
	"strconv"
	"strings"
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

	found := 0
	checked := 0
	for _, p := range suppressionPairs {
		if !r.unsupported(p.suppressed) {
			continue
		}
		checked++
		switch {
		case p.overstatesItsWait(r, in):
			found++
			s.add(p.label, "claimed a wait longer than the scan lasted")
		case p.excused(r):
			s.add(p.label, "reported unsupported")
		case !p.proven(r):
			s.add(p.label, "reported unsupported")
		default:
			found++
			s.add(p.label, "reported unsupported while the facility is present")
		}
	}

	if checked == 0 {
		s.conclude("nothing to compare")
		return s
	}
	if in.ElapsedMS > 0 {
		s.add("scan duration, measured by this server", strconv.Itoa(in.ElapsedMS)+" ms")
	}
	if found == 0 {
		s.conclude("nothing in this payload either confirms or contradicts what this browser reported it could not do")
		return s
	}
	s.Determination = Contradiction
	s.conclude("a probe reported unsupported while this payload shows the facility is present")
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

package scan

import (
	"math"
	"strconv"
)

const dprTolerance = 1e-3

func sectionGeometry(r Request, _ Inputs, c claim) Section {
	s := Section{Determination: Inconclusive}

	screen, haveScreen := r.value("geom.screen")
	css, haveCSS := r.value("geom.css")
	if !haveScreen {
		s.Rows = append(s.Rows, Row{Label: "screen numbers", Value: "not collected", Note: anomalyNote})
		return s
	}

	dim := func(label string, path ...string) (float64, bool) {
		v, ok := num(screen, path...)
		if !ok {
			return 0, false
		}
		s.Rows = append(s.Rows, Row{Label: label, Value: itoa(v), Note: anomalyNote})
		return v, true
	}
	w, haveW := dim("screen.width", "width")
	h, haveH := dim("screen.height", "height")
	aw, haveAW := dim("screen.availWidth", "availWidth")
	ah, haveAH := dim("screen.availHeight", "availHeight")

	screenGetters := explainedBy(c, keyScreenWidth, keyScreenHeight, keyScreenAvailWidth, keyScreenAvailHeight)
	dpr := explainedBy(c, keyMatchMedia, keyDevicePixelRatio)

	var t tally
	if haveW && haveAW && w > 0 && aw > 0 {
		bad := 0
		if aw > w {
			bad = 1
			s.Rows = append(s.Rows, Row{Label: "available width", Value: "exceeds the screen width", Note: anomalyNote})
		}
		t.fold(1, bad, screenGetters)
	}
	if haveH && haveAH && h > 0 && ah > 0 {
		bad := 0
		if ah > h {
			bad = 1
			s.Rows = append(s.Rows, Row{Label: "available height", Value: "exceeds the screen height", Note: anomalyNote})
		}
		t.fold(1, bad, screenGetters)
	}

	jsDPR, haveJS := readDPR(screen)
	if haveJS {
		s.Rows = append(s.Rows, Row{Label: "window.devicePixelRatio", Value: strconv.FormatFloat(jsDPR, 'g', -1, 64), Note: anomalyNote})
	}
	if haveCSS {
		cssDPR, haveCSSDPR := readDPR(css)
		if haveCSSDPR {
			s.Rows = append(s.Rows, Row{Label: "ratio recovered from CSS", Value: strconv.FormatFloat(cssDPR, 'g', -1, 64), Note: anomalyNote})
		}
		if haveJS && haveCSSDPR && jsDPR > 0 && cssDPR > 0 {
			bad := 0
			if math.Abs(jsDPR-cssDPR)/math.Max(jsDPR, cssDPR) > dprTolerance {
				bad = 1
				s.Rows = append(s.Rows, Row{
					Label: "device pixel ratio",
					Value: "JavaScript and CSS report different ratios",
					Note:  anomalyNote,
				})
			}
			t.fold(1, bad, dpr)
		}
	} else {
		s.Rows = append(s.Rows, Row{Label: "CSS readings", Value: "not collected", Note: anomalyNote})
	}

	s.Determination = t.determination()
	switch s.Determination {
	case Inconclusive:
		s.Rows = append(s.Rows, Row{Label: "conclusion", Value: "nothing could be compared", Note: anomalyNote})
	case Contradiction:
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: strconv.Itoa(t.unexplained) + " of " + strconv.Itoa(t.applied) + " requirements did not hold",
			Note:  anomalyNote,
		})
	case Instrumented:
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: strconv.Itoa(t.explained) + " of " + strconv.Itoa(t.applied) + " requirements did not hold",
			Note:  anomalyNote,
		})
	default:
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "the JavaScript and CSS readings agree",
			Note:  anomalyNote,
		})
	}
	return s
}

func readDPR(v any) (float64, bool) {
	for _, k := range []string{"devicePixelRatio", "dpr", "dppx", "resolution", "ratio"} {
		if f, ok := num(v, k); ok && f > 0 {
			return f, true
		}
	}
	return 0, false
}

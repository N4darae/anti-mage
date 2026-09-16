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
		s.add("screen numbers", "not collected")
		return s
	}

	dim := func(label string, path ...string) (float64, bool) {
		v, ok := num(screen, path...)
		if !ok {
			return 0, false
		}
		s.add(label, itoa(v))
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
			s.add("available width", "exceeds the screen width")
		}
		t.fold(1, bad, screenGetters)
	}
	if haveH && haveAH && h > 0 && ah > 0 {
		bad := 0
		if ah > h {
			bad = 1
			s.add("available height", "exceeds the screen height")
		}
		t.fold(1, bad, screenGetters)
	}

	jsDPR, haveJS := readDPR(screen)
	if haveJS {
		s.add("window.devicePixelRatio", strconv.FormatFloat(jsDPR, 'g', -1, 64))
	}
	if haveCSS {
		cssDPR, haveCSSDPR := readDPR(css)
		if haveCSSDPR {
			s.add("ratio recovered from CSS", strconv.FormatFloat(cssDPR, 'g', -1, 64))
		}
		if haveJS && haveCSSDPR && jsDPR > 0 && cssDPR > 0 {
			bad := 0
			if math.Abs(jsDPR-cssDPR)/math.Max(jsDPR, cssDPR) > dprTolerance {
				bad = 1
				s.add("device pixel ratio", "JavaScript and CSS report different ratios")
			}
			t.fold(1, bad, dpr)
		}
	} else {
		s.add("CSS readings", "not collected")
	}

	s.settle(t, conclusions{
		inconclusive:  "nothing could be compared",
		contradiction: outOf(t.unexplained, t.applied) + " requirements did not hold",
		instrumented:  outOf(t.explained, t.applied) + " requirements did not hold",
		consistent:    "the JavaScript and CSS readings agree",
	})
	return s
}

func readDPR(v any) (float64, bool) {
	return firstPositive(v, "devicePixelRatio", "dpr", "dppx", "resolution", "ratio")
}

package scan

import (
	_ "time/tzdata"
)

func sectionViewport(r Request, _ Inputs, _ claim) Section {
	s := Section{Determination: Unverified}

	geom, ok := r.value("geom.screen")
	if !ok {
		s.add("screen and viewport", "not collected")
		return s
	}

	screenW, haveScreenW := num(geom, "width")
	screenH, haveScreenH := num(geom, "height")
	innerW, haveInnerW := num(geom, "innerWidth")
	innerH, haveInnerH := num(geom, "innerHeight")

	s.add("screen, as reported", dimension(screenW, haveScreenW, screenH, haveScreenH))
	s.add("viewport, as reported", dimension(innerW, haveInnerW, innerH, haveInnerH))

	if !haveScreenW || !haveScreenH || !haveInnerW || !haveInnerH {
		s.conclude("one of the four numbers was not reported")
		return s
	}
	if screenW <= 0 || screenH <= 0 || innerW <= 0 || innerH <= 0 {
		s.conclude("a reported size was zero or negative")
		return s
	}

	offsetX, haveOffsetX := num(geom, "screenX")
	offsetY, haveOffsetY := num(geom, "screenY")
	if (haveOffsetX && offsetX < 0) || (haveOffsetY && offsetY < 0) {
		s.add("window offset", offset(offsetX, haveOffsetX, offsetY, haveOffsetY))
		s.conclude("the window is not positioned on the screen that was measured")
		return s
	}

	if innerH > screenH || innerW > screenW {
		s.Determination = Contradiction
		s.conclude("the area this page was given is larger than the screen it is said to be displayed on")
		return s
	}

	s.Determination = Consistent
	s.conclude("the viewport fits within the screen this environment claims")
	return s
}

func dimension(w float64, haveW bool, h float64, haveH bool) string {
	if !haveW || !haveH {
		return "not reported"
	}
	return itoa(w) + " by " + itoa(h)
}

func offset(x float64, haveX bool, y float64, haveY bool) string {
	if !haveX && !haveY {
		return "not reported"
	}
	left, top := "not reported", "not reported"
	if haveX {
		left = itoa(x)
	}
	if haveY {
		top = itoa(y)
	}
	return left + " from the left, " + top + " from the top"
}

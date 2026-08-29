package scan

func sectionViewport(r Request, _ Inputs, _ claim) Section {
	s := Section{Determination: Unverified}

	geom, ok := r.value("geom.screen")
	if !ok {
		s.Rows = append(s.Rows, Row{
			Label: "screen and viewport",
			Value: "not collected",
			Note:  anomalyNote,
		})
		return s
	}

	screenW, haveScreenW := num(geom, "width")
	screenH, haveScreenH := num(geom, "height")
	innerW, haveInnerW := num(geom, "innerWidth")
	innerH, haveInnerH := num(geom, "innerHeight")

	s.Rows = append(s.Rows, Row{Label: "screen, as reported", Value: dimension(screenW, haveScreenW, screenH, haveScreenH), Note: anomalyNote})
	s.Rows = append(s.Rows, Row{Label: "viewport, as reported", Value: dimension(innerW, haveInnerW, innerH, haveInnerH), Note: anomalyNote})

	if !haveScreenW || !haveScreenH || !haveInnerW || !haveInnerH {
		s.Rows = append(s.Rows, Row{Label: "conclusion", Value: "one of the four numbers was not reported", Note: anomalyNote})
		return s
	}
	if screenW <= 0 || screenH <= 0 || innerW <= 0 || innerH <= 0 {
		s.Rows = append(s.Rows, Row{Label: "conclusion", Value: "a reported size was zero or negative", Note: anomalyNote})
		return s
	}

	offsetX, haveOffsetX := num(geom, "screenX")
	offsetY, haveOffsetY := num(geom, "screenY")
	if (haveOffsetX && offsetX < 0) || (haveOffsetY && offsetY < 0) {
		s.Rows = append(s.Rows, Row{Label: "window offset", Value: offset(offsetX, haveOffsetX, offsetY, haveOffsetY), Note: anomalyNote})
		s.Rows = append(s.Rows, Row{Label: "conclusion", Value: "the window is not positioned on the screen that was measured", Note: anomalyNote})
		return s
	}

	if innerH > screenH || innerW > screenW {
		s.Determination = Contradiction
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "the area this page was given is larger than the screen it is said to be displayed on",
			Note:  anomalyNote,
		})
		return s
	}

	s.Determination = Consistent
	s.Rows = append(s.Rows, Row{Label: "conclusion", Value: "the viewport fits within the screen this environment claims", Note: anomalyNote})
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

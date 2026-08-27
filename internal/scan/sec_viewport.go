package scan

func sectionViewport(r Request, _ Inputs, _ claim) Section {
	s := Section{Determination: Unverified}

	geom, ok := r.value("geom.screen")
	if !ok {
		s.Rows = append(s.Rows, Row{
			Label: "screen and viewport",
			Value: "not collected",
			Note:  "this reading compares a viewport against the screen it is displayed on; without both it carries no weight either way",
		})
		return s
	}

	screenW, haveScreenW := num(geom, "width")
	screenH, haveScreenH := num(geom, "height")
	innerW, haveInnerW := num(geom, "innerWidth")
	innerH, haveInnerH := num(geom, "innerHeight")

	s.Rows = append(s.Rows, Row{Label: "screen, as reported", Value: dimension(screenW, haveScreenW, screenH, haveScreenH), Note: "the size of the output device this environment says it is displayed on"})
	s.Rows = append(s.Rows, Row{Label: "viewport, as reported", Value: dimension(innerW, haveInnerW, innerH, haveInnerH), Note: "the size of the area this page was given to draw in"})

	if !haveScreenW || !haveScreenH || !haveInnerW || !haveInnerH {
		s.Rows = append(s.Rows, Row{Label: "conclusion", Value: "one of the four numbers was not reported", Note: "nothing was compared, and this reading carries no weight"})
		return s
	}
	if screenW <= 0 || screenH <= 0 || innerW <= 0 || innerH <= 0 {
		s.Rows = append(s.Rows, Row{Label: "conclusion", Value: "a reported size was zero or negative", Note: "a screen or viewport of no size settles nothing"})
		return s
	}

	offsetX, haveOffsetX := num(geom, "screenX")
	offsetY, haveOffsetY := num(geom, "screenY")
	if (haveOffsetX && offsetX < 0) || (haveOffsetY && offsetY < 0) {
		s.Rows = append(s.Rows, Row{Label: "window offset", Value: offset(offsetX, haveOffsetX, offsetY, haveOffsetY), Note: "an offset before the origin places this window on a display other than the one measured, whose size this reading was not given"})
		s.Rows = append(s.Rows, Row{Label: "conclusion", Value: "the window is not positioned on the screen that was measured", Note: "nothing was compared"})
		return s
	}

	if innerH > screenH || innerW > screenW {
		s.Determination = Contradiction
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "the area this page was given is larger than the screen it is said to be displayed on",
			Note:  "both are reported in the same units and scale together, so a viewport cannot exceed its own output device",
		})
		return s
	}

	s.Determination = Consistent
	s.Rows = append(s.Rows, Row{Label: "conclusion", Value: "the viewport fits within the screen this environment claims", Note: ""})
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

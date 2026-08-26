package scan

func sectionPlatform(r Request, _ Inputs, c claim) Section {
	s := Section{Determination: Inconclusive}

	if !r.ran("scope.main") {
		s.Rows = append(s.Rows, Row{
			Label: "main-thread facts",
			Value: "not collected",
			Note:  "the collector did not report this probe",
		})
		return s
	}
	if r.unsupported("scope.main") {
		s.Rows = append(s.Rows, Row{
			Label: "main-thread facts",
			Value: "unsupported",
			Note:  "the collector could not read them in this browser",
		})
		return s
	}

	add := func(label, raw string, f osFamily) {
		if raw == "" {
			s.Rows = append(s.Rows, Row{Label: label, Value: "not exposed", Note: "not read as evidence"})
			return
		}
		note := ""
		if f == osUnknown {
			note = "no operating-system family in this project's table matches this value"
		}
		s.Rows = append(s.Rows, Row{Label: label, Value: clip(raw, 200), Note: note})
	}
	add("navigator.userAgent", c.UserAgent, c.uaFamily)
	add("navigator.platform", c.NavPlatform, c.platFamily)
	add("NavigatorUAData.platform", c.UADataPlat, c.uaDataFamily)

	if c.surfacesKnown < 2 {
		s.Rows = append(s.Rows, Row{
			Label: "operating system named",
			Value: c.Family.String(),
			Note:  "fewer than two surfaces named a recognised family, so nothing was compared",
		})
		return s
	}
	if !c.Agreed {
		s.Determination = Contradiction
		s.Rows = append(s.Rows, Row{
			Label: "operating system named",
			Value: "the surfaces name different operating systems",
			Note: "navigator.userAgent names " + c.uaFamily.String() +
				", navigator.platform names " + c.platFamily.String() +
				", NavigatorUAData names " + c.uaDataFamily.String(),
		})
		return s
	}
	s.Determination = Consistent
	s.Rows = append(s.Rows, Row{
		Label: "operating system named",
		Value: c.Family.String(),
		Note:  "every readable surface names the same operating-system family",
	})
	return s
}

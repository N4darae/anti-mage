package scan

func sectionPlatform(r Request, _ Inputs, c claim) Section {
	s := Section{Determination: Inconclusive}

	if !r.ran("scope.main") {
		s.Rows = append(s.Rows, Row{
			Label: "main-thread facts",
			Value: "not collected",
			Note:  anomalyNote,
		})
		return s
	}
	if r.unsupported("scope.main") {
		s.Rows = append(s.Rows, Row{
			Label: "main-thread facts",
			Value: "unsupported",
			Note:  anomalyNote,
		})
		return s
	}

	add := func(label, raw string, f osFamily) {
		if raw == "" {
			s.Rows = append(s.Rows, Row{Label: label, Value: "not exposed", Note: anomalyNote})
			return
		}
		s.Rows = append(s.Rows, Row{Label: label, Value: clip(raw, 200), Note: anomalyNote})
	}
	add("navigator.userAgent", c.UserAgent, c.uaFamily)
	add("navigator.platform", c.NavPlatform, c.platFamily)
	add("NavigatorUAData.platform", c.UADataPlat, c.uaDataFamily)

	if c.surfacesKnown < 2 {
		s.Rows = append(s.Rows, Row{
			Label: "operating system named",
			Value: c.Family.String(),
			Note:  anomalyNote,
		})
		return s
	}
	if !c.Agreed {
		s.Determination = Contradiction
		s.Rows = append(s.Rows, Row{
			Label: "operating system named",
			Value: "the surfaces name different operating systems",
			Note:  anomalyNote,
		})
		return s
	}
	s.Determination = Consistent
	s.Rows = append(s.Rows, Row{
		Label: "operating system named",
		Value: c.Family.String(),
		Note:  anomalyNote,
	})
	return s
}

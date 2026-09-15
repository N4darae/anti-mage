package scan

import (
	_ "time/tzdata"
)

func sectionPlatform(r Request, _ Inputs, c claim) Section {
	s := Section{Determination: Inconclusive}

	if !r.ran("scope.main") {
		s.add("main-thread facts", "not collected")
		return s
	}
	if r.unsupported("scope.main") {
		s.add("main-thread facts", "unsupported")
		return s
	}

	add := func(label, raw string, f osFamily) {
		if raw == "" {
			s.add(label, "not exposed")
			return
		}
		s.add(label, clip(raw, 200))
	}
	add("navigator.userAgent", c.UserAgent, c.uaFamily)
	add("navigator.platform", c.NavPlatform, c.platFamily)
	add("NavigatorUAData.platform", c.UADataPlat, c.uaDataFamily)

	if c.surfacesKnown < 2 {
		s.add("operating system named", c.Family.String())
		return s
	}
	if !c.Agreed {
		s.Determination = Contradiction
		s.add("operating system named", "the surfaces name different operating systems")
		return s
	}
	s.Determination = Consistent
	s.add("operating system named", c.Family.String())
	return s
}

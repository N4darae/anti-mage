package scan

import "strconv"

func sectionAutomation(r Request, _ Inputs, _ claim) Section {
	s := Section{Determination: Inconclusive}

	raw, ok := r.value("auto.residue")
	if !ok {
		s.Rows = append(s.Rows, Row{Label: "automation residue", Value: "not collected", Note: "the collector did not report it"})
		return s
	}

	flag, haveFlag := boolean(raw, "webdriver")
	names, haveNames := []string(nil), false
	for _, k := range []string{"driverProperties", "driverNames", "residue", "found", "properties"} {
		if x, have := field(raw, k); have {
			if l, ok := nameSet(x, "found", "present"); ok {
				names, haveNames = l, true
				break
			}
		}
	}

	if haveFlag {
		v := "false"
		note := "the browser does not declare itself under remote control"
		if flag {
			v = "true"
			note = "the browser declares itself under remote control, which is what this member is defined to report"
		}
		s.Rows = append(s.Rows, Row{Label: "navigator.webdriver", Value: v, Note: note})
	} else {
		s.Rows = append(s.Rows, Row{Label: "navigator.webdriver", Value: "not reported", Note: "not read as evidence"})
	}
	if haveNames {
		s.Rows = append(s.Rows, Row{
			Label: "global names defined only by a remote-control tool",
			Value: strconv.Itoa(len(names)),
			Note:  joinLimit(names, 6),
		})
	}

	if !haveFlag && !haveNames {
		s.Rows = append(s.Rows, Row{Label: "conclusion", Value: "nothing to read", Note: "no observation was reported"})
		return s
	}
	if (haveFlag && flag) || len(names) > 0 {
		s.Determination = Instrumented
		what := "the browser declares itself under remote control"
		if len(names) > 0 {
			what = "a global name that only a remote-control tool defines is present"
			if haveFlag && flag {
				what = "the browser declares itself under remote control, and a global name that only a remote-control tool defines is present"
			}
		}
		s.Rows = append(s.Rows, Row{Label: "conclusion", Value: "this environment is being driven", Note: what})
		return s
	}
	s.Determination = Consistent
	s.Rows = append(s.Rows, Row{
		Label: "conclusion",
		Value: "no declaration of remote control and no driver-defined global name",
		Note:  "this is the ordinary answer and is not by itself evidence of anything",
	})
	return s
}

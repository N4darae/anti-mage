package scan

import "strconv"

func stackTraceLimitStripped(raw any) (stripped, known bool) {
	kind, haveKind := str(raw, "captureStackTraceType")
	if !haveKind || kind != "function" {
		return false, false
	}
	_, isNumber := num(raw, "stackTraceLimit")
	return !isNumber, true
}

func sectionAutomation(r Request, _ Inputs, _ claim) Section {
	s := Section{Determination: Inconclusive}

	raw, ok := r.value("auto.residue")
	if !ok {
		s.Rows = append(s.Rows, Row{Label: "automation residue", Value: "not collected", Note: anomalyNote})
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
		if flag {
			v = "true"
		}
		s.Rows = append(s.Rows, Row{Label: "navigator.webdriver", Value: v, Note: anomalyNote})
	} else {
		s.Rows = append(s.Rows, Row{Label: "navigator.webdriver", Value: "not reported", Note: anomalyNote})
	}
	if haveNames {
		s.Rows = append(s.Rows, Row{
			Label: "global names defined only by a remote-control tool",
			Value: strconv.Itoa(len(names)),
			Note:  anomalyNote,
		})
	}

	stripped, haveLimit := stackTraceLimitStripped(raw)
	if haveLimit {
		v := "present"
		if stripped {
			v = "removed"
		}
		s.Rows = append(s.Rows, Row{Label: "Error.stackTraceLimit", Value: v, Note: anomalyNote})
	}

	if !haveFlag && !haveNames && !haveLimit {
		s.Rows = append(s.Rows, Row{Label: "conclusion", Value: "nothing to read", Note: anomalyNote})
		return s
	}
	if (haveFlag && flag) || len(names) > 0 {
		s.Determination, s.weight = Instrumented, weightOnlyDeliberate
		s.Rows = append(s.Rows, Row{Label: "conclusion", Value: "this environment is being driven", Note: anomalyNote})
		return s
	}
	if stripped {
		s.Determination, s.weight = Instrumented, weightOnlyDeliberate
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "a member of the engine's own stack-trace interface has been removed",
			Note:  anomalyNote,
		})
		return s
	}
	s.Determination = Consistent
	s.Rows = append(s.Rows, Row{
		Label: "conclusion",
		Value: "no declaration of remote control and no driver-defined global name",
		Note:  anomalyNote,
	})
	return s
}

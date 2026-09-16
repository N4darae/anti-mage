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
		s.add("automation residue", "not collected")
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
		s.add("navigator.webdriver", boolWord(flag))
	} else {
		s.add("navigator.webdriver", "not reported")
	}
	if haveNames {
		s.add("global names defined only by a remote-control tool", strconv.Itoa(len(names)))
	}

	stripped, haveLimit := stackTraceLimitStripped(raw)
	if haveLimit {
		value := "present"
		if stripped {
			value = "removed"
		}
		s.add("Error.stackTraceLimit", value)
	}

	if !haveFlag && !haveNames && !haveLimit {
		s.conclude("nothing to read")
		return s
	}
	if (haveFlag && flag) || len(names) > 0 {
		s.Determination, s.weight = Instrumented, weightOnlyDeliberate
		s.conclude("this environment is being driven")
		return s
	}
	if stripped {
		s.Determination, s.weight = Instrumented, weightOnlyDeliberate
		s.conclude("a member of the engine's own stack-trace interface has been removed")
		return s
	}
	s.Determination = Consistent
	s.conclude("no declaration of remote control and no driver-defined global name")
	return s
}

package scan

import (
	"sort"
	"strconv"
	_ "time/tzdata"

	"github.com/N4darae/anti-mage/reference"
)

func sectionThrows(r Request, _ Inputs, _ claim) Section {
	s := Section{Determination: Inconclusive}

	if !reference.TrustedErrorNames.Verified || !reference.TrustedDOMExceptionNames.Verified {
		s.Determination = Unverified
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "the error-name registries this section depends on are not themselves verified",
			Note:  anomalyNote,
		})
		return s
	}

	type caseFailure struct {
		c       throwsMandatedCase
		reasons []string
	}
	var failures []*caseFailure
	byCase := map[string]*caseFailure{}
	addViolation := func(c throwsMandatedCase, what string) {
		f := byCase[c.id]
		if f == nil {
			f = &caseFailure{c: c}
			byCase[c.id] = f
			failures = append(failures, f)
		}
		f.reasons = append(f.reasons, what)
	}

	applied := 0
	notApplicable := 0
	allNames := map[string]bool{}

	v, haveMandated := r.value("throw.mandated")
	var mandatedMap map[string]any
	if haveMandated {
		mandatedMap, haveMandated = v.(map[string]any)
	}

	if haveMandated {
		for _, c := range throwsMandatedCases {
			raw, present := mandatedMap[c.id]
			if !present {
				notApplicable++
				continue
			}
			sub, isMap := raw.(map[string]any)
			if !isMap {
				notApplicable++
				continue
			}
			available, haveAvailable := boolean(sub, "available")
			if !haveAvailable || !available {

				notApplicable++
				continue
			}
			threw, haveThrew := boolean(sub, "threw")
			if !haveThrew {
				notApplicable++
				continue
			}
			applied++
			if !threw {
				addViolation(c, "completed without throwing, though "+c.mandate+" requires it to throw here")
				continue
			}

			name, haveName := str(sub, "name")
			if n := clip(name, 80); haveName && n != "" {
				allNames[n] = true
			}
			ctorName, haveCtor := str(sub, "ctor")
			instanceOf, haveInstanceOf := boolean(sub, "instanceOf")
			ctorAvailable, haveCtorAvailable := boolean(sub, "ctorGlobalAvailable")
			tag, haveTag := str(sub, "toStringTag")

			if haveCtor && ctorName != c.ctor {
				addViolation(c, "constructor "+clip(ctorName, 60)+", mandate is "+c.ctor)
			}
			if haveInstanceOf && haveCtorAvailable && ctorAvailable && !instanceOf {
				addViolation(c, "thrown value is not instanceof "+c.ctor)
			}
			if haveTag {
				wantTag := "[object Error]"
				if c.domException {
					wantTag = "[object DOMException]"
				}
				if tag != wantTag {
					addViolation(c, "toString tag "+clip(tag, 60)+" instead of "+wantTag)
				}
			}

			if c.domException && haveName && name != c.name {
				addViolation(c, "name "+clip(name, 60)+", mandate is "+c.name)
			}

			if a2, haveA2 := object(sub, "attempt2"); haveA2 {
				threw2, haveThrew2 := boolean(a2, "threw")
				switch {
				case haveThrew2 && !threw2:
					addViolation(c, "threw on first attempt, not on identical second")
				case haveThrew2 && threw2:
					if n2, ok2 := str(a2, "name"); ok2 && haveName && n2 != name {
						addViolation(c, "different exception name on second identical attempt")
					}
					if c2, ok2 := str(a2, "ctor"); ok2 && haveCtor && c2 != ctorName {
						addViolation(c, "different constructor on second identical attempt")
					}
					if t2, ok2 := str(a2, "toStringTag"); ok2 && haveTag && t2 != tag {
						addViolation(c, "toString tag disagreed between two identical attempts")
					}
					if i2, ok2 := boolean(a2, "instanceOf"); ok2 && haveInstanceOf && i2 != instanceOf {
						addViolation(c, "instanceof disagreed between two identical attempts")
					}
					if n2, ok2 := str(a2, "name"); ok2 {
						if nn := clip(n2, 80); nn != "" {
							allNames[nn] = true
						}
					}
				}
			}
		}
	}

	if names, ok := r.value("throw.names"); ok {
		if m, isMap := names.(map[string]any); isMap {
			for _, id := range keys(m) {
				threw, _ := boolean(m[id], "threw")
				if !threw {
					continue
				}
				if n, ok := str(m[id], "name"); ok {
					if nn := clip(n, 80); nn != "" {
						allNames[nn] = true
					}
				}
			}
		}
	}

	var outsideRegistry []string
	for n := range allNames {
		if !contains(reference.TrustedErrorNames.Values, n) && !contains(reference.TrustedDOMExceptionNames.Values, n) {
			outsideRegistry = append(outsideRegistry, n)
		}
	}
	sort.Strings(outsideRegistry)

	s.Rows = append(s.Rows, Row{
		Label: "mandated cases evaluated",
		Value: strconv.Itoa(applied) + " case(s) evaluated",
		Note:  anomalyNote,
	})
	if len(allNames) > 0 {
		s.Rows = append(s.Rows, Row{
			Label: "error names observed, against the two verified registries",
			Value: strconv.Itoa(len(allNames)) + " distinct name(s)",
			Note:  anomalyNote,
		})
	}

	if applied == 0 {
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "no mandated case could be evaluated",
			Note:  anomalyNote,
		})
		return s
	}

	if len(failures) == 0 {
		s.Determination = Consistent
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "every mandated case produced exactly the exception type its specification requires",
			Note:  anomalyNote,
		})
		return s
	}

	shown := failures
	if len(shown) > 8 {
		shown = shown[:8]
	}
	for _, f := range shown {
		s.Rows = append(s.Rows, Row{
			Label: clip(f.c.label, 120),
			Value: "did not meet its mandated exception type",
			Note:  anomalyNote,
		})

		s.Rows = append(s.Rows, Row{
			Label: "specification basis",
			Value: clip(f.c.label, 120),
			Note:  anomalyNote,
		})
	}
	s.Determination = Contradiction
	s.Rows = append(s.Rows, Row{
		Label: "conclusion",
		Value: strconv.Itoa(len(failures)) + " of " + strconv.Itoa(applied) + " mandated case(s) disagreed with the type their specification fixes",
		Note:  anomalyNote,
	})
	return s
}

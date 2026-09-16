package scan

import (
	_ "time/tzdata"

	"github.com/N4darae/anti-mage/reference"
)

type throwsFailure struct {
	c       throwsMandatedCase
	reasons []string
}

type throwsReading struct {
	applied  int
	failures []*throwsFailure
	byCase   map[string]*throwsFailure
	names    map[string]bool
}

func (tr *throwsReading) note(c throwsMandatedCase, what string) {
	f := tr.byCase[c.id]
	if f == nil {
		f = &throwsFailure{c: c}
		tr.byCase[c.id] = f
		tr.failures = append(tr.failures, f)
	}
	f.reasons = append(f.reasons, what)
}

func (tr *throwsReading) sawName(n string) {
	if clipped := clip(n, 80); clipped != "" {
		tr.names[clipped] = true
	}
}

func readThrows(r Request) throwsReading {
	tr := throwsReading{byCase: map[string]*throwsFailure{}, names: map[string]bool{}}

	if v, ok := r.value("throw.mandated"); ok {
		if mandated, isMap := v.(map[string]any); isMap {
			for _, c := range throwsMandatedCases {
				sub, isMap := mandated[c.id].(map[string]any)
				if !isMap {
					continue
				}
				tr.evaluate(c, sub)
			}
		}
	}

	if v, ok := r.value("throw.names"); ok {
		if m, isMap := v.(map[string]any); isMap {
			for _, id := range keys(m) {
				if threw, _ := boolean(m[id], "threw"); !threw {
					continue
				}
				if n, ok := str(m[id], "name"); ok {
					tr.sawName(n)
				}
			}
		}
	}
	return tr
}

func (tr *throwsReading) evaluate(c throwsMandatedCase, sub map[string]any) {
	available, haveAvailable := boolean(sub, "available")
	if !haveAvailable || !available {
		return
	}
	threw, haveThrew := boolean(sub, "threw")
	if !haveThrew {
		return
	}
	tr.applied++
	if !threw {
		tr.note(c, "completed without throwing, though "+c.mandate+" requires it to throw here")
		return
	}

	name, haveName := str(sub, "name")
	if haveName {
		tr.sawName(name)
	}
	ctorName, haveCtor := str(sub, "ctor")
	instanceOf, haveInstanceOf := boolean(sub, "instanceOf")
	ctorAvailable, haveCtorAvailable := boolean(sub, "ctorGlobalAvailable")
	tag, haveTag := str(sub, "toStringTag")

	if haveCtor && ctorName != c.ctor {
		tr.note(c, "constructor "+clip(ctorName, 60)+", mandate is "+c.ctor)
	}
	if haveInstanceOf && haveCtorAvailable && ctorAvailable && !instanceOf {
		tr.note(c, "thrown value is not instanceof "+c.ctor)
	}
	if haveTag {
		wantTag := "[object Error]"
		if c.domException {
			wantTag = "[object DOMException]"
		}
		if tag != wantTag {
			tr.note(c, "toString tag "+clip(tag, 60)+" instead of "+wantTag)
		}
	}
	if c.domException && haveName && name != c.name {
		tr.note(c, "name "+clip(name, 60)+", mandate is "+c.name)
	}

	a2, haveA2 := object(sub, "attempt2")
	if !haveA2 {
		return
	}
	threw2, haveThrew2 := boolean(a2, "threw")
	switch {
	case haveThrew2 && !threw2:
		tr.note(c, "threw on first attempt, not on identical second")
	case haveThrew2 && threw2:
		if n2, ok2 := str(a2, "name"); ok2 && haveName && n2 != name {
			tr.note(c, "different exception name on second identical attempt")
		}
		if c2, ok2 := str(a2, "ctor"); ok2 && haveCtor && c2 != ctorName {
			tr.note(c, "different constructor on second identical attempt")
		}
		if t2, ok2 := str(a2, "toStringTag"); ok2 && haveTag && t2 != tag {
			tr.note(c, "toString tag disagreed between two identical attempts")
		}
		if i2, ok2 := boolean(a2, "instanceOf"); ok2 && haveInstanceOf && i2 != instanceOf {
			tr.note(c, "instanceof disagreed between two identical attempts")
		}
		if n2, ok2 := str(a2, "name"); ok2 {
			tr.sawName(n2)
		}
	}
}

const throwsMaxRows = 8

func sectionThrows(r Request, _ Inputs, _ claim) Section {
	s := Section{Determination: Inconclusive}

	if !reference.TrustedErrorNames.Verified || !reference.TrustedDOMExceptionNames.Verified {
		s.Determination = Unverified
		s.conclude("the error-name registries this section depends on are not themselves verified")
		return s
	}

	tr := readThrows(r)

	s.add("mandated cases evaluated", countOf(tr.applied, "case(s) evaluated"))
	if len(tr.names) > 0 {
		s.add("error names observed, against the two verified registries", countOf(len(tr.names), "distinct name(s)"))
	}

	if tr.applied == 0 {
		s.conclude("no mandated case could be evaluated")
		return s
	}
	if len(tr.failures) == 0 {
		s.Determination = Consistent
		s.conclude("every mandated case produced exactly the exception type its specification requires")
		return s
	}

	shown := tr.failures
	if len(shown) > throwsMaxRows {
		shown = shown[:throwsMaxRows]
	}
	for _, f := range shown {
		s.add(clip(f.c.label, 120), "did not meet its mandated exception type")
		s.add("specification basis", clip(f.c.label, 120))
	}
	s.Determination = Contradiction
	s.conclude(outOf(len(tr.failures), tr.applied) + " mandated case(s) disagreed with the type their specification fixes")
	return s
}

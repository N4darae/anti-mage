package scan

import (
	"sort"
	"strconv"
	"strings"

	"github.com/N4darae/anti-mage/reference"
)

func sectionThrows(r Request, _ Inputs, _ claim) Section {
	s := Section{Determination: Inconclusive}

	if !reference.TrustedErrorNames.Verified || !reference.TrustedDOMExceptionNames.Verified {
		s.Determination = Unverified
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "the error-name registries this section depends on are not themselves verified",
			Note:  "no observation here carries evidentiary weight until that reference data is confirmed",
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
				addViolation(c, "the operation completed without throwing, though "+c.mandate+" requires it to throw here")
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
				addViolation(c, "reported "+clip(ctorName, 60)+" where "+c.mandate+" mandates "+c.ctor)
			}
			if haveInstanceOf && haveCtorAvailable && ctorAvailable && !instanceOf {
				addViolation(c, "the thrown value is not an instanceof "+c.ctor+", which "+c.mandate+" requires it to be")
			}
			if haveTag {
				wantTag := "[object Error]"
				if c.domException {
					wantTag = "[object DOMException]"
				}
				if tag != wantTag {
					addViolation(c, "Object.prototype.toString reports "+clip(tag, 60)+" instead of "+wantTag)
				}
			}

			if c.domException && haveName && name != c.name {
				addViolation(c, "the thrown value reports the name "+clip(name, 60)+" where "+c.mandate+" mandates "+c.name)
			}

			if a2, haveA2 := object(sub, "attempt2"); haveA2 {
				threw2, haveThrew2 := boolean(a2, "threw")
				switch {
				case haveThrew2 && !threw2:
					addViolation(c, "threw on the first attempt but not on an identical second attempt")
				case haveThrew2 && threw2:
					if n2, ok2 := str(a2, "name"); ok2 && haveName && n2 != name {
						addViolation(c, "reported a different exception name on an identical second attempt")
					}
					if c2, ok2 := str(a2, "ctor"); ok2 && haveCtor && c2 != ctorName {
						addViolation(c, "reported a different constructor on an identical second attempt")
					}
					if t2, ok2 := str(a2, "toStringTag"); ok2 && haveTag && t2 != tag {
						addViolation(c, "Object.prototype.toString disagreed with itself between two identical attempts")
					}
					if i2, ok2 := boolean(a2, "instanceOf"); ok2 && haveInstanceOf && i2 != instanceOf {
						addViolation(c, "the instanceof result disagreed with itself between two identical attempts")
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
		Note:  "not applicable (did not run, or the interface it needs is absent here): " + strconv.Itoa(notApplicable),
	})
	if len(allNames) > 0 {
		registryNote := "every observed name is a recognised language error constructor or a registered DOMException name"
		if len(outsideRegistry) > 0 {
			registryNote = "not present in either verified registry: " + joinLimit(outsideRegistry, 8) +
				"; an honest extension, accessibility tool or content blocker can surface its own error type into a page, " +
				"and this project's own registry is not yet complete, so this by itself is not read as evidence of anything"
		}
		s.Rows = append(s.Rows, Row{
			Label: "error names observed, against the two verified registries",
			Value: strconv.Itoa(len(allNames)) + " distinct name(s)",
			Note:  registryNote,
		})
	}

	if applied == 0 {
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "no mandated case could be evaluated",
			Note:  "the collector reported nothing this engine could test against a specification clause",
		})
		return s
	}

	if len(failures) == 0 {
		s.Determination = Consistent
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "every mandated case produced exactly the exception type its specification requires",
			Note:  "type, identifier, instanceof, Object.prototype.toString and repeatability all agreed",
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
			Note:  strings.Join(f.reasons, "; "),
		})

		s.Rows = append(s.Rows, Row{
			Label: "specification basis",
			Value: clip(f.c.label, 120),
			Note:  f.c.clause,
		})
	}
	s.Determination = Contradiction
	s.Rows = append(s.Rows, Row{
		Label: "conclusion",
		Value: strconv.Itoa(len(failures)) + " of " + strconv.Itoa(applied) + " mandated case(s) disagreed with the type their specification fixes",
		Note:  "a specification-fixed exception type is a fact the environment cannot choose; something between the page and the built-in reported it wrong",
	})
	return s
}

type throwsMandatedCase struct {
	id    string
	label string

	mandate      string
	clause       string
	ctor         string
	name         string
	domException bool
}

var throwsMandatedCases = []throwsMandatedCase{
	{
		id:           "atob.invalidChars",
		mandate:      "the HTML atob() method",
		label:        "atob() on input the forgiving-base64 grammar rejects",
		clause:       "the HTML atob(data) method throws an InvalidCharacterError DOMException when the Infra forgiving-base64 decode algorithm returns failure, which it does whenever the input's code point length is congruent to 1 modulo 4 (html.spec.whatwg.org/multipage/webappapis.html#dom-windowbase64-atob; infra.spec.whatwg.org/#forgiving-base64-decode)",
		ctor:         "DOMException",
		name:         "InvalidCharacterError",
		domException: true,
	},
	{
		id:           "createElement.invalidName",
		mandate:      "DOM's Document.createElement()",
		label:        "document.createElement('') with an empty local name",
		clause:       "DOM's Document.createElement(localName) validate-and-extract step throws an InvalidCharacterError DOMException when localName does not match the XML Name production, which the empty string does not (dom.spec.whatwg.org/#dom-document-createelement, dom.spec.whatwg.org/#validate)",
		ctor:         "DOMException",
		name:         "InvalidCharacterError",
		domException: true,
	},
	{
		id:      "json.malformed",
		mandate: "ECMA-262's JSON.parse",
		label:   "JSON.parse on text that is not valid JSON",
		clause:  "ECMA-262's JSON.parse(text[, reviver]) throws a SyntaxError when text does not parse as a JSON text (tc39.es/ecma262/#sec-json.parse)",
		ctor:    "SyntaxError",
		name:    "SyntaxError",
	},
	{
		id:      "property.accessOnNull",
		mandate: "ECMA-262's property-accessor semantics",
		label:   "reading a property of null",
		clause:  "ECMA-262's runtime semantics for property accessors throw a TypeError when the base value is undefined or null (tc39.es/ecma262/#sec-property-accessors-runtime-semantics-evaluation)",
		ctor:    "TypeError",
		name:    "TypeError",
	},
	{
		id:      "call.nonCallable",
		mandate: "ECMA-262's Call abstract operation",
		label:   "calling a plain object as a function",
		clause:  "ECMA-262's Call abstract operation requires IsCallable(F) to be true and throws a TypeError otherwise (tc39.es/ecma262/#sec-call)",
		ctor:    "TypeError",
		name:    "TypeError",
	},
	{
		id:      "array.negativeLength",
		mandate: "ECMA-262's Array constructor",
		label:   "new Array(-1)",
		clause:  "ECMA-262's Array constructor, given a single numeric argument that is not a valid array length, throws a RangeError (tc39.es/ecma262/#sec-array-len)",
		ctor:    "RangeError",
		name:    "RangeError",
	},
	{
		id:      "decodeURIComponent.malformed",
		mandate: "ECMA-262's decodeURIComponent",
		label:   "decodeURIComponent('%')",
		clause:  "ECMA-262's decodeURIComponent invokes the Decode abstract operation, which throws a URIError when its input is not a valid percent-encoded sequence (tc39.es/ecma262/#sec-decodeuricomponent-uri, tc39.es/ecma262/#sec-decode)",
		ctor:    "URIError",
		name:    "URIError",
	},
	{
		id:           "structuredClone.function",
		mandate:      "HTML's StructuredSerialize",
		label:        "structuredClone of a Function value",
		clause:       "HTML's StructuredSerialize throws a DataCloneError DOMException when asked to serialize a value with no defined serialization, which a callable object is (html.spec.whatwg.org/multipage/structured-data.html#structuredserialize)",
		ctor:         "DOMException",
		name:         "DataCloneError",
		domException: true,
	},
}

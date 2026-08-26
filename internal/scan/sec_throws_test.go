package scan

import (
	"strings"
	"testing"

	"github.com/N4darae/anti-mage/reference"
)

const throwsCaseGood = `{
	"available": true,
	"threw": true,
	"name": "InvalidCharacterError",
	"ctor": "DOMException",
	"ctorGlobalAvailable": true,
	"instanceOf": true,
	"toStringTag": "[object DOMException]",
	"hasStack": true,
	"attempt2": {
		"threw": true,
		"name": "InvalidCharacterError",
		"ctor": "DOMException",
		"ctorGlobalAvailable": true,
		"instanceOf": true,
		"toStringTag": "[object DOMException]",
		"hasStack": true
	}
}`

func throwsAllGoodPayload() map[string]string {
	return map[string]string{
		"throw.mandated": ok(`{
			"atob.invalidChars": ` + throwsCaseGood + `,
			"createElement.invalidName": {"available": false},
			"json.malformed": {
				"available": true, "threw": true, "name": "SyntaxError", "ctor": "SyntaxError",
				"ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object Error]", "hasStack": true,
				"attempt2": {"threw": true, "name": "SyntaxError", "ctor": "SyntaxError", "ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object Error]"}
			},
			"property.accessOnNull": {
				"available": true, "threw": true, "name": "TypeError", "ctor": "TypeError",
				"ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object Error]", "hasStack": true,
				"attempt2": {"threw": true, "name": "TypeError", "ctor": "TypeError", "ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object Error]"}
			},
			"call.nonCallable": {
				"available": true, "threw": true, "name": "TypeError", "ctor": "TypeError",
				"ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object Error]", "hasStack": true,
				"attempt2": {"threw": true, "name": "TypeError", "ctor": "TypeError", "ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object Error]"}
			},
			"array.negativeLength": {
				"available": true, "threw": true, "name": "RangeError", "ctor": "RangeError",
				"ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object Error]", "hasStack": true,
				"attempt2": {"threw": true, "name": "RangeError", "ctor": "RangeError", "ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object Error]"}
			},
			"decodeURIComponent.malformed": {
				"available": true, "threw": true, "name": "URIError", "ctor": "URIError",
				"ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object Error]", "hasStack": true,
				"attempt2": {"threw": true, "name": "URIError", "ctor": "URIError", "ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object Error]"}
			},
			"structuredClone.function": {
				"available": true, "threw": true, "name": "DataCloneError", "ctor": "DOMException",
				"ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object DOMException]", "hasStack": true,
				"attempt2": {"threw": true, "name": "DataCloneError", "ctor": "DOMException", "ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object DOMException]"}
			}
		}`),
	}
}

func TestThrowsEveryMandatedCaseAgreeingIsConsistent(t *testing.T) {
	r := probes(t, throwsAllGoodPayload())
	sec := sectionThrows(r, Inputs{}, claim{})
	if sec.Determination != Consistent {
		t.Fatalf("determination = %q, want consistent; rows: %+v", sec.Determination, sec.Rows)
	}
}

func TestThrowsNoMandatedProbeIsInconclusive(t *testing.T) {
	r := probes(t, map[string]string{})
	sec := sectionThrows(r, Inputs{}, claim{})
	if sec.Determination != Inconclusive {
		t.Fatalf("determination = %q, want inconclusive on an empty request", sec.Determination)
	}
}

func TestThrowsUnsupportedMandatedProbeIsInconclusive(t *testing.T) {
	r := probes(t, map[string]string{
		"throw.mandated": unsup("no throw probe"),
	})
	sec := sectionThrows(r, Inputs{}, claim{})
	if sec.Determination != Inconclusive {
		t.Fatalf("determination = %q, want inconclusive on an unsupported probe", sec.Determination)
	}
}

func TestThrowsWrongConstructorIsContradiction(t *testing.T) {
	payload := throwsAllGoodPayload()
	payload["throw.mandated"] = ok(`{
		"atob.invalidChars": {
			"available": true, "threw": true, "name": "Error", "ctor": "Error",
			"ctorGlobalAvailable": true, "instanceOf": false, "toStringTag": "[object Error]", "hasStack": true,
			"attempt2": {"threw": true, "name": "Error", "ctor": "Error", "ctorGlobalAvailable": true, "instanceOf": false, "toStringTag": "[object Error]"}
		}
	}`)
	r := probes(t, payload)
	sec := sectionThrows(r, Inputs{}, claim{})
	if sec.Determination != Contradiction {
		t.Fatalf("determination = %q, want contradiction; rows: %+v", sec.Determination, sec.Rows)
	}
}

func TestThrowsWrongDOMExceptionNameIsContradiction(t *testing.T) {
	payload := throwsAllGoodPayload()
	payload["throw.mandated"] = ok(`{
		"structuredClone.function": {
			"available": true, "threw": true, "name": "NotSupportedError", "ctor": "DOMException",
			"ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object DOMException]", "hasStack": true,
			"attempt2": {"threw": true, "name": "NotSupportedError", "ctor": "DOMException", "ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object DOMException]"}
		}
	}`)
	r := probes(t, payload)
	sec := sectionThrows(r, Inputs{}, claim{})
	if sec.Determination != Contradiction {
		t.Fatalf("determination = %q, want contradiction; rows: %+v", sec.Determination, sec.Rows)
	}
}

func TestThrowsNoThrowAtAllIsContradiction(t *testing.T) {
	payload := throwsAllGoodPayload()
	payload["throw.mandated"] = ok(`{
		"json.malformed": {"available": true, "threw": false}
	}`)
	r := probes(t, payload)
	sec := sectionThrows(r, Inputs{}, claim{})
	if sec.Determination != Contradiction {
		t.Fatalf("determination = %q, want contradiction; rows: %+v", sec.Determination, sec.Rows)
	}
}

func TestThrowsInstanceofDisagreementIsContradiction(t *testing.T) {
	payload := throwsAllGoodPayload()
	payload["throw.mandated"] = ok(`{
		"call.nonCallable": {
			"available": true, "threw": true, "name": "TypeError", "ctor": "TypeError",
			"ctorGlobalAvailable": true, "instanceOf": false, "toStringTag": "[object Error]", "hasStack": true,
			"attempt2": {"threw": true, "name": "TypeError", "ctor": "TypeError", "ctorGlobalAvailable": true, "instanceOf": false, "toStringTag": "[object Error]"}
		}
	}`)
	r := probes(t, payload)
	sec := sectionThrows(r, Inputs{}, claim{})
	if sec.Determination != Contradiction {
		t.Fatalf("determination = %q, want contradiction; rows: %+v", sec.Determination, sec.Rows)
	}
}

func TestThrowsSecondAttemptDisagreementIsContradiction(t *testing.T) {
	payload := throwsAllGoodPayload()
	payload["throw.mandated"] = ok(`{
		"array.negativeLength": {
			"available": true, "threw": true, "name": "RangeError", "ctor": "RangeError",
			"ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object Error]", "hasStack": true,
			"attempt2": {"threw": false}
		}
	}`)
	r := probes(t, payload)
	sec := sectionThrows(r, Inputs{}, claim{})
	if sec.Determination != Contradiction {
		t.Fatalf("determination = %q, want contradiction; rows: %+v", sec.Determination, sec.Rows)
	}
}

func TestThrowsUnavailableInterfaceIsNotAContradiction(t *testing.T) {
	r := probes(t, map[string]string{
		"throw.mandated": ok(`{"structuredClone.function": {"available": false}}`),
	})
	sec := sectionThrows(r, Inputs{}, claim{})
	if sec.Determination == Contradiction {
		t.Fatalf("determination = contradiction on an unavailable interface")
	}
}

func TestThrowsUnknownNameAloneIsNotAContradiction(t *testing.T) {
	payload := throwsAllGoodPayload()
	payload["throw.names"] = ok(`{"custom.op": {"threw": true, "name": "SomeToolInternalError"}}`)
	r := probes(t, payload)
	sec := sectionThrows(r, Inputs{}, claim{})
	if sec.Determination != Consistent {
		t.Fatalf("determination = %q, want consistent: an unknown name must not defeat an otherwise-agreeing mandated battery", sec.Determination)
	}
	found := false
	for _, row := range sec.Rows {
		if row.Label == "error names observed, against the two verified registries" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a row reporting the names check; rows: %+v", sec.Rows)
	}
}

func TestThrowsMandatedDOMExceptionNamesAreRecognised(t *testing.T) {
	for _, want := range []string{"InvalidCharacterError", "DataCloneError"} {
		if !contains(reference.TrustedDOMExceptionNames.Values, want) {
			t.Fatalf("%q is missing from the DOMException name registry this section reads", want)
		}
	}
	r := probes(t, throwsAllGoodPayload())
	sec := sectionThrows(r, Inputs{}, claim{})
	if sec.Determination != Consistent {
		t.Fatalf("determination = %q, want consistent", sec.Determination)
	}
	found := false
	for _, row := range sec.Rows {
		if row.Label == "error names observed, against the two verified registries" {
			found = true
			if strings.Contains(row.Note, "not present in either verified registry") {
				t.Errorf("an honest run reported a name as unrecognised: %q", row.Note)
			}
		}
	}
	if !found {
		t.Fatalf("no names row; rows: %+v", sec.Rows)
	}
}

func TestThrowsOneCaseFailingSeveralRequirementsCountsAsOneCase(t *testing.T) {
	payload := throwsAllGoodPayload()
	payload["throw.mandated"] = ok(`{
		"atob.invalidChars": {
			"available": true, "threw": true, "name": "Error", "ctor": "Error",
			"ctorGlobalAvailable": true, "instanceOf": false, "toStringTag": "[object Error]", "hasStack": true,
			"attempt2": {"threw": true, "name": "Error", "ctor": "Error", "ctorGlobalAvailable": true, "instanceOf": false, "toStringTag": "[object Error]"}
		}
	}`)
	sec := sectionThrows(probes(t, payload), Inputs{}, claim{})
	if sec.Determination != Contradiction {
		t.Fatalf("determination = %q, want contradiction", sec.Determination)
	}
	var conclusion string
	caseRows := 0
	for _, row := range sec.Rows {
		if row.Label == "conclusion" {
			conclusion = row.Value
		}
		if row.Value == "did not meet its mandated exception type" {
			caseRows++
		}
	}
	if conclusion != "1 of 1 mandated case(s) disagreed with the type their specification fixes" {
		t.Errorf("conclusion = %q; one case failing four requirements is one case", conclusion)
	}
	if caseRows != 1 {
		t.Errorf("%d rows for one disagreeing case, want 1", caseRows)
	}
}

func TestThrowsClauseCitationIsNotSplicedIntoProse(t *testing.T) {
	payload := throwsAllGoodPayload()
	payload["throw.mandated"] = ok(`{
		"atob.invalidChars": {
			"available": true, "threw": true, "name": "Error", "ctor": "Error",
			"ctorGlobalAvailable": true, "instanceOf": false, "toStringTag": "[object Error]"
		}
	}`)
	sec := sectionThrows(probes(t, payload), Inputs{}, claim{})
	basis := ""
	for _, row := range sec.Rows {
		if row.Label == "specification basis" {
			basis = row.Note
			continue
		}
		if strings.Contains(row.Note, "html.spec.whatwg.org") {
			t.Errorf("row %q splices a citation into its prose: %q", row.Label, row.Note)
		}
	}
	if !strings.Contains(basis, "html.spec.whatwg.org") {
		t.Errorf("no specification-basis row carried the clause; got %q", basis)
	}
}

func TestThrowsNameRowDoesNotAssertATypeItDisproved(t *testing.T) {
	payload := throwsAllGoodPayload()
	payload["throw.mandated"] = ok(`{
		"atob.invalidChars": {
			"available": true, "threw": true, "name": "Error", "ctor": "Error",
			"ctorGlobalAvailable": true, "instanceOf": false, "toStringTag": "[object Error]"
		}
	}`)
	sec := sectionThrows(probes(t, payload), Inputs{}, claim{})
	for _, row := range sec.Rows {
		if row.Label == "specification basis" {
			continue
		}
		if strings.Contains(row.Note, "the DOMException's name") {
			t.Errorf("row %q calls a plain Error a DOMException: %q", row.Label, row.Note)
		}
	}
}

func TestThrowsHostileMandatedShapeNeverPanics(t *testing.T) {
	hostile := []string{
		`null`,
		`42`,
		`"a bare string"`,
		`[]`,
		`{"atob.invalidChars": null}`,
		`{"atob.invalidChars": "not an object"}`,
		`{"atob.invalidChars": {"available": "not a bool"}}`,
		`{"atob.invalidChars": {"available": true, "threw": "not a bool"}}`,
		`{"atob.invalidChars": {"available": true, "threw": true, "name": {"nested": "object instead of a string"}}}`,
		`{"atob.invalidChars": {"available": true, "threw": true, "name": ` + hugeStringLiteral() + `}}`,
		`{"atob.invalidChars": {"available": true, "threw": true, "attempt2": "not an object"}}`,
		`{"atob.invalidChars": {"available": true, "threw": true, "attempt2": {"threw": null}}}`,
		deeplyNestedJSON(200),
	}
	for i, body := range hostile {
		r := probes(t, map[string]string{"throw.mandated": ok(body)})
		func() {
			defer func() {
				if p := recover(); p != nil {
					t.Errorf("case %d panicked: %v", i, p)
				}
			}()
			sec := sectionThrows(r, Inputs{}, claim{})
			if sec.Determination == "" {
				t.Errorf("case %d: empty determination", i)
			}
		}()
	}
}

func TestThrowsHostileNamesShapeNeverPanics(t *testing.T) {
	hostile := []string{
		`null`,
		`[1,2,3]`,
		`{"x": null}`,
		`{"x": {"threw": true, "name": 12345}}`,
		`{"x": {"threw": true, "name": ` + hugeStringLiteral() + `}}`,
	}
	for i, body := range hostile {
		payload := throwsAllGoodPayload()
		payload["throw.names"] = ok(body)
		r := probes(t, payload)
		func() {
			defer func() {
				if p := recover(); p != nil {
					t.Errorf("case %d panicked: %v", i, p)
				}
			}()
			sectionThrows(r, Inputs{}, claim{})
		}()
	}
}

func hugeStringLiteral() string {
	b := make([]byte, 0, 200002)
	b = append(b, '"')
	for i := 0; i < 200000; i++ {
		b = append(b, 'x')
	}
	b = append(b, '"')
	return string(b)
}

func deeplyNestedJSON(depth int) string {
	open := `{"atob.invalidChars":`
	close := `}`
	s := "1"
	for i := 0; i < depth; i++ {
		s = `{"a":` + s + `}`
	}
	return open + s + close
}

func TestThrowsRowsAreClippedNotUnbounded(t *testing.T) {
	payload := throwsAllGoodPayload()
	payload["throw.mandated"] = ok(`{
		"atob.invalidChars": {
			"available": true, "threw": true, "name": ` + hugeStringLiteral() + `, "ctor": "DOMException",
			"ctorGlobalAvailable": true, "instanceOf": true, "toStringTag": "[object DOMException]"
		}
	}`)
	r := probes(t, payload)
	sec := sectionThrows(r, Inputs{}, claim{})
	for _, row := range sec.Rows {
		if len(row.Note) > 100000 || len(row.Value) > 100000 {
			t.Fatalf("row %q is unbounded: %d/%d bytes", row.Label, len(row.Value), len(row.Note))
		}
	}
}

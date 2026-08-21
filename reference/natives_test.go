package reference

import (
	"strings"
	"testing"
)

func TestNativeToStringAcceptedForms(t *testing.T) {
	forms := NativeToStringForms("isBrave")
	all := append(append([]string{}, forms.Measured.Values...), forms.OtherEngine.Values...)
	if len(all) != 6 {
		t.Fatalf("expected 6 accepted forms, got %d", len(all))
	}
	seen := map[string]bool{}
	for _, f := range all {
		if !strings.Contains(f, "[native code]") {
			t.Errorf("form %q does not contain the native marker", f)
		}
		if seen[f] {
			t.Errorf("duplicate form %q", f)
		}
		seen[f] = true
	}
	if !seen["function isBrave() { [native code] }"] {
		t.Error("the measured V8 form is missing from the accepted set")
	}
	if !forms.Measured.Verified {
		t.Error("Measured forms must be verified")
	}
	if forms.OtherEngine.Verified {
		t.Error("OtherEngine forms must not be verified")
	}
}

func TestMeasuredNativeToStringFormsExcludeWhitespaceVariants(t *testing.T) {
	forms := NativeToStringForms("isBrave")
	measured := forms.Measured.Values
	if len(measured) != 3 {
		t.Fatalf("expected 3 measured forms, got %d: %v", len(measured), measured)
	}
	for _, f := range measured {
		if strings.Contains(f, "\n") {
			t.Errorf("measured form %q contains an other-engine whitespace variant", f)
		}
	}
	for _, f := range forms.OtherEngine.Values {
		if !strings.Contains(f, "\n") {
			t.Errorf("other-engine form %q is missing its whitespace variant", f)
		}
	}
}

func TestBraveNativeToStringMatchesMeasuredForm(t *testing.T) {
	forms := NativeToStringForms("isBrave")
	if len(BraveNativeToString.Values) != 1 || BraveNativeToString.Values[0] != forms.Measured.Values[0] {
		t.Errorf("BraveNativeToString = %v, want %q", BraveNativeToString.Values, forms.Measured.Values[0])
	}
}

func TestNativeFunctionOwnKeys(t *testing.T) {
	if strings.Join(NativeFunctionOwnKeys.Values, ",") != "length,name" {
		t.Errorf("own-key set drifted from the measurement: %v", NativeFunctionOwnKeys.Values)
	}
}

func TestErrorNamesCoverStandardConstructors(t *testing.T) {
	for _, want := range []string{
		"AggregateError", "Error", "EvalError", "RangeError",
		"ReferenceError", "SyntaxError", "TypeError", "URIError",
	} {
		found := false
		for _, got := range TrustedErrorNames.Values {
			if got == want {
				found = true
			}
		}
		if !found {
			t.Errorf("standard error constructor %q is missing from the allowlist", want)
		}
	}
	for _, dom := range TrustedDOMExceptionNames.Values {
		for _, got := range TrustedErrorNames.Values {
			if got == dom {
				t.Errorf("DOMException name %q must not appear among the Error constructors", dom)
			}
		}
	}
}

package scan

import "testing"

func automationOf(t *testing.T, residue string) Section {
	t.Helper()
	return sectionAutomation(probes(t, map[string]string{"auto.residue": ok(residue)}), Inputs{}, claim{})
}

func TestAutomationAgreesWhenTheStackTraceLimitIsPresent(t *testing.T) {
	got := automationOf(t, `{"webdriver":false,"driverProperties":[],"captureStackTraceType":"function","stackTraceLimit":128}`)
	if got.Determination != Consistent {
		t.Fatalf("determination = %q, want %q; rows: %+v", got.Determination, Consistent, got.Rows)
	}
}

func TestAutomationReportsModificationWhenTheStackTraceLimitWasRemoved(t *testing.T) {
	got := automationOf(t, `{"webdriver":false,"driverProperties":[],"captureStackTraceType":"function","stackTraceLimit":null}`)
	if got.Determination != Instrumented {
		t.Fatalf("determination = %q, want %q: an engine that defines captureStackTrace defines stackTraceLimit beside it; rows: %+v", got.Determination, Instrumented, got.Rows)
	}
}

func TestAutomationReadsNothingIntoAMissingLimitWhenTheEngineHasNoCaptureStackTrace(t *testing.T) {
	got := automationOf(t, `{"webdriver":false,"driverProperties":[],"captureStackTraceType":"undefined","stackTraceLimit":null}`)
	if got.Determination != Consistent {
		t.Fatalf("determination = %q, want %q: an engine is never read as lacking a feature it never had", got.Determination, Consistent)
	}
}

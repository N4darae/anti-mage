package scan

import (
	"encoding/json"
	"testing"
)

func mustJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

func scopeProbe(userAgent, platform string, hwc float64, language string, languages []string, timeZone, locale string) Probe {
	obj := map[string]any{
		"userAgent":           userAgent,
		"platform":            platform,
		"hardwareConcurrency": hwc,
	}
	if language != "" {
		obj["language"] = language
	}
	if len(languages) > 0 {
		obj["languages"] = languages
	}
	if timeZone != "" {
		obj["timeZone"] = timeZone
	}
	if locale != "" {
		obj["locale"] = locale
	}
	return Probe{Status: StatusOK, Value: mustJSON(obj)}
}

func honestThreeScopeRequest(t *testing.T) Request {
	t.Helper()
	r := Request{Probes: map[string]Probe{}}
	r.Probes["scope.main"] = scopeProbe("UA/1", "Win32", 8, "en-US", []string{"en-US", "en"}, "Asia/Bangkok", "en-US")
	r.Probes["scope.worker"] = scopeProbe("UA/1", "Win32", 8, "en-US", []string{"en-US", "en"}, "Asia/Bangkok", "en-US")
	r.Probes["scope.iframe"] = scopeProbe("UA/1", "Win32", 8, "en-US", []string{"en-US", "en"}, "Asia/Bangkok", "en-US")
	return r
}

func TestScopesOldThreeScopePayloadStillReportsConsistent(t *testing.T) {
	got := sectionScopes(honestThreeScopeRequest(t), Inputs{}, claim{})
	if got.Determination != Consistent {
		t.Fatalf("determination = %q, want %q: a payload carrying only scope.main, scope.worker and scope.iframe must score exactly as it did before the new realms were added", got.Determination, Consistent)
	}
}

func TestScopesOldThreeScopePayloadDoesNotMoveTheSummaryComparedToBeforeTheNewRealmsExisted(t *testing.T) {
	sections := []Section{
		{ID: "a", Determination: Consistent},
		{ID: "b", Determination: Consistent},
		{ID: "c", Determination: Inconclusive},
	}

	got := sectionScopes(honestThreeScopeRequest(t), Inputs{}, claim{})
	got.ID = "scopes"

	withOldReading := summarise(append(append([]Section{}, sections...), normalise(Section{ID: "scopes", Determination: Consistent})))
	withCurrentCode := summarise(append(sections, normalise(got)))

	if withOldReading.Band != withCurrentCode.Band {
		t.Errorf("band moved from %q to %q for a payload that predates the new realms", withOldReading.Band, withCurrentCode.Band)
	}
	if withOldReading.HumanConfidence != withCurrentCode.HumanConfidence {
		t.Errorf("human confidence moved from %d to %d for a payload that predates the new realms", withOldReading.HumanConfidence, withCurrentCode.HumanConfidence)
	}
	if withOldReading.BotLikeness != withCurrentCode.BotLikeness {
		t.Errorf("bot likeness moved from %d to %d for a payload that predates the new realms", withOldReading.BotLikeness, withCurrentCode.BotLikeness)
	}
}

func TestScopesNewRealmsAgreeingStillReportConsistent(t *testing.T) {
	r := honestThreeScopeRequest(t)
	r.Probes["scope.workerNested"] = scopeProbe("UA/1", "Win32", 8, "", nil, "", "")
	r.Probes["scope.iframeSrcdoc"] = scopeProbe("UA/1", "Win32", 8, "en-US", []string{"en-US", "en"}, "Asia/Bangkok", "en-US")
	r.Probes["scope.iframeBlob"] = scopeProbe("UA/1", "Win32", 8, "en-US", []string{"en-US", "en"}, "Asia/Bangkok", "en-US")

	got := sectionScopes(r, Inputs{}, claim{})
	if got.Determination != Consistent {
		t.Fatalf("determination = %q, want %q: every realm reported the same facts, including two workers honestly missing timeZone/locale", got.Determination, Consistent)
	}
}

func TestScopesASrcdocFrameThatDisagreesIsInstrumented(t *testing.T) {
	r := honestThreeScopeRequest(t)
	r.Probes["scope.iframeSrcdoc"] = scopeProbe("UA/DIFFERENT", "Win32", 8, "", nil, "", "")

	got := sectionScopes(r, Inputs{}, claim{})
	if got.Determination != Instrumented {
		t.Fatalf("determination = %q, want %q: the srcdoc frame's userAgent does not match the other scopes of the same page", got.Determination, Instrumented)
	}
}

func TestScopesANestedWorkerThatDisagreesIsInstrumented(t *testing.T) {
	r := honestThreeScopeRequest(t)
	r.Probes["scope.workerNested"] = scopeProbe("UA/1", "Linux armv7l", 8, "", nil, "", "")

	got := sectionScopes(r, Inputs{}, claim{})
	if got.Determination != Instrumented {
		t.Fatalf("determination = %q, want %q: the nested worker's platform does not match the other scopes of the same page", got.Determination, Instrumented)
	}
}

func TestScopesARealmThatCouldNotBeCreatedIsNotEvidence(t *testing.T) {
	r := honestThreeScopeRequest(t)
	r.Probes["scope.workerNested"] = Probe{Status: StatusUnsupported}

	got := sectionScopes(r, Inputs{}, claim{})
	if got.Determination != Consistent {
		t.Fatalf("determination = %q, want %q: a browser that refused to create a realm must score exactly as one that created it and agreed", got.Determination, Consistent)
	}
	found := false
	for _, row := range got.Rows {
		if row.Label == "worker spawned by another worker" && row.Value == "the browser would not create this scope" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected a row explaining the nested worker was never created, got %+v", got.Rows)
	}
}

func TestScopesAnErroredRealmIsNotEvidence(t *testing.T) {
	r := honestThreeScopeRequest(t)
	r.Probes["scope.iframeSrcdoc"] = Probe{Status: StatusError}

	got := sectionScopes(r, Inputs{}, claim{})
	if got.Determination != Consistent {
		t.Fatalf("determination = %q, want %q: a scope the collector could not read from must not lower the score", got.Determination, Consistent)
	}
}

func TestScopesIsOneOfTheSectionsAnalyzeBuilds(t *testing.T) {
	found := false
	for _, s := range order {
		if s.id == "scopes" {
			found = true
		}
	}
	if !found {
		t.Fatal("the reading is not registered in the section order, so no scan runs it")
	}
}

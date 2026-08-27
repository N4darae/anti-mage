package scan

import "testing"

func permissionsRequest(t *testing.T, body string) Request {
	t.Helper()
	if body == "" {
		return Request{Probes: map[string]Probe{}}
	}
	return Request{Probes: map[string]Probe{
		"perm.state": {Status: StatusOK, Value: []byte(body)},
	}}
}

func TestPermissionsIsUnverifiedWhenNothingWasCollected(t *testing.T) {
	got := sectionPermissions(permissionsRequest(t, ""), Inputs{}, claim{})
	if got.Determination != Unverified {
		t.Fatalf("determination = %q, want %q: a payload with no permission observation leaves this reading nothing to apply to", got.Determination, Unverified)
	}
}

func TestPermissionsIsUnverifiedWhenTheObservationNamesNeitherInterface(t *testing.T) {
	got := sectionPermissions(permissionsRequest(t, `{}`), Inputs{}, claim{})
	if got.Determination != Unverified {
		t.Fatalf("determination = %q, want %q", got.Determination, Unverified)
	}
}

func TestPermissionsCarriesNoWeightWhenOnlyOneInterfaceWasReported(t *testing.T) {
	got := sectionPermissions(permissionsRequest(t, `{"notifications":{"query":"prompt"}}`), Inputs{}, claim{})
	if got.Determination != Unverified {
		t.Fatalf("determination = %q, want %q: one interface reporting is not two facts to compare, so the reading has nothing to apply to", got.Determination, Unverified)
	}
}

func TestPermissionsCarriesNoWeightOnAnUntabulatedState(t *testing.T) {
	got := sectionPermissions(permissionsRequest(t, `{"notifications":{"query":"unknown-state","actual":"default"}}`), Inputs{}, claim{})
	if got.Determination != Unverified {
		t.Fatalf("determination = %q, want %q: a state this project's table cannot place carries no weight", got.Determination, Unverified)
	}
}

func TestPermissionsAgreesWhenBothInterfacesReportTheEquivalentState(t *testing.T) {
	cases := []string{
		`{"notifications":{"query":"granted","actual":"granted"}}`,
		`{"notifications":{"query":"denied","actual":"denied"}}`,
		`{"notifications":{"query":"prompt","actual":"default"}}`,
	}
	for _, body := range cases {
		got := sectionPermissions(permissionsRequest(t, body), Inputs{}, claim{})
		if got.Determination != Consistent {
			t.Errorf("body %s: determination = %q, want %q", body, got.Determination, Consistent)
		}
	}
}

func TestPermissionsContradictsWhenTheTwoInterfacesDisagree(t *testing.T) {
	got := sectionPermissions(permissionsRequest(t, `{"notifications":{"query":"denied","actual":"granted"}}`), Inputs{}, claim{})
	if got.Determination != Contradiction {
		t.Fatalf("determination = %q, want %q", got.Determination, Contradiction)
	}
}

func TestPermissionsAcceptsTheOlderFlatShape(t *testing.T) {
	got := sectionPermissions(permissionsRequest(t, `{"notifications":"granted","notificationPermission":"granted"}`), Inputs{}, claim{})
	if got.Determination != Consistent {
		t.Fatalf("determination = %q, want %q: this shape predates the nested one and must keep reading", got.Determination, Consistent)
	}
}

func TestPermissionsNamesNoVendorInItsConclusion(t *testing.T) {
	got := sectionPermissions(permissionsRequest(t, `{"notifications":{"query":"denied","actual":"granted"}}`), Inputs{}, claim{})
	for _, row := range got.Rows {
		if row.Label != "conclusion" {
			continue
		}
		for _, forbidden := range []string{"Chrome", "Chromium", "Firefox", "Safari", "Edge"} {
			if contains([]string{row.Value, row.Note}, forbidden) {
				t.Errorf("the conclusion names %q: %q / %q", forbidden, row.Value, row.Note)
			}
		}
	}
}

func TestPermissionsDoesNotDiluteAPayloadThatPredatesIt(t *testing.T) {
	sections := []Section{
		{ID: "a", Determination: Consistent},
		{ID: "b", Determination: Consistent},
		{ID: "c", Determination: Inconclusive},
	}
	before := summarise(sections)

	absent := sectionPermissions(permissionsRequest(t, ""), Inputs{}, claim{})
	absent.ID = "permissions"
	after := summarise(append(sections, normalise(absent)))

	if before.Band != after.Band {
		t.Errorf("band moved from %q to %q merely by adding a reading the payload has no data for", before.Band, after.Band)
	}
	if before.HumanConfidence != after.HumanConfidence {
		t.Errorf("human confidence moved from %d to %d", before.HumanConfidence, after.HumanConfidence)
	}
	if before.BotLikeness != after.BotLikeness {
		t.Errorf("bot likeness moved from %d to %d", before.BotLikeness, after.BotLikeness)
	}
}

func TestPermissionsIsOneOfTheSectionsAnalyzeBuilds(t *testing.T) {
	found := false
	for _, s := range order {
		if s.id == "permissions" {
			found = true
		}
	}
	if !found {
		t.Fatal("the reading is not registered in the section order, so no scan runs it")
	}
}

func TestPermissionsCarriesNoWeightOnEveryPathThatSettlesNothing(t *testing.T) {
	bodies := map[string]string{
		"not collected":            "",
		"neither interface named":  `{"geolocation":{"query":"prompt"}}`,
		"only the query reported":  `{"notifications":{"query":"prompt"}}`,
		"only the other reported":  `{"notificationPermission":"default"}`,
		"a state this table lacks": `{"notifications":{"query":"somethingelse","actual":"default"}}`,
	}

	settled := []Section{
		{ID: "a", Determination: Consistent},
		{ID: "b", Determination: Consistent},
		{ID: "c", Determination: Inconclusive},
	}
	before := summarise(settled)

	for name, body := range bodies {
		r := Request{Probes: map[string]Probe{}}
		if body != "" {
			r.Probes["perm.state"] = Probe{Status: StatusOK, Value: []byte(body)}
		}
		got := sectionPermissions(r, Inputs{}, claim{})
		if got.Determination != Unverified {
			t.Errorf("%s: determination = %q, want %q", name, got.Determination, Unverified)
		}
		got.ID = "permissions"
		after := summarise(append(settled, normalise(got)))
		if after.Band != before.Band || after.HumanConfidence != before.HumanConfidence || after.BotLikeness != before.BotLikeness {
			t.Errorf("%s: the summary moved from %+v to %+v", name, before, after)
		}
	}
}

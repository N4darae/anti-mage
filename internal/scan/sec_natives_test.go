package scan

import "testing"

var domChainNativeMembers = []struct {
	id     string
	name   string
	getter bool
}{
	{"navigator.platform", "platform", true},
	{"screen.width", "width", true},
	{"Node.prototype.nodeType", "nodeType", true},
	{"Node.prototype.textContent", "textContent", true},
	{"Element.prototype.tagName", "tagName", true},
	{"Element.prototype.getAttribute", "getAttribute", false},
	{"EventTarget.prototype.addEventListener", "addEventListener", false},
	{"HTMLElement.prototype.click", "click", false},
	{"Document.prototype.createElement", "createElement", false},
}

func honestNativePayload(members []struct {
	id     string
	name   string
	getter bool
}) map[string]string {
	tostring, ownkeys, descriptor, receiver := "{", "{", "{", "{"
	for i, m := range members {
		if i > 0 {
			tostring += ","
			ownkeys += ","
			descriptor += ","
			receiver += ","
		}
		form := "function " + m.name + "() { [native code] }"
		if m.getter {
			form = "function get " + m.name + "() { [native code] }"
		}
		tostring += `"` + m.id + `":"` + form + `"`
		ownkeys += `"` + m.id + `":{"ownKeys":["length","name"],"getOwnPropertyNames":["length","name"],"descriptors":["length","name"]}`
		descriptor += `"` + m.id + `":{"onPrototype":true}`
		receiver += `"` + m.id + `":{"threw":true,"name":"TypeError"}`
	}
	tostring += "}"
	ownkeys += "}"
	descriptor += "}"
	receiver += "}"
	return map[string]string{
		"native.tostring":   ok(tostring),
		"native.ownkeys":    ok(ownkeys),
		"native.descriptor": ok(descriptor),
		"native.receiver":   ok(receiver),
	}
}

func TestSectionNativesConsistentAcrossAWidenedDOMChainSweep(t *testing.T) {
	r := probes(t, honestNativePayload(domChainNativeMembers))
	s := section(t, Analyze(r, Inputs{}), "natives")
	if s.Determination != Consistent {
		t.Fatalf("determination = %q, want %q: every requirement held on every accessor, across five prototypes, exactly as the four-member navigator-only sweep already reports it", s.Determination, Consistent)
	}
}

func TestSectionNativesInstrumentedNamesOnlyTheDOMChainMemberThatFailed(t *testing.T) {
	kv := honestNativePayload(domChainNativeMembers)
	kv["native.receiver"] = ok(`{"navigator.platform":{"threw":true,"name":"TypeError"},` +
		`"screen.width":{"threw":true,"name":"TypeError"},` +
		`"Node.prototype.nodeType":{"threw":false,"resultType":"number"},` +
		`"Node.prototype.textContent":{"threw":true,"name":"TypeError"},` +
		`"Element.prototype.tagName":{"threw":true,"name":"TypeError"},` +
		`"Element.prototype.getAttribute":{"threw":true,"name":"TypeError"},` +
		`"EventTarget.prototype.addEventListener":{"threw":true,"name":"TypeError"},` +
		`"HTMLElement.prototype.click":{"threw":true,"name":"TypeError"},` +
		`"Document.prototype.createElement":{"threw":true,"name":"TypeError"}}`)
	r := probes(t, kv)
	s := section(t, Analyze(r, Inputs{}), "natives")
	if s.Determination != Instrumented {
		t.Fatalf("determination = %q, want %q", s.Determination, Instrumented)
	}
	sawFailing, sawOthers := false, false
	for _, row := range s.Rows {
		if row.Label == "Node.prototype.nodeType" {
			sawFailing = true
		}
		for _, other := range []string{"textContent", "tagName", "getAttribute", "addEventListener", "click", "createElement", "navigator.platform", "screen.width"} {
			if row.Label == other || row.Label == "Element.prototype."+other || row.Label == "EventTarget.prototype."+other || row.Label == "HTMLElement.prototype."+other || row.Label == "Document.prototype."+other || row.Label == "Node.prototype."+other {
				sawOthers = true
			}
		}
	}
	if !sawFailing {
		t.Errorf("no row named the one accessor that failed its brand check")
	}
	if sawOthers {
		t.Errorf("a row named an accessor whose evidence came through cleanly; a disagreement must only be excused, or reported, for the accessor its own evidence came through")
	}
}

func TestWideningTheDOMChainSweepDoesNotMoveTheHonestScore(t *testing.T) {
	narrow := []struct {
		id     string
		name   string
		getter bool
	}{
		{"navigator.platform", "platform", true},
		{"screen.width", "width", true},
	}
	baseSections := []Section{
		{ID: "a", Determination: Consistent},
		{ID: "b", Determination: Consistent},
	}

	narrowReq := probes(t, honestNativePayload(narrow))
	narrowSection := sectionNatives(narrowReq, Inputs{}, readClaim(narrowReq))
	narrowSection.ID = "natives"
	before := summarise(append(append([]Section{}, baseSections...), normalise(narrowSection)))

	wideReq := probes(t, honestNativePayload(domChainNativeMembers))
	wideSection := sectionNatives(wideReq, Inputs{}, readClaim(wideReq))
	wideSection.ID = "natives"
	after := summarise(append(append([]Section{}, baseSections...), normalise(wideSection)))

	if before.Band != after.Band {
		t.Errorf("band moved from %q to %q merely by widening how many accessors an honest, agreeing sweep examined", before.Band, after.Band)
	}
	if before.HumanConfidence != after.HumanConfidence {
		t.Errorf("human confidence moved from %d to %d merely by widening the sweep", before.HumanConfidence, after.HumanConfidence)
	}
	if before.BotLikeness != after.BotLikeness {
		t.Errorf("bot likeness moved from %d to %d merely by widening the sweep", before.BotLikeness, after.BotLikeness)
	}
	if narrowSection.Determination != Consistent || wideSection.Determination != Consistent {
		t.Fatalf("both the narrow and the widened sweep must agree on an honest payload: narrow=%q wide=%q", narrowSection.Determination, wideSection.Determination)
	}
}

func TestSectionNativesStillAbstainsWithNoDOMChainDataAtAll(t *testing.T) {
	r := probes(t, map[string]string{})
	s := section(t, Analyze(r, Inputs{}), "natives")
	if s.Determination == Contradiction || s.Determination == Instrumented {
		t.Fatalf("determination = %q on a payload with no native.* data at all; absence must never score", s.Determination)
	}
}

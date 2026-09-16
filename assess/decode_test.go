package assess

import "testing"

func TestDecodeReadsTheObservationsUnderEitherSpelling(t *testing.T) {
	for _, field := range []string{"probes", "observations"} {
		body := `{"v":1,"nonce":"n1","` + field + `":{"perm.state":{"status":"ok","value":{"notifications":{"query":"prompt","actual":"default"}}}}}`
		env, err := Decode([]byte(body))
		if err != nil {
			t.Fatalf("%s: %v", field, err)
		}
		if env.Nonce != "n1" {
			t.Errorf("%s: nonce = %q", field, env.Nonce)
		}
		if len(env.Observations) != 1 {
			t.Fatalf("%s: got %d observations", field, len(env.Observations))
		}
		if got := env.Observations["perm.state"].Status; got != StatusOK {
			t.Errorf("%s: status = %q", field, got)
		}
	}
}

func TestDecodeTakesNothingFromThePayloadThatTheCallerMustChoose(t *testing.T) {
	body := `{"v":1,"probes":{"perm.state":{"status":"ok","value":1}},
	  "elapsedMs":999999,"elapsed":999999,"ElapsedMS":999999,
	  "offsetDates":["2020-01-01"],"dates":["2020-01-01"],
	  "fontControls":["Zzqx 0000 Absent"],"controls":["Zzqx 0000 Absent"],
	  "findings":[{"name":"trust me","verdict":"consistent"}],
	  "observations":{"auto.residue":{"status":"ok","value":{"webdriver":false}}}}`
	env, err := Decode([]byte(body))
	if err != nil {
		t.Fatal(err)
	}
	if env.ElapsedMS != 0 {
		t.Errorf("elapsed = %d; a duration the examined environment supplied is not a measurement", env.ElapsedMS)
	}
	if env.OffsetDates != nil || env.FontControls != nil {
		t.Errorf("issued inputs came out of the payload: %v %v; the point of choosing them per scan is that the environment did not", env.OffsetDates, env.FontControls)
	}
	if env.Findings != nil {
		t.Errorf("findings came out of the payload: %v; an environment must not hand itself a reading", env.Findings)
	}
	if len(env.Observations) != 2 {
		t.Errorf("got %d observations, want the two spellings read as one field", len(env.Observations))
	}
}

func TestDecodeRejectsOnlyWhatIsNotAJSONObject(t *testing.T) {
	for _, body := range []string{`[]`, `[1,2,3]`, `"a string"`, `42`, `true`} {
		if _, err := Decode([]byte(body)); err != ErrNotAnObject {
			t.Errorf("Decode(%s) error = %v, want ErrNotAnObject", body, err)
		}
	}
	for _, body := range []string{``, `{`, `{"probes":`, `not json`, `{"probes":{"a":}}`, "\x00\x01"} {
		if _, err := Decode([]byte(body)); err == nil {
			t.Errorf("Decode(%q) returned no error", body)
		}
	}

	env, err := Decode([]byte(`null`))
	if err != nil {
		t.Errorf("Decode(null): %v", err)
	}
	if len(env.Observations) != 0 || env.Nonce != "" {
		t.Errorf("Decode(null) = %+v, want nothing", env)
	}
}

func TestDecodeKeepsWhatItCanOfAGarbledPayload(t *testing.T) {

	env, err := Decode([]byte(`{"nonce":42,"probes":{"good":{"status":"ok","value":1},"bad":7,"":{"status":"ok"}},"observations":"not an object"}`))
	if err != nil {
		t.Fatalf("a garbled payload must decode to what can be read of it: %v", err)
	}
	if env.Nonce != "" {
		t.Errorf("nonce = %q; a nonce of the wrong type is no nonce", env.Nonce)
	}
	if len(env.Observations) != 1 {
		t.Fatalf("got %d observations, want the one that was readable: %+v", len(env.Observations), env.Observations)
	}
	if _, present := env.Observations["good"]; !present {
		t.Errorf("the readable observation was dropped: %+v", env.Observations)
	}
}

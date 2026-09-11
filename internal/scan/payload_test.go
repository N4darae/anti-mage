package scan

import (
	"testing"
	_ "time/tzdata"
)

func TestReadReportsWhyAProbeCouldNotBeRead(t *testing.T) {
	r := Request{Probes: map[string]Probe{
		"ok":          {Status: StatusOK, Value: []byte(`{"a":1}`)},
		"unsupported": {Status: StatusUnsupported, Value: []byte(`{"reason":"no"}`)},
		"errored":     {Status: StatusError, Value: []byte(`{"a":1}`)},
		"empty":       {Status: StatusOK},
		"broken":      {Status: StatusOK, Value: []byte(`{`)},
		"null":        {Status: StatusOK, Value: []byte(`null`)},
		"odd":         {Status: "something else", Value: []byte(`{"a":1}`)},
	}}

	for _, tc := range []struct {
		id   string
		want probeOutcome
	}{
		{"ok", probeReadable},
		{"absent", probeAbsent},
		{"unsupported", probeUnsupported},
		{"errored", probeErrored},
		{"empty", probeUnreadable},
		{"broken", probeUnreadable},
		{"null", probeUnreadable},
		{"odd", probeUnreadable},
	} {
		v, got := r.read(tc.id)
		if got != tc.want {
			t.Errorf("read(%q) outcome = %d, want %d", tc.id, got, tc.want)
		}
		if (v != nil) != (tc.want == probeReadable) {
			t.Errorf("read(%q) returned value %v for outcome %d; only a readable probe carries one", tc.id, v, got)
		}
	}
}

func TestReadAgreesWithTheProbeAccessorsItReplaces(t *testing.T) {
	r := Request{Probes: map[string]Probe{
		"ok":          {Status: StatusOK, Value: []byte(`{"a":1}`)},
		"unsupported": {Status: StatusUnsupported},
		"errored":     {Status: StatusError},
	}}
	for _, id := range []string{"ok", "unsupported", "errored", "absent"} {
		_, outcome := r.read(id)
		if ran := outcome != probeAbsent; ran != r.ran(id) {
			t.Errorf("read(%q) says ran=%v, ran() says %v", id, ran, r.ran(id))
		}
		if unsup := outcome == probeUnsupported; unsup != r.unsupported(id) {
			t.Errorf("read(%q) says unsupported=%v, unsupported() says %v", id, unsup, r.unsupported(id))
		}
		_, readable := r.value(id)
		if got := outcome == probeReadable; got != readable {
			t.Errorf("read(%q) says readable=%v, value() says %v", id, got, readable)
		}
	}
}

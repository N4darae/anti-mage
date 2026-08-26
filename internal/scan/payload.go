package scan

import "encoding/json"

const (
	StatusOK          = "ok"
	StatusUnsupported = "unsupported"
	StatusError       = "error"
)

type Probe struct {
	Status string          `json:"status"`
	Value  json.RawMessage `json:"value"`
}

type Request struct {
	V      int              `json:"v"`
	Mode   string           `json:"mode"`
	Probes map[string]Probe `json:"probes"`

	Nonce string `json:"nonce"`
}

type Inputs struct {
	Nonce string

	OffsetDates []string

	FontControls []string

	ElapsedMS int
}

func DecodeRequest(b []byte) (Request, error) {
	var r Request
	if err := json.Unmarshal(b, &r); err != nil {
		return Request{}, err
	}
	if r.Probes == nil {
		r.Probes = map[string]Probe{}
	}
	return r, nil
}

func (r Request) value(id string) (any, bool) {
	p, ok := r.Probes[id]
	if !ok || p.Status != StatusOK || len(p.Value) == 0 {
		return nil, false
	}
	var v any
	if err := json.Unmarshal(p.Value, &v); err != nil {
		return nil, false
	}
	if v == nil {
		return nil, false
	}
	return v, true
}

func (r Request) ran(id string) bool {
	_, ok := r.Probes[id]
	return ok
}

func (r Request) unsupported(id string) bool {
	p, ok := r.Probes[id]
	return ok && p.Status == StatusUnsupported
}

func (r Request) ReportedNames(id string) []string {
	v, ok := r.value(id)
	if !ok {
		return nil
	}
	if m, isMap := v.(map[string]any); isMap {
		return keys(m)
	}
	names, _ := nameSet(v)
	return names
}

func (r Request) ReportedFontControls() []string {
	return r.ReportedNames("font.controls")
}

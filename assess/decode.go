package assess

import (
	"encoding/json"
	"errors"
)

var ErrNotAnObject = errors.New("assess: the payload is not a JSON object")

var observationFields = []string{"probes", "observations"}

func Decode(b []byte) (Environment, error) {
	var top map[string]json.RawMessage
	if err := json.Unmarshal(b, &top); err != nil {
		var typeErr *json.UnmarshalTypeError
		if errors.As(err, &typeErr) {
			return Environment{}, ErrNotAnObject
		}
		return Environment{}, err
	}

	env := Environment{}
	if raw, ok := top["nonce"]; ok {
		var s string
		if json.Unmarshal(raw, &s) == nil {
			env.Nonce = s
		}
	}

	obs := map[string]Observation{}
	for _, field := range observationFields {
		raw, ok := top[field]
		if !ok {
			continue
		}
		var entries map[string]json.RawMessage
		if json.Unmarshal(raw, &entries) != nil {
			continue
		}
		for id, entry := range entries {
			if id == "" {
				continue
			}
			var o Observation
			if json.Unmarshal(entry, &o) != nil {
				continue
			}
			obs[id] = o
		}
	}
	if len(obs) == 0 {
		obs = nil
	}
	env.Observations = obs
	return env, nil
}

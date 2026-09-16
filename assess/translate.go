package assess

import (
	"sort"
	"strings"

	"github.com/N4darae/anti-mage/internal/scan"
)

func probesOf(env Environment) map[string]scan.Probe {
	out := make(map[string]scan.Probe, len(env.Observations))
	for id, o := range env.Observations {
		if id == "" {
			continue
		}
		out[id] = scan.Probe{Status: normaliseStatus(o.Status), Value: o.Value}
	}
	return out
}

func normaliseStatus(s string) string {
	switch t := strings.ToLower(strings.TrimSpace(s)); t {
	case StatusOK, StatusUnsupported, StatusError:
		return t
	default:
		return s
	}
}

func elapsedOf(env Environment) int {
	if env.ElapsedMS < 0 {
		return 0
	}
	return env.ElapsedMS
}

func suppliedSections(fs []Finding) []scan.Section {
	if len(fs) == 0 {
		return nil
	}
	byName := make(map[string]Verdict, len(fs))
	for _, f := range fs {
		name := strings.TrimSpace(f.Name)
		if name == "" {
			continue
		}
		byName[name] = f.Verdict
	}
	names := make([]string, 0, len(byName))
	for n := range byName {
		names = append(names, n)
	}
	sort.Strings(names)

	out := make([]scan.Section, 0, len(names))
	for _, n := range names {
		out = append(out, scan.Section{
			ID:            "supplied:" + n,
			Title:         n,
			Determination: determinationFor(byName[n]),
		})
	}
	return out
}

func determinationFor(v Verdict) scan.Determination {
	switch v {
	case Consistent:
		return scan.Consistent
	case Contradiction:
		return scan.Contradiction
	case Unverified:
		return scan.Unverified
	case Modified:
		return scan.Instrumented
	}
	return scan.Inconclusive
}

func determinationOf(band string) Determination {
	switch band {
	case scan.BandInsufficient:
		return Insufficient
	case scan.BandCoherent:
		return Coherent
	case scan.BandDiscrepant:
		return Discrepant
	case scan.BandInstrumented:
		return Instrumented
	}
	return NotEvaluated
}

func clampScore(n int) int {
	if n < 0 {
		return 0
	}
	if n > 100 {
		return 100
	}
	return n
}

func suppliedIDs(obs map[string]Observation) []string {
	out := make([]string, 0, len(obs))
	for id := range obs {
		if id == "" {
			continue
		}
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}

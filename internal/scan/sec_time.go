package scan

import (
	"strconv"
	"time"
)

const transitionGuard = 48 * time.Hour

func sectionTime(r Request, in Inputs, _ claim) Section {
	s := Section{Determination: Inconclusive}

	zoneRaw, haveZone := r.value("time.zone")
	if !haveZone {
		s.Rows = append(s.Rows, Row{Label: "time zone", Value: "not collected", Note: "the collector did not report it"})
		return s
	}
	zone, ok := str(zoneRaw, "timeZone")
	if !ok {
		zone, ok = str(zoneRaw, "zone")
	}
	if !ok {
		if s2, isStr := zoneRaw.(string); isStr {
			zone, ok = s2, true
		}
	}
	if locale, have := str(zoneRaw, "locale"); have {
		s.Rows = append(s.Rows, Row{Label: "resolved locale", Value: clip(locale, 80), Note: "reported, not compared: a locale is a preference, not a fact about the host"})
	}
	if !ok || zone == "" {
		s.Rows = append(s.Rows, Row{Label: "time zone", Value: "not reported", Note: "no zone to compare offsets against"})
		return s
	}
	s.Rows = append(s.Rows, Row{Label: "time zone", Value: clip(zone, 80), Note: "from Intl.DateTimeFormat resolvedOptions"})

	loc, err := time.LoadLocation(zone)
	if err != nil {
		s.Determination = Unverified
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "this build's copy of the zone database does not contain that zone",
			Note:  "the browser may carry a newer database than this build; nothing is concluded",
		})
		return s
	}

	samples, haveSamples := readOffsets(r)
	if !haveSamples {
		s.Rows = append(s.Rows, Row{Label: "measured offsets", Value: "not collected", Note: "the collector did not sample getTimezoneOffset"})
		return s
	}

	issued := map[int64]bool{}
	for _, d := range in.OffsetDates {
		if t, err := time.Parse("2006-01-02", d); err == nil {
			issued[t.Add(12*time.Hour).UnixMilli()] = true
		}
	}
	matchedIssued := 0
	for ms := range samples {
		if issued[ms] {
			matchedIssued++
		}
	}
	if len(issued) > 0 && matchedIssued == 0 {
		s.Determination = Unverified
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "none of the instants this server issued were sampled",
			Note:  "the offsets came back for instants the collector chose, which this server cannot treat as unpredictable",
		})
		return s
	}

	usable, mismatched, guarded := 0, 0, 0
	var firstBad string
	for _, ms := range sortedKeys(samples) {
		reported := samples[ms]
		at := time.UnixMilli(ms).UTC()
		if nearTransition(at, loc) {
			guarded++
			continue
		}
		_, offsetSeconds := at.In(loc).Zone()

		want := -offsetSeconds / 60
		usable++
		if int(reported) != want {
			mismatched++
			if firstBad == "" {
				firstBad = at.Format("2006-01-02") + ": reported " + itoa(reported) + ", the zone gives " + strconv.Itoa(want)
			}
		}
	}

	s.Rows = append(s.Rows, Row{
		Label: "instants compared",
		Value: strconv.Itoa(usable) + " of " + strconv.Itoa(len(samples)),
		Note:  strconv.Itoa(guarded) + " discarded for falling within 48 hours of a transition in that zone",
	})
	if len(issued) > 0 {
		s.Rows = append(s.Rows, Row{
			Label: "instants this server issued",
			Value: strconv.Itoa(matchedIssued) + " of " + strconv.Itoa(len(issued)),
			Note:  "chosen per scan so they cannot be answered from a table",
		})
	}
	if usable < 2 {
		s.Rows = append(s.Rows, Row{Label: "conclusion", Value: "too few instants survived", Note: "at least two are required"})
		return s
	}
	if mismatched >= 2 {
		s.Determination = Contradiction
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: strconv.Itoa(mismatched) + " of " + strconv.Itoa(usable) + " offsets do not match the named zone",
			Note:  firstBad,
		})
		return s
	}
	if mismatched == 1 {
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "one offset out of " + strconv.Itoa(usable) + " did not match",
			Note:  "one is not enough: a single zone rule that changed between this build and the browser's would do the same. " + firstBad,
		})
		return s
	}
	s.Determination = Consistent
	s.Rows = append(s.Rows, Row{
		Label: "conclusion",
		Value: "every offset matches the zone the browser names",
		Note:  strconv.Itoa(usable) + " instants compared against this host's IANA zone database",
	})
	return s
}

func readOffsets(r Request) (map[int64]float64, bool) {
	raw, ok := r.value("time.offsets")
	if !ok {
		return nil, false
	}
	out := map[int64]float64{}
	switch t := raw.(type) {
	case map[string]any:
		for k, v := range t {
			f, isNum := v.(float64)
			if !isNum {
				continue
			}
			if ms, ok := parseInstant(k); ok {
				out[ms] = f
			}
		}
	case []any:
		for _, e := range t {
			off, haveOff := num(e, "offset")
			if !haveOff {
				off, haveOff = num(e, "offsetMinutes")
			}
			if !haveOff {
				continue
			}
			if ms, haveMS := num(e, "epochMs"); haveMS {
				out[int64(ms)] = off
				continue
			}
			if d, haveDate := str(e, "date"); haveDate {
				if ms, ok := parseInstant(d); ok {
					out[ms] = off
				}
			}
		}
	}
	if len(out) == 0 {
		return nil, false
	}
	return out, true
}

func parseInstant(k string) (int64, bool) {
	if n, err := strconv.ParseInt(k, 10, 64); err == nil {
		return n, true
	}
	if t, err := time.Parse("2006-01-02", k); err == nil {
		return t.Add(12 * time.Hour).UnixMilli(), true
	}
	if t, err := time.Parse(time.RFC3339, k); err == nil {
		return t.UnixMilli(), true
	}
	return 0, false
}

func nearTransition(at time.Time, loc *time.Location) bool {
	_, mid := at.In(loc).Zone()
	_, before := at.Add(-transitionGuard).In(loc).Zone()
	_, after := at.Add(transitionGuard).In(loc).Zone()
	return before != mid || after != mid
}

func sortedKeys(m map[int64]float64) []int64 {
	out := make([]int64, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j-1] > out[j]; j-- {
			out[j-1], out[j] = out[j], out[j-1]
		}
	}
	return out
}

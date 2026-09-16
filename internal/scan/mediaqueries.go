package scan

import "sort"

const mediaPathMaxEntries = 500

type mediaNumericEntry struct {
	feature, op string
	px          float64
	js, css     bool
	answered    bool
}

type mediaDiscreteEntry struct {
	feature, value string
	js, css        bool
}

type mediaStylesheet struct {
	controlOk               bool
	widthValid, heightValid bool

	numeric  []mediaNumericEntry
	discrete []mediaDiscreteEntry
}

type mediaComplementEntry struct {
	query             string
	matches, negation bool
}

type mediaBracketFlag struct {
	matched, wantTrue bool
	what              string
}

type mediaBracketEntry struct {
	feature string
	flags   []mediaBracketFlag
}

type mediaComplement struct {
	complements []mediaComplementEntry
	brackets    []mediaBracketEntry
}

var mediaBracketChecks = []struct {
	key      string
	wantTrue bool
	what     string
}{
	{"minInside", true, "brackets the viewport a pixel below the value it reads, but did not match"},
	{"maxInside", true, "brackets the viewport a pixel above the value it reads, but did not match"},
	{"minOutside", false, "excludes a value two pixels past the one it reads, but matched anyway"},
	{"maxOutside", false, "excludes a value two pixels before the one it reads, but matched anyway"},
}

func mediaEntries(v any, key string) []any {
	arr, ok := list(v, key)
	if !ok {
		return nil
	}
	if len(arr) > mediaPathMaxEntries {
		arr = arr[:mediaPathMaxEntries]
	}
	return arr
}

func readMediaStylesheet(v any) mediaStylesheet {
	var sheet mediaStylesheet
	sheet.controlOk, _ = boolean(v, "controlOk")
	sheet.widthValid, _ = boolean(v, "widthValid")
	sheet.heightValid, _ = boolean(v, "heightValid")
	if !sheet.controlOk {
		return sheet
	}

	for _, e := range mediaEntries(v, "numeric") {
		feature, haveFeature := str(e, "feature")
		if !haveFeature {
			continue
		}
		entry := mediaNumericEntry{feature: feature}
		var haveJS, haveCSS bool
		entry.js, haveJS = boolean(e, "jsMatches")
		entry.css, haveCSS = boolean(e, "cssMatches")
		entry.answered = haveJS && haveCSS
		entry.op, _ = str(e, "op")
		entry.px, _ = num(e, "px")
		sheet.numeric = append(sheet.numeric, entry)
	}

	for _, e := range mediaEntries(v, "discrete") {
		feature, haveFeature := str(e, "feature")
		value, haveValue := str(e, "value")
		if !haveFeature || !haveValue {
			continue
		}
		entry := mediaDiscreteEntry{feature: feature, value: value}
		entry.js, _ = boolean(e, "jsMatches")
		entry.css, _ = boolean(e, "cssMatches")
		sheet.discrete = append(sheet.discrete, entry)
	}
	return sheet
}

func readMediaComplement(v any) mediaComplement {
	var mc mediaComplement
	for _, e := range mediaEntries(v, "complements") {
		matches, haveMatches := boolean(e, "matches")
		negation, haveNegation := boolean(e, "negationMatches")
		if !haveMatches || !haveNegation {
			continue
		}
		query, _ := str(e, "query")
		mc.complements = append(mc.complements, mediaComplementEntry{query, matches, negation})
	}

	for _, e := range mediaEntries(v, "brackets") {
		feature, haveFeature := str(e, "feature")
		if !haveFeature {
			continue
		}
		entry := mediaBracketEntry{feature: feature}
		for _, c := range mediaBracketChecks {
			matched, have := boolean(e, c.key)
			if !have {
				continue
			}
			entry.flags = append(entry.flags, mediaBracketFlag{matched, c.wantTrue, c.what})
		}
		mc.brackets = append(mc.brackets, entry)
	}
	return mc
}

func (sheet mediaStylesheet) validFor(feature string) bool {
	return (feature == "width" && sheet.widthValid) || (feature == "height" && sheet.heightValid)
}

type mediaDiscreteMismatch struct {
	feature  string
	js, css  string
	disagree bool
}

func (sheet mediaStylesheet) discreteReadings() []mediaDiscreteMismatch {
	type bucket struct{ jsTrue, cssTrue []string }
	buckets := map[string]*bucket{}
	var order []string
	for _, e := range sheet.discrete {
		b := buckets[e.feature]
		if b == nil {
			b = &bucket{}
			buckets[e.feature] = b
			order = append(order, e.feature)
		}
		if e.js {
			b.jsTrue = append(b.jsTrue, e.value)
		}
		if e.css {
			b.cssTrue = append(b.cssTrue, e.value)
		}
	}
	sort.Strings(order)

	var out []mediaDiscreteMismatch
	for _, feature := range order {
		b := buckets[feature]
		if len(b.jsTrue) != 1 || len(b.cssTrue) != 1 {
			continue
		}
		out = append(out, mediaDiscreteMismatch{
			feature:  feature,
			js:       b.jsTrue[0],
			css:      b.cssTrue[0],
			disagree: b.jsTrue[0] != b.cssTrue[0],
		})
	}
	return out
}

package scan

import (
	"strconv"
	_ "time/tzdata"

	"github.com/N4darae/anti-mage/osfont"
)

func sectionFonts(r Request, in Inputs, c claim) Section {
	s := Section{Determination: Inconclusive}

	resolved, missing := readResolvedFonts(r)
	if missing != "" {
		s.add("families resolved", missing)
		return s
	}

	controlNames, controlsResolved, haveControls := readFontControls(r)
	issuedSeen := 0
	for _, n := range in.FontControls {
		if contains(controlNames, n) || contains(controlsResolved, n) {
			issuedSeen++
		}
	}
	if haveControls {
		s.add("invented control families that resolved", outOf(len(controlsResolved), len(controlNames)))
	}

	dropped := coverageFailures(r, resolved)
	res := osfont.EvaluateWindows(resolved)
	tiers := countWindowsTiers(resolved)

	s.add("families resolved by advance width", strconv.Itoa(len(resolved)))
	if len(dropped) > 0 {
		s.add("resolved but own script did not render", joinLimit(dropped, 6))
	}
	for i, t := range tiers.tiers {
		s.add(t.label, outOf(tiers.counts[i], len(t.table.Values)))
	}
	if len(res.Skipped) > 0 {
		s.add("resolved but not read as evidence", joinLimit(res.Skipped, 6))
	}
	s.add("operating system claimed", c.Family.String())

	if !haveControls {
		s.Determination = Unverified
		s.conclude("the width probe carried no invented-name controls")
		return s
	}
	if len(controlsResolved) > 0 {
		s.conclude("the width probe resolved a family that does not exist")
		return s
	}
	if !c.Agreed || c.Family != osWindows {
		s.conclude("no comparison was made")
		return s
	}
	if tiers.verifiedTiersAllEmpty() {
		s.Determination = Contradiction
		s.conclude("Windows is claimed and no family the vendor publishes for Windows resolved")
		return s
	}

	floor := osfont.ReleaseFloor(resolved)
	if len(floor.AboveGap) > 0 {
		s.add("reported above a gap", joinLimit(floor.AboveGap, 3))
	}
	s.Determination = Consistent
	if floor.Release == "" {
		s.add("release floor", "no release narrowed")
	} else {
		s.add("release floor", "Windows "+floor.Release+" or later")
	}
	s.conclude("the resolved families are compatible with a genuine Windows release")
	return s
}

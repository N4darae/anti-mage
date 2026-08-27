package scan

import (
	"regexp"
	"strconv"

	"github.com/N4darae/anti-mage/reference"
)

var chromeMajorPattern = regexp.MustCompile(`Chrome/([0-9]+)`)

func claimedEngineMajor(c claim) (int, bool) {
	m := chromeMajorPattern.FindStringSubmatch(c.UserAgent)
	if m == nil {
		return 0, false
	}
	n, err := strconv.Atoi(m[1])
	if err != nil {
		return 0, false
	}
	return n, true
}

func sectionEngineVersion(r Request, _ Inputs, c claim) Section {
	s := Section{Determination: Unverified}

	claimed, ok := claimedEngineMajor(c)
	if !ok {
		s.Rows = append(s.Rows, Row{
			Label: "engine version, as claimed",
			Value: "not reported",
			Note:  "this reading compares an engine's reported capabilities against the version it claims; without a parseable version there is nothing to compare them to",
		})
		return s
	}
	s.Rows = append(s.Rows, Row{Label: "engine version, as claimed", Value: strconv.Itoa(claimed), Note: "the major version this environment reports for itself"})

	features, ok := r.value("engine.features")
	if !ok {
		s.Rows = append(s.Rows, Row{Label: "engine feature observations", Value: "not collected", Note: "no capability answers were reported, so nothing was compared"})
		return s
	}

	applicable := 0
	laterPresent := []Row{}
	for _, fe := range reference.EngineFeatures {
		if !fe.Verified {
			continue
		}
		if fe.ShipsInMajor <= claimed {
			continue
		}
		present, known := boolean(features, fe.ID)
		if !known {
			continue
		}
		applicable++
		if present {
			laterPresent = append(laterPresent, Row{
				Label: fe.Description,
				Value: "present",
				Note:  "tabulated as shipping at major " + strconv.Itoa(fe.ShipsInMajor) + ", which is later than the version this environment claims; source: " + clip(fe.Source.Origin, 120),
			})
		}
	}

	if applicable == 0 {
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "no tabulated, verified capability ships after the version this environment claims",
			Note:  "either the claimed version already covers everything this table has confirmed on a real system, or none of the tabulated capabilities were reported, so this reading has nothing to apply to this payload",
		})
		return s
	}

	if len(laterPresent) > 0 {
		s.Determination = Contradiction
		s.Rows = append(s.Rows, laterPresent...)
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "this environment demonstrates a capability tabulated as shipping later than the version it claims",
			Note:  "a genuine engine cannot run a capability that its own claimed version predates; the claim and the capability cannot both describe one engine",
		})
		return s
	}

	s.Determination = Consistent
	s.Rows = append(s.Rows, Row{
		Label: "conclusion",
		Value: "no capability later than the claimed version was demonstrated",
		Note:  "",
	})
	return s
}

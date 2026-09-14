package scan

import (
	"regexp"
	"strconv"
	_ "time/tzdata"

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
		s.add("engine version, as claimed", "not reported")
		return s
	}
	s.add("engine version, as claimed", strconv.Itoa(claimed))

	features, ok := r.value("engine.features")
	if !ok {
		s.add("engine feature observations", "not collected")
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
			laterPresent = append(laterPresent, noted(fe.Description, "present"))
		}
	}

	if applicable == 0 {
		s.conclude("no tabulated, verified capability ships after the version this environment claims")
		return s
	}

	if len(laterPresent) > 0 {
		s.Determination = Contradiction
		s.Rows = append(s.Rows, laterPresent...)
		s.conclude("this environment demonstrates a capability tabulated as shipping later than the version it claims")
		return s
	}

	s.Determination = Consistent
	s.conclude("no capability later than the claimed version was demonstrated")
	return s
}

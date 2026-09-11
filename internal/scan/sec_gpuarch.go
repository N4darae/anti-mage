package scan

import (
	"strconv"
	"strings"
	_ "time/tzdata"
)

func sectionGPUArch(r Request, _ Inputs, _ claim) Section {
	s := Section{Determination: Unverified}

	reading, declined, settled := readGPUArch(r)
	if reading.renderer != "" {
		s.add("graphics device, as reported", clip(reading.renderer, 90))
	}
	if !settled {
		s.conclude(declined)
		return s
	}

	s.add("identifier the device named carries", "0x"+strings.ToUpper(strconv.FormatUint(uint64(reading.device), 16)))
	s.add("generation the newer graphics interface reports", reading.reported)
	s.add("generation this project's table places that identifier in", reading.tabulated)

	if reading.agrees {
		s.Determination = Consistent
		s.conclude("the device named and the generation reported for it are the same device")
		return s
	}

	s.Determination, s.weight = Contradiction, weightOnlyDeliberate
	s.conclude("one graphics interface names a device and the other reports a generation that device is not from")
	return s
}

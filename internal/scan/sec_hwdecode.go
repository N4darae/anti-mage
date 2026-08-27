package scan

import (
	"regexp"

	"github.com/N4darae/anti-mage/reference"
)

const (
	hwDecodeControl = "H.264 high, MP4"
	hwDecodeSubject = "AV1, WebM"
)

var geForceRTXModel = regexp.MustCompile(`GeForce RTX ([0-9]{4})`)

func sectionHWDecode(r Request, _ Inputs, _ claim) Section {
	s := Section{Determination: Unverified}

	renderer, ok := hwDecodeRenderer(r)
	if !ok {
		s.Rows = append(s.Rows, Row{
			Label: "graphics device, as reported",
			Value: "not collected",
			Note:  "this reading compares a tabulated device against the decoders reported for it; a payload that names no device leaves it nothing to apply to, so it carries no weight either way",
		})
		return s
	}
	s.Rows = append(s.Rows, Row{Label: "graphics device, as reported", Value: clip(renderer, 90), Note: "the unmasked renderer string, which names the device this environment claims"})

	series, entry, placed := hwDecodeEntry(renderer)
	if !placed {
		s.Rows = append(s.Rows, Row{Label: "conclusion", Value: "this project's table does not place the reported device", Note: "no decoder capability follows from a device this project has not tabulated, so this reading carries no weight for it"})
		return s
	}
	s.Rows = append(s.Rows, Row{Label: "device generation", Value: entry.Family, Note: "read from the " + series + " series in this project's table"})

	if !entry.Verified {
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "the table entry for this generation has not been confirmed on a real system of that configuration",
			Note:  "its documented capability is not read as evidence until it has been observed; source: " + clip(entry.Source.Origin, 120),
		})
		return s
	}

	matrix, ok := r.value("media.matrix")
	if !ok {
		s.Rows = append(s.Rows, Row{Label: "conclusion", Value: "the media capability matrix was not collected", Note: "nothing was compared"})
		return s
	}

	controlEfficient, haveControl := boolean(matrix, hwDecodeControl, "powerEfficient")
	subjectSupported, haveSupport := boolean(matrix, hwDecodeSubject, "decodingInfoSupported")
	subjectEfficient, haveSubject := boolean(matrix, hwDecodeSubject, "powerEfficient")

	s.Rows = append(s.Rows, Row{Label: "a widely implemented codec, decoded in hardware", Value: answerOrAbsent(controlEfficient, haveControl), Note: "the control: it shows whether this environment has a working hardware decoder at all"})
	s.Rows = append(s.Rows, Row{Label: "the tabulated codec, decodable", Value: answerOrAbsent(subjectSupported, haveSupport), Note: "whether this build can decode it by any path"})
	s.Rows = append(s.Rows, Row{Label: "the tabulated codec, decoded in hardware", Value: answerOrAbsent(subjectEfficient, haveSubject), Note: "the reading this generation's tabulated decoder is compared against"})

	if !haveControl || !haveSupport || !haveSubject {
		s.Rows = append(s.Rows, Row{Label: "conclusion", Value: "one of the three readings was not reported", Note: "nothing was compared"})
		return s
	}
	if !controlEfficient {
		s.Rows = append(s.Rows, Row{Label: "conclusion", Value: "no codec was decoded in hardware here", Note: "an environment with no working hardware decoder cannot be read as missing one decoder in particular"})
		return s
	}
	if !subjectSupported {
		s.Rows = append(s.Rows, Row{Label: "conclusion", Value: "this build cannot decode the tabulated codec at all", Note: "a codec a build does not carry is a fact about the build, never a mark against it"})
		return s
	}
	if !entry.AV1Decode {
		s.Rows = append(s.Rows, Row{Label: "conclusion", Value: "this generation is not tabulated as carrying the decoder", Note: "there is nothing for the reported answer to contradict"})
		return s
	}
	if !subjectEfficient {
		s.Determination = Contradiction
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "this environment decodes one codec in hardware and reports no hardware decoder for another that the generation it names carries",
			Note:  "the device named, the working hardware decoder, and the missing one cannot all describe one machine",
		})
		return s
	}

	s.Determination = Consistent
	s.Rows = append(s.Rows, Row{Label: "conclusion", Value: "the decoders reported are the ones the generation this device names carries", Note: ""})
	return s
}

func hwDecodeRenderer(r Request) (string, bool) {
	raw, ok := r.value("gpu.renderer")
	if !ok {
		return "", false
	}
	for _, key := range []string{"unmaskedRenderer", "renderer"} {
		if s, ok := str(raw, key); ok && s != "" {
			return s, true
		}
	}
	return "", false
}

func hwDecodeEntry(renderer string) (string, reference.GPUDecode, bool) {
	m := geForceRTXModel.FindStringSubmatch(renderer)
	if m == nil {
		return "", reference.GPUDecode{}, false
	}
	series := m[1][:2]
	entry, ok := reference.NvidiaGeForceDecode[series]
	if !ok {
		return series, reference.GPUDecode{}, false
	}
	return series, entry, true
}

func answerOrAbsent(value, present bool) string {
	if !present {
		return "not reported"
	}
	if value {
		return "yes"
	}
	return "no"
}

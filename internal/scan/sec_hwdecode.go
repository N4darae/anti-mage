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
			Note:  anomalyNote,
		})
		return s
	}
	s.Rows = append(s.Rows, Row{Label: "graphics device, as reported", Value: clip(renderer, 90), Note: anomalyNote})

	_, entry, placed := hwDecodeEntry(renderer)
	if !placed {
		s.Rows = append(s.Rows, Row{Label: "conclusion", Value: "this project's table does not place the reported device", Note: anomalyNote})
		return s
	}
	s.Rows = append(s.Rows, Row{Label: "device generation", Value: entry.Family, Note: anomalyNote})

	if !entry.Verified {
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "the table entry for this generation has not been confirmed on a real system of that configuration",
			Note:  anomalyNote,
		})
		return s
	}

	matrix, ok := r.value("media.matrix")
	if !ok {
		s.Rows = append(s.Rows, Row{Label: "conclusion", Value: "the media capability matrix was not collected", Note: anomalyNote})
		return s
	}

	controlEfficient, haveControl := boolean(matrix, hwDecodeControl, "powerEfficient")
	subjectSupported, haveSupport := boolean(matrix, hwDecodeSubject, "decodingInfoSupported")
	subjectEfficient, haveSubject := boolean(matrix, hwDecodeSubject, "powerEfficient")

	s.Rows = append(s.Rows, Row{Label: "a widely implemented codec, decoded in hardware", Value: answerOrAbsent(controlEfficient, haveControl), Note: anomalyNote})
	s.Rows = append(s.Rows, Row{Label: "the tabulated codec, decodable", Value: answerOrAbsent(subjectSupported, haveSupport), Note: anomalyNote})
	s.Rows = append(s.Rows, Row{Label: "the tabulated codec, decoded in hardware", Value: answerOrAbsent(subjectEfficient, haveSubject), Note: anomalyNote})

	if !haveControl || !haveSupport || !haveSubject {
		s.Rows = append(s.Rows, Row{Label: "conclusion", Value: "one of the three readings was not reported", Note: anomalyNote})
		return s
	}
	if !controlEfficient {
		s.Rows = append(s.Rows, Row{Label: "conclusion", Value: "no codec was decoded in hardware here", Note: anomalyNote})
		return s
	}
	if !subjectSupported {
		s.Rows = append(s.Rows, Row{Label: "conclusion", Value: "this build cannot decode the tabulated codec at all", Note: anomalyNote})
		return s
	}
	if !entry.AV1Decode {
		s.Rows = append(s.Rows, Row{Label: "conclusion", Value: "this generation is not tabulated as carrying the decoder", Note: anomalyNote})
		return s
	}
	if !subjectEfficient {
		s.Determination = Contradiction
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "this environment decodes one codec in hardware and reports no hardware decoder for another that the generation it names carries",
			Note:  anomalyNote,
		})
		return s
	}

	s.Determination = Consistent
	s.Rows = append(s.Rows, Row{Label: "conclusion", Value: "the decoders reported are the ones the generation this device names carries", Note: anomalyNote})
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

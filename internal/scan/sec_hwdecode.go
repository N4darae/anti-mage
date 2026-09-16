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

	renderer, ok := readGPURenderer(r)
	if !ok {
		s.add("graphics device, as reported", "not collected")
		return s
	}
	s.add("graphics device, as reported", clip(renderer, 90))

	if gpuArchContradicted(r) {
		s.conclude("another reading has already placed the device named in a generation it does not report, so the decoders it should carry settle nothing")
		return s
	}

	_, entry, placed := hwDecodeEntry(renderer)
	if !placed {
		s.conclude("this project's table does not place the reported device")
		return s
	}
	s.add("device generation", entry.Family)

	if !entry.Verified {
		s.conclude("the table entry for this generation has not been confirmed on a real system of that configuration")
		return s
	}

	matrix, ok := r.value("media.matrix")
	if !ok {
		s.conclude("the media capability matrix was not collected")
		return s
	}

	controlEfficient, haveControl := boolean(matrix, hwDecodeControl, "powerEfficient")
	subjectSupported, haveSupport := boolean(matrix, hwDecodeSubject, "decodingInfoSupported")
	subjectEfficient, haveSubject := boolean(matrix, hwDecodeSubject, "powerEfficient")

	s.add("a widely implemented codec, decoded in hardware", answerOrAbsent(controlEfficient, haveControl))
	s.add("the tabulated codec, decodable", answerOrAbsent(subjectSupported, haveSupport))
	s.add("the tabulated codec, decoded in hardware", answerOrAbsent(subjectEfficient, haveSubject))

	if !haveControl || !haveSupport || !haveSubject {
		s.conclude("one of the three readings was not reported")
		return s
	}
	if !controlEfficient {
		s.conclude("no codec was decoded in hardware here")
		return s
	}
	if !subjectSupported {
		s.conclude("this build cannot decode the tabulated codec at all")
		return s
	}
	if !entry.AV1Decode {
		s.conclude("this generation is not tabulated as carrying the decoder")
		return s
	}
	if !subjectEfficient {
		s.Determination = Contradiction
		s.conclude("this environment decodes one codec in hardware and reports no hardware decoder for another that the generation it names carries")
		return s
	}

	s.Determination = Consistent
	s.conclude("the decoders reported are the ones the generation this device names carries")
	return s
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

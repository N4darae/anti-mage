package scan

import (
	_ "time/tzdata"
)

func sectionCanvasSerial(r Request, _ Inputs, _ claim) Section {
	s := Section{Determination: Unverified}

	v, ok := r.value("canvas.serial")
	if !ok {
		s.conclude("not collected")
		return s
	}

	dataAvail, haveDataAvail := boolean(v, "dataUrlAvailable")
	blobAvail, haveBlobAvail := boolean(v, "blobAvailable")
	if !haveDataAvail || !dataAvail || !haveBlobAvail || !blobAvail {
		s.add("canvas serialisation, both paths", "one or both paths unavailable")
		return s
	}

	rawHash, haveRaw := str(v, "rawHash")
	dataHash, haveDataHash := str(v, "dataUrlHash")
	dataMatch, haveDataMatch := boolean(v, "dataUrlPixelsMatchRaw")
	blobHash, haveBlobHash := str(v, "blobHash")
	blobMatch, haveBlobMatch := boolean(v, "blobPixelsMatchRaw")

	if !haveRaw || !haveDataHash || !haveDataMatch || !haveBlobHash || !haveBlobMatch {
		s.conclude("one of the readings was not reported")
		return s
	}

	s.add("raw canvas pixels", rawHash)
	s.add("toDataURL PNG, decoded pixels vs raw canvas", answerOrAbsent(dataMatch, true))
	s.add("toBlob PNG, decoded pixels vs raw canvas", answerOrAbsent(blobMatch, true))

	if !dataMatch || !blobMatch {
		s.conclude("one of the encoded paths did not decode back to the canvas's own pixels")
		return s
	}

	s.add("toDataURL PNG bytes", canvasSerialByteSummary(v, "dataUrlLength", dataHash))
	s.add("toBlob PNG bytes", canvasSerialByteSummary(v, "blobLength", blobHash))

	if dataHash == blobHash {
		s.conclude("both serialisation paths encoded the identical pixels to the identical bytes")
		return s
	}

	s.conclude("both serialisation paths encoded the identical pixels to different bytes")
	return s
}

func canvasSerialByteSummary(v any, lengthKey, hash string) string {
	length, haveLength := num(v, lengthKey)
	if !haveLength {
		return hash + ", length not reported"
	}
	return hash + ", " + itoa(length) + " bytes"
}

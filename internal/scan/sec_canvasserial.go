package scan

func sectionCanvasSerial(r Request, _ Inputs, _ claim) Section {
	s := Section{Determination: Unverified}

	v, ok := r.value("canvas.serial")
	if !ok {
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "not collected",
			Note:  anomalyNote,
		})
		return s
	}

	dataAvail, haveDataAvail := boolean(v, "dataUrlAvailable")
	blobAvail, haveBlobAvail := boolean(v, "blobAvailable")
	if !haveDataAvail || !dataAvail || !haveBlobAvail || !blobAvail {
		s.Rows = append(s.Rows, Row{
			Label: "canvas serialisation, both paths",
			Value: "one or both paths unavailable",
			Note:  anomalyNote,
		})
		return s
	}

	rawHash, haveRaw := str(v, "rawHash")
	dataHash, haveDataHash := str(v, "dataUrlHash")
	dataMatch, haveDataMatch := boolean(v, "dataUrlPixelsMatchRaw")
	blobHash, haveBlobHash := str(v, "blobHash")
	blobMatch, haveBlobMatch := boolean(v, "blobPixelsMatchRaw")

	if !haveRaw || !haveDataHash || !haveDataMatch || !haveBlobHash || !haveBlobMatch {
		s.Rows = append(s.Rows, Row{Label: "conclusion", Value: "one of the readings was not reported", Note: anomalyNote})
		return s
	}

	s.Rows = append(s.Rows, Row{Label: "raw canvas pixels", Value: rawHash, Note: anomalyNote})
	s.Rows = append(s.Rows, Row{Label: "toDataURL PNG, decoded pixels vs raw canvas", Value: answerOrAbsent(dataMatch, true), Note: anomalyNote})
	s.Rows = append(s.Rows, Row{Label: "toBlob PNG, decoded pixels vs raw canvas", Value: answerOrAbsent(blobMatch, true), Note: anomalyNote})

	if !dataMatch || !blobMatch {
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "one of the encoded paths did not decode back to the canvas's own pixels",
			Note:  anomalyNote,
		})
		return s
	}

	s.Rows = append(s.Rows, Row{Label: "toDataURL PNG bytes", Value: canvasSerialByteSummary(v, "dataUrlLength", dataHash), Note: anomalyNote})
	s.Rows = append(s.Rows, Row{Label: "toBlob PNG bytes", Value: canvasSerialByteSummary(v, "blobLength", blobHash), Note: anomalyNote})

	if dataHash == blobHash {
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "both serialisation paths encoded the identical pixels to the identical bytes",
			Note:  anomalyNote,
		})
		return s
	}

	s.Rows = append(s.Rows, Row{
		Label: "conclusion",
		Value: "both serialisation paths encoded the identical pixels to different bytes",
		Note:  anomalyNote,
	})
	return s
}

func canvasSerialByteSummary(v any, lengthKey, hash string) string {
	length, haveLength := num(v, lengthKey)
	if !haveLength {
		return hash + ", length not reported"
	}
	return hash + ", " + itoa(length) + " bytes"
}

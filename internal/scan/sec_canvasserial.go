package scan

func sectionCanvasSerial(r Request, _ Inputs, _ claim) Section {
	s := Section{Determination: Unverified}

	v, ok := r.value("canvas.serial")
	if !ok {
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "not collected",
			Note:  "this reading compares two canvas serialisation paths against each other; a payload that ran neither leaves it nothing to apply to",
		})
		return s
	}

	dataAvail, haveDataAvail := boolean(v, "dataUrlAvailable")
	blobAvail, haveBlobAvail := boolean(v, "blobAvailable")
	if !haveDataAvail || !dataAvail || !haveBlobAvail || !blobAvail {
		s.Rows = append(s.Rows, Row{
			Label: "canvas serialisation, both paths",
			Value: "one or both paths unavailable",
			Note:  "a browser that cannot serialise a canvas by one of these paths is not read as suspicious for lacking it",
		})
		return s
	}

	rawHash, haveRaw := str(v, "rawHash")
	dataHash, haveDataHash := str(v, "dataUrlHash")
	dataMatch, haveDataMatch := boolean(v, "dataUrlPixelsMatchRaw")
	blobHash, haveBlobHash := str(v, "blobHash")
	blobMatch, haveBlobMatch := boolean(v, "blobPixelsMatchRaw")

	if !haveRaw || !haveDataHash || !haveDataMatch || !haveBlobHash || !haveBlobMatch {
		s.Rows = append(s.Rows, Row{Label: "conclusion", Value: "one of the readings was not reported", Note: "nothing was compared"})
		return s
	}

	s.Rows = append(s.Rows, Row{Label: "raw canvas pixels", Value: rawHash, Note: "a hash of the untouched canvas's own pixel buffer, read before either serialisation path ran"})
	s.Rows = append(s.Rows, Row{Label: "toDataURL PNG, decoded pixels vs raw canvas", Value: answerOrAbsent(dataMatch, true), Note: ""})
	s.Rows = append(s.Rows, Row{Label: "toBlob PNG, decoded pixels vs raw canvas", Value: answerOrAbsent(blobMatch, true), Note: ""})

	if !dataMatch || !blobMatch {
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "one of the encoded paths did not decode back to the canvas's own pixels",
			Note:  "the two encoders are not compared to each other while either has already drifted from the content itself; this reading has nothing to say about the encoders until the content is first established as identical",
		})
		return s
	}

	s.Rows = append(s.Rows, Row{Label: "toDataURL PNG bytes", Value: canvasSerialByteSummary(v, "dataUrlLength", dataHash), Note: "a hash and a length, never the encoded bytes themselves"})
	s.Rows = append(s.Rows, Row{Label: "toBlob PNG bytes", Value: canvasSerialByteSummary(v, "blobLength", blobHash), Note: "a hash and a length, never the encoded bytes themselves"})

	if dataHash == blobHash {
		s.Rows = append(s.Rows, Row{
			Label: "conclusion",
			Value: "both serialisation paths encoded the identical pixels to the identical bytes",
			Note:  "recorded, not scored: this project has not established byte-identical output as a specification requirement, only observed it on one baseline build",
		})
		return s
	}

	s.Rows = append(s.Rows, Row{
		Label: "conclusion",
		Value: "both serialisation paths encoded the identical pixels to different bytes",
		Note:  "recorded, not scored: the specification does not require the two paths to produce identical bytes, and this project has not established the honest range of that difference across browser builds; a version difference between the two things being compared could explain it as easily as anything else",
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

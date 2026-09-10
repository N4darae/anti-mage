package scan

import (
	_ "time/tzdata"
)

const audioDurationEpsilonSeconds = 1e-6

const (
	audioScaleMaxRelativeResidual = 2.384185791015625e-07

	audioScaleMinSamples = 32
)

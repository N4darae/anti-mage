package scan

import (
	_ "time/tzdata"
)

const rectTolerance = 1e-4

type rectPoint struct {
	x, y, left, top, right, bottom, width, height                                 float64
	haveX, haveY, haveLeft, haveTop, haveRight, haveBottom, haveWidth, haveHeight bool
}

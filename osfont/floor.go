package osfont

import "github.com/N4darae/anti-mage/reference"

type Floor struct {
	Release  string
	AboveGap []string
	Skipped  []string
}

var releaseOrder = []string{"10", "11"}

func ReleaseFloor(resolved []string) Floor {
	have := make(map[string]bool, len(resolved))
	for _, f := range resolved {
		have[f] = true
	}

	var f Floor
	for _, np := range reference.WindowsNonProbativeMarkers.Values {
		if have[np] {
			f.Skipped = append(f.Skipped, np)
			delete(have, np)
		}
	}

	reached := make([]bool, len(releaseOrder))
	for i, rel := range releaseOrder {
		t, ok := reference.WindowsVersionMarkerFonts[rel]
		if !ok {
			continue
		}
		for _, fam := range t.Values {
			if have[fam] {
				reached[i] = true
				break
			}
		}
	}

	base := false
	for _, fam := range reference.WindowsBaseFonts.Values {
		if have[fam] {
			base = true
			break
		}
	}
	if !base && !anyTrue(reached) {
		return f
	}

	last := -1
	for i := range releaseOrder {
		if !reached[i] {
			break
		}
		last = i
	}
	if last >= 0 {
		f.Release = releaseOrder[last]
	}
	for i := last + 2; i < len(releaseOrder); i++ {
		if reached[i] {
			f.AboveGap = append(f.AboveGap, releaseOrder[i])
		}
	}
	return f
}

func anyTrue(bs []bool) bool {
	for _, b := range bs {
		if b {
			return true
		}
	}
	return false
}

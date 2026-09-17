package server

import (
	"testing"
	"time"
)

func TestOffsetInstantsSpanBothSeasons(t *testing.T) {
	now := time.Date(2026, time.August, 26, 0, 0, 0, 0, time.UTC)
	got, err := offsetInstants(now, 8)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 8 {
		t.Fatalf("got %d instants, want 8", len(got))
	}
	jan, jul := 0, 0
	for _, o := range got {
		switch o.Date[5:7] {
		case "01":
			jan++
		case "07":
			jul++
		}
	}
	if jan == 0 || jul == 0 {
		t.Errorf("got %d January and %d July instants; both seasons are needed so a daylight rule is sampled on each side", jan, jul)
	}
	for _, o := range got {
		if !time.UnixMilli(o.EpochMs).UTC().Before(now) {
			t.Errorf("instant %s is not in the past; the offset for an instant that has not happened is a prediction in both databases, and two predictions disagreeing is not a browser contradicting itself", o.Date)
		}
	}
}

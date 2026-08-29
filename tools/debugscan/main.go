package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/N4darae/anti-mage/internal/scan"
)

type input struct {
	Probes map[string]scan.Probe `json:"probes"`
}

func main() {
	raw, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var in input
	if err := json.Unmarshal(raw, &in); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	rep := scan.AnalyzeWith(scan.Request{V: 1, Mode: "public", Probes: in.Probes}, scan.Inputs{}, nil)
	fmt.Printf("band=%s botLikeness=%d headline=%s\n\n", rep.Summary.Band, rep.Summary.BotLikeness, rep.Summary.Headline)
	for _, s := range rep.Sections {
		fmt.Printf("== %-12s %-14s %s\n", s.ID, s.Determination, s.Title)
		if s.Determination == scan.Contradiction || s.Determination == scan.Instrumented {
			for _, r := range s.Rows {
				fmt.Printf("     %-40s %-40s %s\n", r.Label, r.Value, r.Note)
			}
		}
	}
}

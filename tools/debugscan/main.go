package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/N4darae/anti-mage/internal/scan"
)

type input struct {
	Probes map[string]scan.Probe `json:"probes"`
}

func main() {
	flag.Parse()
	if err := run(flag.Arg(0), os.Stdout); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(path string, w io.Writer) error {
	in, err := load(path)
	if err != nil {
		return err
	}
	req := scan.Request{V: 1, Mode: "public", Probes: in.Probes}
	report(w, scan.AnalyzeWith(req, scan.Inputs{}, nil))
	return nil
}

func load(path string) (input, error) {
	if path == "" {
		return input{}, errors.New("usage: debugscan <payload.json>")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return input{}, err
	}
	var in input
	if err := json.Unmarshal(raw, &in); err != nil {
		return input{}, err
	}
	return in, nil
}

func report(w io.Writer, rep scan.Report) {
	fmt.Fprintf(w, "band=%s botLikeness=%d headline=%s\n\n", rep.Summary.Band, rep.Summary.BotLikeness, rep.Summary.Headline)
	for _, s := range rep.Sections {
		fmt.Fprintf(w, "== %-12s %-14s %s\n", s.ID, s.Determination, s.Title)
		if s.Determination == scan.Contradiction || s.Determination == scan.Instrumented {
			for _, r := range s.Rows {
				fmt.Fprintf(w, "     %-40s %-40s %s\n", r.Label, r.Value, r.Note)
			}
		}
	}
}

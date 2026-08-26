package server

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/N4darae/anti-mage/reference"
)

type Bootstrap struct {
	V        int           `json:"v"`
	Nonce    string        `json:"nonce"`
	IssuedAt string        `json:"issuedAt"`
	Font     BootstrapFont `json:"font"`
	Time     BootstrapTime `json:"time"`

	Fonts BootstrapFont `json:"fonts"`
	Dates []string      `json:"dates"`
}

type BootstrapFont struct {
	Controls      []string `json:"controls"`
	MeasureString string   `json:"measureString"`
	Bases         []string `json:"bases"`
}

type BootstrapTime struct {
	Offsets []OffsetSample `json:"offsets"`
}

type OffsetSample struct {
	Date    string `json:"date"`
	EpochMs int64  `json:"epochMs"`
	ISO     string `json:"iso"`
}

const offsetSampleCount = 8

const controlFamilyCount = 6

const issueTTL = 30 * time.Minute

const maxLiveIssues = 4096

type scanInputs struct {
	Nonce        string
	OffsetDates  []string
	FontControls []string
	ElapsedMS    int
}

type issuer struct {
	mu   sync.Mutex
	live map[string]issued
}

type issued struct {
	inputs  scanInputs
	at      time.Time
	expires time.Time
}

func newIssuer() *issuer { return &issuer{live: map[string]issued{}} }

func (i *issuer) issue(now time.Time) (Bootstrap, error) {
	nonce, err := randomHex(16)
	if err != nil {
		return Bootstrap{}, err
	}
	controls, err := inventedFamilies(controlFamilyCount)
	if err != nil {
		return Bootstrap{}, err
	}
	samples, err := offsetInstants(now, offsetSampleCount)
	if err != nil {
		return Bootstrap{}, err
	}

	dates := make([]string, 0, len(samples))
	iso := make([]string, 0, len(samples))
	for _, s := range samples {
		dates = append(dates, s.Date)
		iso = append(iso, s.ISO)
	}

	i.mu.Lock()
	for k, v := range i.live {
		if now.After(v.expires) {
			delete(i.live, k)
		}
	}
	for len(i.live) >= maxLiveIssues {
		oldest, at := "", time.Time{}
		for k, v := range i.live {
			if oldest == "" || v.at.Before(at) {
				oldest, at = k, v.at
			}
		}
		delete(i.live, oldest)
	}
	i.live[nonce] = issued{
		at: now,
		inputs: scanInputs{
			Nonce:        nonce,
			OffsetDates:  dates,
			FontControls: controls,
		},
		expires: now.Add(issueTTL),
	}
	i.mu.Unlock()

	font := BootstrapFont{
		Controls:      controls,
		MeasureString: reference.FontMeasurementString,
		Bases:         reference.FontMeasurementBases.Values,
	}
	return Bootstrap{
		V:        1,
		Nonce:    nonce,
		IssuedAt: now.UTC().Format(time.RFC3339),
		Font:     font,
		Fonts:    font,
		Dates:    iso,
		Time:     BootstrapTime{Offsets: samples},
	}, nil
}

func (i *issuer) resolve(nonce string, now time.Time) scanInputs {
	if nonce == "" {
		return scanInputs{}
	}
	i.mu.Lock()
	defer i.mu.Unlock()
	e, ok := i.live[nonce]
	if !ok || now.After(e.expires) {
		return scanInputs{}
	}
	return withElapsed(e, now)
}

func withElapsed(e issued, now time.Time) scanInputs {
	in := e.inputs
	if e.at.IsZero() {
		return in
	}
	d := now.Sub(e.at)
	if d < 0 || d > 24*time.Hour {
		return in
	}
	in.ElapsedMS = int(d / time.Millisecond)
	return in
}

func (i *issuer) resolveByControls(reported []string, now time.Time) scanInputs {
	if len(reported) == 0 {
		return scanInputs{}
	}
	have := make(map[string]bool, len(reported))
	for _, n := range reported {
		have[n] = true
	}
	i.mu.Lock()
	defer i.mu.Unlock()
	for _, e := range i.live {
		if now.After(e.expires) || len(e.inputs.FontControls) == 0 {
			continue
		}
		all := true
		for _, n := range e.inputs.FontControls {
			if !have[n] {
				all = false
				break
			}
		}
		if all {
			return withElapsed(e, now)
		}
	}
	return scanInputs{}
}

func offsetInstants(now time.Time, n int) ([]OffsetSample, error) {
	if n < 2 {
		n = 2
	}
	utc := now.UTC()
	out := make([]OffsetSample, 0, n)
	for y := utc.Year(); len(out) < n; y-- {
		for _, month := range []time.Month{time.July, time.January} {
			if len(out) >= n {
				break
			}
			day, err := randomInt(3, 25)
			if err != nil {
				return nil, err
			}
			t := time.Date(y, month, day, 12, 0, 0, 0, time.UTC)
			if !t.Before(utc) {
				continue
			}
			out = append(out, OffsetSample{
				Date:    t.Format("2006-01-02"),
				EpochMs: t.UnixMilli(),
				ISO:     t.Format(time.RFC3339),
			})
		}
	}
	return out, nil
}

func inventedFamilies(n int) ([]string, error) {

	shapes := []func(string) string{
		func(t string) string { return t + " Sans" },
		func(t string) string { return t + " Grotesk" },
		func(t string) string { return "Segoe " + t },
		func(t string) string { return t + " Text Pro" },
		func(t string) string { return "Helvetica " + t },
		func(t string) string { return t + " Display" },
		func(t string) string { return t + " Neue" },
	}
	out := make([]string, 0, n)
	for k := 0; k < n; k++ {
		t, err := inventedToken()
		if err != nil {
			return nil, err
		}
		out = append(out, shapes[k%len(shapes)](t))
	}
	return out, nil
}

const (
	tokenConsonants = "bcdfghjklmnprstvwz"
	tokenVowels     = "aeiou"
	tokenSyllables  = 5
)

func inventedToken() (string, error) {
	b := make([]byte, 0, tokenSyllables*2)
	for k := 0; k < tokenSyllables; k++ {
		c, err := randomInt(0, len(tokenConsonants)-1)
		if err != nil {
			return "", err
		}
		v, err := randomInt(0, len(tokenVowels)-1)
		if err != nil {
			return "", err
		}
		b = append(b, tokenConsonants[c], tokenVowels[v])
	}

	return strings.ToUpper(string(b[0:1])) + string(b[1:]), nil
}

func randomHex(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func randomInt(lo, hi int) (int, error) {
	if hi <= lo {
		return lo, nil
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(hi-lo+1)))
	if err != nil {
		return 0, err
	}
	return lo + int(n.Int64()), nil
}

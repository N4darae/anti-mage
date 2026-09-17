package server

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"strings"
	"time"
)

const offsetSampleCount = 8

const controlFamilyCount = 6

const (
	tokenConsonants = "bcdfghjklmnprstvwz"
	tokenVowels     = "aeiou"
	tokenSyllables  = 5
)

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

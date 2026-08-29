package assess_test

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/N4darae/anti-mage/assess"
)

func Example() {
	issued := time.Now().Add(-412 * time.Millisecond)

	env := assess.Environment{
		Observations: map[string]assess.Observation{
			"scope.main": {Status: assess.StatusOK, Value: json.RawMessage(
				`{"userAgent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/149.0.0.0","platform":"Win32"}`)},
			"native.tostring": {Status: assess.StatusOK, Value: json.RawMessage(
				`{"navigator.platform":"function get platform() { [native code] }"}`)},
			"font.resolved": {Status: assess.StatusUnsupported},
		},

		FontControls: []string{"Zzqx 0000 Absent"},
		OffsetDates:  []string{"2026-01-12", "2026-07-04"},
		ElapsedMS:    int(time.Since(issued) / time.Millisecond),
	}

	a := assess.Evaluate(env)

	if a.Determination.AtLeast(assess.Discrepant) && a.Score >= 30 {
		fmt.Println("acting on it")
	}
	if !a.Determination.Established() {
		fmt.Println("too little was read to characterise this environment either way")
	}
	fmt.Println(a.Determination, a.Score, len(a.Supplied))

	// Output:
	// too little was read to characterise this environment either way
	// insufficient 0 3
}

func ExampleEnvironment_Findings() {
	a := assess.Evaluate(assess.Environment{
		Findings: []assess.Finding{
			{Name: "audio pipeline", Verdict: assess.Contradiction},
			{Name: "codec matrix", Verdict: assess.Unverified},
		},
	})
	fmt.Println(a.Determination, a.Score)

	// Output:
	// discrepant 40
}

func ExampleDecode() {
	body := []byte(`{"v":1,"nonce":"n1",
	  "probes":{"auto.residue":{"status":"ok","value":{"webdriver":true}}},
	  "elapsedMs":999999,"findings":[{"name":"trust me","verdict":"consistent"}]}`)

	env, err := assess.Decode(body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(env.Nonce, len(env.Observations), env.ElapsedMS, len(env.Findings))
	fmt.Println(assess.Evaluate(env).Determination)

	// Output:
	// n1 1 0 0
	// instrumented
}

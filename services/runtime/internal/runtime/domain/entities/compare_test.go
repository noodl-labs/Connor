package entities

import (
	"errors"
	"testing"
)

func sampleArtifact(suiteID string, p95 int64, model string) RunArtifact {
	return sampleArtifactWithPassRate(suiteID, p95, model, 100)
}

func sampleArtifactWithPassRate(suiteID string, p95 int64, model string, passRate float64) RunArtifact {
	return RunArtifact{
		Version: RunArtifactVersion,
		SuiteID: suiteID,
		Cases: []RunCase{
			{ID: "c1", Model: model, Passed: passRate >= 100, LatencyMs: p95},
		},
		Summary: RunSummary{Total: 1, Passed: 1, P95Ms: p95, P50Ms: p95, PassRate: passRate},
	}
}

func TestValidateComparable_ok(t *testing.T) {
	b := sampleArtifact("suite-a", 100, "gpt-4o-mini")
	c := sampleArtifact("suite-a", 120, "gpt-4o-mini")
	if err := ValidateComparable(b, c); err != nil {
		t.Fatal(err)
	}
}

func TestValidateComparable_suiteMismatch(t *testing.T) {
	b := sampleArtifact("a", 100, "m")
	c := sampleArtifact("b", 100, "m")
	if err := ValidateComparable(b, c); !errors.Is(err, ErrNotComparable) {
		t.Fatalf("got %v", err)
	}
}

func TestValidateComparable_modelMismatch(t *testing.T) {
	b := sampleArtifact("s", 100, "gpt-4")
	c := sampleArtifact("s", 100, "gpt-5")
	if err := ValidateComparable(b, c); !errors.Is(err, ErrNotComparable) {
		t.Fatalf("got %v", err)
	}
}

func TestP95RegressionPercent(t *testing.T) {
	if got := P95RegressionPercent(100, 108); got < 7.9 || got > 8.1 {
		t.Fatalf("got %v want ~8", got)
	}
	if got := P95RegressionPercent(754, 1800); got < 137 || got > 139 {
		t.Fatalf("got %v want ~138", got)
	}
}

func TestCompareRuns_p95Pass(t *testing.T) {
	b := sampleArtifact("s", 100, "m")
	c := sampleArtifact("s", 108, "m")
	max := 20.0
	result, err := CompareRuns(b, c, &max, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Passed || !result.P95.Passed {
		t.Fatalf("result: %+v", result)
	}
}

func TestCompareRuns_p95Fail(t *testing.T) {
	b := sampleArtifact("s", 100, "m")
	c := sampleArtifact("s", 150, "m")
	max := 20.0
	result, err := CompareRuns(b, c, &max, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Passed || result.P95.Passed {
		t.Fatalf("result: %+v", result)
	}
	if !result.P95.Driver.Found || result.P95.Driver.CaseID != "c1" {
		t.Fatalf("driver: %+v", result.P95.Driver)
	}
}

func TestFindP95Driver_multiCase(t *testing.T) {
	baseline := RunArtifact{
		Version: RunArtifactVersion,
		SuiteID: "glm-qwen-smoke",
		Cases: []RunCase{
			{ID: "intent-glm", Model: "z-ai/glm-5.2", LatencyMs: 1214},
			{ID: "intent-qwen", Model: "qwen/qwen3.7-plus", LatencyMs: 5066},
		},
		Summary: RunSummary{P95Ms: 5066},
	}
	candidate := RunArtifact{
		Version: RunArtifactVersion,
		SuiteID: "glm-qwen-smoke",
		Cases: []RunCase{
			{ID: "intent-glm", Model: "z-ai/glm-5.2", LatencyMs: 1090},
			{ID: "intent-qwen", Model: "qwen/qwen3.7-plus", LatencyMs: 18717},
		},
		Summary: RunSummary{P95Ms: 18717},
	}

	driver := FindP95Driver(baseline, candidate)
	if driver.CaseID != "intent-qwen" || driver.Model != "qwen/qwen3.7-plus" {
		t.Fatalf("driver: %+v", driver)
	}
	if driver.BaselineMs != 5066 || driver.CandidateMs != 18717 {
		t.Fatalf("latency: %+v", driver)
	}
	if driver.DeltaPercent < 268 || driver.DeltaPercent > 270 {
		t.Fatalf("delta: %v want ~269", driver.DeltaPercent)
	}
}

func TestCompareRuns_skipGateWhenNil(t *testing.T) {
	b := sampleArtifact("s", 100, "m")
	c := sampleArtifact("s", 500, "m")
	result, err := CompareRuns(b, c, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Passed || result.P95.Checked || result.PassRate.Checked {
		t.Fatalf("result: %+v", result)
	}
}

func TestCompareRuns_passRatePass(t *testing.T) {
	b := sampleArtifactWithPassRate("s", 100, "m", 100)
	c := sampleArtifactWithPassRate("s", 100, "m", 97)
	min := 95.0
	result, err := CompareRuns(b, c, nil, &min)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Passed || !result.PassRate.Passed || !result.PassRate.Checked {
		t.Fatalf("result: %+v", result)
	}
	if result.PassRate.CandidatePassRate != 97 {
		t.Fatalf("pass rate: %+v", result.PassRate)
	}
}

func TestCompareRuns_passRateFail(t *testing.T) {
	b := sampleArtifactWithPassRate("s", 100, "m", 100)
	c := sampleArtifactWithPassRate("s", 100, "m", 89)
	min := 95.0
	result, err := CompareRuns(b, c, nil, &min)
	if err != nil {
		t.Fatal(err)
	}
	if result.Passed || result.PassRate.Passed {
		t.Fatalf("result: %+v", result)
	}
	if result.PassRate.Threshold != 95 || result.PassRate.CandidatePassRate != 89 {
		t.Fatalf("pass rate: %+v", result.PassRate)
	}
}

func TestCompareRuns_passRateBoundary(t *testing.T) {
	b := sampleArtifactWithPassRate("s", 100, "m", 100)
	c := sampleArtifactWithPassRate("s", 100, "m", 95)
	min := 95.0
	result, err := CompareRuns(b, c, nil, &min)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Passed || !result.PassRate.Passed {
		t.Fatalf("exact 95.0 must PASS: %+v", result)
	}
}

func TestCompareRuns_p95OkPassRateFail(t *testing.T) {
	b := sampleArtifactWithPassRate("s", 100, "m", 100)
	c := sampleArtifactWithPassRate("s", 108, "m", 89)
	max := 20.0
	min := 95.0
	result, err := CompareRuns(b, c, &max, &min)
	if err != nil {
		t.Fatal(err)
	}
	if !result.P95.Passed {
		t.Fatalf("p95 should pass: %+v", result.P95)
	}
	if result.Passed || result.PassRate.Passed {
		t.Fatalf("AND fail expected: %+v", result)
	}
}

func TestParseRunArtifactJSON_version(t *testing.T) {
	_, err := ParseRunArtifactJSON([]byte(`{"version":99,"suite_id":"x","cases":[]}`))
	if err == nil {
		t.Fatal("expected version error")
	}
}

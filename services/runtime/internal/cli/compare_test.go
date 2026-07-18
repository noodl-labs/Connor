package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/noodl-labs/ConnorLLM/services/runtime/internal/runtime/domain/entities"
)

func TestPrintCompareResult_pass(t *testing.T) {
	var buf bytes.Buffer
	printCompareResult(&buf, entities.CompareResult{
		Passed: true,
		P95: entities.P95CompareResult{
			Checked: true, Passed: true, DeltaPercent: 8, Threshold: 20,
		},
	})
	if !strings.Contains(buf.String(), "PASS  p95 +8%") {
		t.Fatalf("got %q", buf.String())
	}
}

func TestPrintCompareResult_fail(t *testing.T) {
	var buf bytes.Buffer
	printCompareResult(&buf, entities.CompareResult{
		Passed: false,
		P95: entities.P95CompareResult{
			Checked: true, Passed: false, DeltaPercent: 87, Threshold: 20,
			Driver: entities.P95Driver{
				Found: true, CaseID: "intent-qwen", Model: "qwen/qwen3.7-plus",
				BaselineMs: 5066, CandidateMs: 18717, DeltaPercent: 269,
			},
		},
	})
	out := buf.String()
	if !strings.Contains(out, "FAIL  p95 +87%") {
		t.Fatalf("got %q", out)
	}
	if !strings.Contains(out, "driver  intent-qwen  qwen/qwen3.7-plus  5066ms → 18717ms") {
		t.Fatalf("got %q", out)
	}
}

func TestPrintCompareResult_passRateFail(t *testing.T) {
	var buf bytes.Buffer
	printCompareResult(&buf, entities.CompareResult{
		Passed: false,
		P95: entities.P95CompareResult{
			Checked: true, Passed: true, DeltaPercent: 8, Threshold: 20,
		},
		PassRate: entities.PassRateCompareResult{
			Checked: true, Passed: false, CandidatePassRate: 89, Threshold: 95,
		},
	})
	out := buf.String()
	if !strings.Contains(out, "PASS  p95 +8%") {
		t.Fatalf("got %q", out)
	}
	if !strings.Contains(out, "FAIL  pass rate 89%  (threshold: 95%)") {
		t.Fatalf("got %q", out)
	}
}

func TestPrintCompareResult_passRatePass(t *testing.T) {
	var buf bytes.Buffer
	printCompareResult(&buf, entities.CompareResult{
		Passed: true,
		P95: entities.P95CompareResult{
			Checked: true, Passed: true, DeltaPercent: 8, Threshold: 20,
		},
		PassRate: entities.PassRateCompareResult{
			Checked: true, Passed: true, CandidatePassRate: 97, Threshold: 95,
		},
	})
	out := buf.String()
	if !strings.Contains(out, "PASS  pass rate 97%") {
		t.Fatalf("got %q", out)
	}
}

func TestLoadRunArtifact_roundTrip(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/run.json"
	artifact := entities.RunArtifact{
		Version: entities.RunArtifactVersion,
		SuiteID: "serving-smoke",
		Cases: []entities.RunCase{
			{ID: "a", Model: "m", Passed: true, LatencyMs: 100, Attempts: 1},
		},
		Summary: entities.RunSummary{Total: 1, Passed: 1, P95Ms: 100, P50Ms: 100, PassRate: 100},
	}
	if err := writeRunArtifact(path, artifact); err != nil {
		t.Fatal(err)
	}
	loaded, err := loadRunArtifact(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.SuiteID != "serving-smoke" || loaded.Summary.P95Ms != 100 {
		t.Fatalf("loaded: %+v", loaded)
	}
}

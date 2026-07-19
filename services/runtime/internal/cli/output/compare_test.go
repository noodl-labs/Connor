package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/noodl-labs/ConnorLLM/services/runtime/internal/runtime/domain/entities"
)

func TestPrintCompare_pass(t *testing.T) {
	var buf bytes.Buffer
	PrintCompare(&buf, CompareView{
		Version: "v0.1.0",
		SuiteID: "regression-demo",
		Baseline: entities.RunSummary{P95Ms: 160, PassRate: 100},
		Candidate: entities.RunSummary{P95Ms: 170, PassRate: 100},
		Result: entities.CompareResult{
			Passed: true,
			P95: entities.P95CompareResult{
				Checked: true, Passed: true, DeltaPercent: 8, Threshold: 20,
			},
			PassRate: entities.PassRateCompareResult{
				Checked: true, Passed: true, CandidatePassRate: 100, Threshold: 95,
			},
		},
	})
	out := buf.String()
	if !strings.Contains(out, "PASS  p95 +8%") {
		t.Fatalf("got %q", out)
	}
	if !strings.Contains(out, "PASS  pass rate 100%") {
		t.Fatalf("got %q", out)
	}
	if !strings.Contains(out, "GATE PASSED — safe to merge") {
		t.Fatalf("got %q", out)
	}
	if strings.Contains(out, "\x1b[") {
		t.Fatalf("unexpected ANSI in buffer: %q", out)
	}
}

func TestPrintCompare_fail(t *testing.T) {
	var buf bytes.Buffer
	PrintCompare(&buf, CompareView{
		Version: "v0.1.0",
		SuiteID: "regression-demo",
		Baseline: entities.RunSummary{P95Ms: 160, PassRate: 100},
		Candidate: entities.RunSummary{P95Ms: 400, PassRate: 75},
		Result: entities.CompareResult{
			Passed: false,
			P95: entities.P95CompareResult{
				Checked: true, Passed: false, DeltaPercent: 150, Threshold: 20,
				Driver: entities.P95Driver{
					Found: true, CaseID: "d", Model: "m",
					BaselineMs: 160, CandidateMs: 400, DeltaPercent: 150,
				},
			},
			PassRate: entities.PassRateCompareResult{
				Checked: true, Passed: false, CandidatePassRate: 75, Threshold: 95,
			},
		},
	})
	out := buf.String()
	if !strings.Contains(out, "FAIL  p95 +150%") {
		t.Fatalf("got %q", out)
	}
	if !strings.Contains(out, "driver  d  m  160ms → 400ms") {
		t.Fatalf("got %q", out)
	}
	if !strings.Contains(out, "FAIL  pass rate 75%  (threshold: 95%)") {
		t.Fatalf("got %q", out)
	}
	if !strings.Contains(out, "GATE FAILED — do not merge") {
		t.Fatalf("got %q", out)
	}
}

func TestPrintCompare_p95SkipNoThreshold(t *testing.T) {
	var buf bytes.Buffer
	PrintCompare(&buf, CompareView{
		Result: entities.CompareResult{
			Passed: true,
			P95: entities.P95CompareResult{
				Checked: false, DeltaPercent: 12,
			},
		},
	})
	out := buf.String()
	if !strings.Contains(out, "PASS  p95 +12%") || !strings.Contains(out, "no threshold set") {
		t.Fatalf("got %q", out)
	}
}

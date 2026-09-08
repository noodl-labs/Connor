package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/noodl-labs/ConnorLLM/services/runtime/internal/runtime/domain/entities"
)

func TestPrintInspect_trajectory(t *testing.T) {
	var buf bytes.Buffer
	PrintInspect(&buf, InspectView{
		Version: "v0.1.0",
		Report: entities.InspectReport{
			SuiteID: "test-agent",
			Cases: []entities.InspectCase{{
				ID:         "test-agent",
				Passed:     true,
				RunID:      "run_a1b2c3d4e5f60718",
				DurationMs: 40,
				ToolCalls:  2,
				Tree: []entities.InspectNode{
					{Name: "search_customer", Kind: entities.SpanKindTool, Status: entities.SpanStatusOK, DurationMs: 10},
					{Name: "refund", Kind: entities.SpanKindTool, Status: entities.SpanStatusOK, DurationMs: 15},
				},
			}},
		},
	})
	out := buf.String()
	for _, want := range []string{
		"Connor  v0.1.0",
		"RUN     run_a1b2c3d4e5f60718    test-agent",
		"Status       PASSED",
		"Duration     40ms",
		"Tool calls   2",
		"Trajectory",
		"search_customer",
		"refund",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in %q", want, out)
		}
	}
	if strings.Contains(out, "GATE") || strings.Contains(out, "\x1b[") {
		t.Fatalf("unexpected gate/ANSI: %q", out)
	}
}

func TestPrintInspect_httpOnly(t *testing.T) {
	var buf bytes.Buffer
	PrintInspect(&buf, InspectView{
		Version: "v0.1.0",
		Report: entities.InspectReport{
			SuiteID: "regression-demo",
			Cases: []entities.InspectCase{
				{ID: "a", Model: "m", Passed: true, LatencyMs: 100, HTTPOnly: true},
				{ID: "b", Model: "m", Passed: false, Reason: "content_mismatch", LatencyMs: 80, HTTPOnly: true},
			},
		},
	})
	out := buf.String()
	if !strings.Contains(out, "HTTP-only artifact — no trajectory") {
		t.Fatalf("missing note: %q", out)
	}
	if !strings.Contains(out, "content_mismatch") {
		t.Fatalf("missing reason: %q", out)
	}
	if strings.Contains(out, "Trajectory") {
		t.Fatalf("HTTP-only should not print Trajectory: %q", out)
	}
}

func TestPrintInspect_errorAndDuplicate(t *testing.T) {
	var buf bytes.Buffer
	PrintInspect(&buf, InspectView{
		Version: "v0.1.0",
		Report: entities.InspectReport{
			SuiteID: "support-agent",
			Cases: []entities.InspectCase{{
				Passed:     false,
				RunID:      "run_x",
				DurationMs: 8200,
				ToolCalls:  3,
				Tree: []entities.InspectNode{
					{Name: "refund", Status: entities.SpanStatusError, DurationMs: 5000, Error: "timeout", Kind: entities.SpanKindTool, CallIndex: 1},
					{Name: "refund", Status: entities.SpanStatusOK, DurationMs: 200, Duplicate: true, CallIndex: 2, Kind: entities.SpanKindTool},
					{Name: "send_email", Status: entities.SpanStatusError, DurationMs: 0, Error: "ConnectionError", Kind: entities.SpanKindTool, CallIndex: 1},
				},
			}},
		},
	})
	out := buf.String()
	if !strings.Contains(out, "Status       FAILED") {
		t.Fatalf("status: %q", out)
	}
	if !strings.Contains(out, "Duration     8.2s") {
		t.Fatalf("duration: %q", out)
	}
	if !strings.Contains(out, "└─  error  timeout") {
		t.Fatalf("error line: %q", out)
	}
	if !strings.Contains(out, "duplicate name (2nd call)") {
		t.Fatalf("duplicate: %q", out)
	}
}

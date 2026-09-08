package entities

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestSpanValidate_ok(t *testing.T) {
	s := validSpan()
	if err := s.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestSpanValidate_missingSpanID(t *testing.T) {
	s := validSpan()
	s.SpanID = ""
	if err := s.Validate(); !errors.Is(err, ErrInvalidSpan) {
		t.Fatalf("got %v want ErrInvalidSpan", err)
	}
}

func TestTrajectoryValidate_runIDMismatch(t *testing.T) {
	s := validSpan()
	s.RunID = "run_other"
	tr := Trajectory{
		RunID: "run_aaa",
		Spans: []Span{s},
	}
	if err := tr.Validate(); !errors.Is(err, ErrInvalidTrajectory) {
		t.Fatalf("got %v", err)
	}
}

func TestTrajectoryValidate_ok(t *testing.T) {
	tr := Trajectory{RunID: "run_aaa", Spans: []Span{validSpan()}}
	if err := tr.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestParseRunArtifactJSON_trajectoryGolden(t *testing.T) {
	data, err := os.ReadFile(findRepoFile(t, "benchmarks/examples/agent-trace-demo/run.json"))
	if err != nil {
		t.Fatal(err)
	}
	art, err := ParseRunArtifactJSON(data)
	if err != nil {
		t.Fatal(err)
	}
	if art.Version != RunArtifactVersion {
		t.Fatalf("version: %d", art.Version)
	}
	if len(art.Cases) != 1 {
		t.Fatalf("cases: %d", len(art.Cases))
	}
	tr := art.Cases[0].Trajectory
	if tr == nil {
		t.Fatal("expected trajectory")
	}
	if err := tr.Validate(); err != nil {
		t.Fatal(err)
	}
	if tr.RunID != "run_a1b2c3d4e5f60718" {
		t.Fatalf("run_id: %q", tr.RunID)
	}
	if len(tr.Spans) != 3 {
		t.Fatalf("spans: %d want 3", len(tr.Spans))
	}
	if tr.Spans[0].Kind != SpanKindRun || tr.Spans[0].ParentSpanID != "" {
		t.Fatalf("root: %+v", tr.Spans[0])
	}
	if tr.Spans[1].Name != "search_customer" || tr.Spans[1].ParentSpanID != tr.Spans[0].SpanID {
		t.Fatalf("search: %+v", tr.Spans[1])
	}
	if tr.Spans[2].Name != "refund" || tr.Spans[2].Kind != SpanKindTool {
		t.Fatalf("refund: %+v", tr.Spans[2])
	}
	if art.Cases[0].LatencyMs != tr.Spans[0].DurationMs {
		t.Fatalf("latency_ms %d != root duration %d", art.Cases[0].LatencyMs, tr.Spans[0].DurationMs)
	}
}

func TestBuildRunArtifact_omitsTrajectory(t *testing.T) {
	art, err := BuildRunArtifact("s", "https://gw/v1",
		[]RunCaseMeta{{ID: "a", Model: "m"}},
		[]CaseResult{{CaseID: "a", Passed: true, Response: Response{LatencyMs: 10, Attempts: 1}}},
	)
	if err != nil {
		t.Fatal(err)
	}
	if art.Cases[0].Trajectory != nil {
		t.Fatal("HTTP export must not set trajectory")
	}
}

func validSpan() Span {
	return Span{
		TraceID:    "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		RunID:      "run_aaa",
		SpanID:     "bbbbbbbbbbbbbbbb",
		Name:       "test-agent",
		Kind:       SpanKindRun,
		Status:     SpanStatusOK,
		StartedAt:  "2026-09-08T19:00:00.000Z",
		EndedAt:    "2026-09-08T19:00:00.050Z",
		DurationMs: 50,
	}
}

func findRepoFile(t *testing.T, rel string) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		p := filepath.Join(dir, rel)
		if _, err := os.Stat(p); err == nil {
			return p
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("repo file not found: %s", rel)
		}
		dir = parent
	}
}

package entities

import "testing"

func TestBuildRunArtifact_passRateAndPercentiles(t *testing.T) {
	metas := []RunCaseMeta{
		{ID: "a", Model: "m1"},
		{ID: "b", Model: "m2"},
		{ID: "c", Model: "m3"},
	}
	results := []CaseResult{
		{CaseID: "a", Passed: true, Response: Response{LatencyMs: 100, Attempts: 1}},
		{CaseID: "b", Passed: false, Reason: FailReasonCallFailed, Response: Response{LatencyMs: 300, Attempts: 3}},
		{CaseID: "c", Passed: true, Response: Response{LatencyMs: 200, Attempts: 1}},
	}

	art, err := BuildRunArtifact("serving-smoke", "https://gw/v1", metas, results)
	if err != nil {
		t.Fatal(err)
	}

	if art.Version != RunArtifactVersion {
		t.Fatalf("version: got %d want %d", art.Version, RunArtifactVersion)
	}
	if art.SuiteID != "serving-smoke" {
		t.Fatalf("suite_id: got %q", art.SuiteID)
	}
	if art.Summary.Total != 3 || art.Summary.Passed != 2 || art.Summary.Failed != 1 {
		t.Fatalf("summary counts: got total=%d passed=%d failed=%d", art.Summary.Total, art.Summary.Passed, art.Summary.Failed)
	}
	if art.Summary.PassRate < 66.6 || art.Summary.PassRate > 66.7 {
		t.Fatalf("pass_rate: got %v want ~66.7", art.Summary.PassRate)
	}
	if art.Summary.P50Ms != 200 {
		t.Fatalf("p50_ms: got %d want 200", art.Summary.P50Ms)
	}
	if art.Summary.P95Ms != 300 {
		t.Fatalf("p95_ms: got %d want 300", art.Summary.P95Ms)
	}

	for i, c := range art.Cases {
		if c.ID == "" || c.Model == "" {
			t.Fatalf("case %d missing id or model", i)
		}
	}
	if art.Cases[1].Reason != "call_failed" {
		t.Fatalf("reason: got %q want call_failed", art.Cases[1].Reason)
	}
	if art.Cases[0].Reason != "" {
		t.Fatalf("passed case reason should be empty, got %q", art.Cases[0].Reason)
	}
}

func TestBuildRunArtifact_toolCallsAndTokens(t *testing.T) {
	metas := []RunCaseMeta{
		{ID: "a", Model: "m"},
		{ID: "b", Model: "m"},
	}
	results := []CaseResult{
		{
			CaseID: "a", Passed: true,
			Response: Response{
				LatencyMs: 10, Attempts: 1,
				ToolCalls:        []ToolCall{{Name: "search"}},
				PromptTokens:     100,
				CompletionTokens: 20,
			},
		},
		{
			CaseID: "b", Passed: true,
			Response: Response{
				LatencyMs: 20, Attempts: 1,
				ToolCalls:        []ToolCall{{Name: "search"}, {Name: "book"}},
				PromptTokens:     50,
				CompletionTokens: 10,
			},
		},
	}

	art, err := BuildRunArtifact("agent-support", "https://gw/v1", metas, results)
	if err != nil {
		t.Fatal(err)
	}
	if len(art.Cases[0].ToolCalls) != 1 || art.Cases[0].ToolCalls[0].Name != "search" {
		t.Fatalf("case0 tools: %+v", art.Cases[0].ToolCalls)
	}
	if len(art.Cases[1].ToolCalls) != 2 {
		t.Fatalf("case1 tools: %+v", art.Cases[1].ToolCalls)
	}
	if art.Cases[0].PromptTokens != 100 || art.Cases[0].CompletionTokens != 20 {
		t.Fatalf("case0 tokens: %+v", art.Cases[0])
	}
	if art.Summary.PromptTokens != 150 || art.Summary.CompletionTokens != 30 {
		t.Fatalf("summary tokens: %+v", art.Summary)
	}
	if art.Summary.TotalTokens != 180 {
		t.Fatalf("total_tokens: got %d want 180", art.Summary.TotalTokens)
	}
}

func TestBuildRunArtifact_metaMismatch(t *testing.T) {
	_, err := BuildRunArtifact("s", "t",
		[]RunCaseMeta{{ID: "a", Model: "m"}},
		[]CaseResult{},
	)
	if err != ErrRunMetaMismatch {
		t.Fatalf("got %v want ErrRunMetaMismatch", err)
	}
}

func TestBuildRunArtifact_caseIDMismatch(t *testing.T) {
	_, err := BuildRunArtifact("s", "t",
		[]RunCaseMeta{{ID: "a", Model: "m"}},
		[]CaseResult{{CaseID: "b", Passed: true, Response: Response{LatencyMs: 1, Attempts: 1}}},
	)
	if err == nil {
		t.Fatal("expected error")
	}
}

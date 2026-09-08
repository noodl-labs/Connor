package entities

import (
	"os"
	"testing"
)

func TestInspectViewFromArtifact_golden(t *testing.T) {
	data, err := os.ReadFile(findRepoFile(t, "benchmarks/examples/agent-trace-demo/run.json"))
	if err != nil {
		t.Fatal(err)
	}
	art, err := ParseRunArtifactJSON(data)
	if err != nil {
		t.Fatal(err)
	}
	rep := InspectViewFromArtifact(art)
	if rep.SuiteID != "test-agent" {
		t.Fatalf("suite: %q", rep.SuiteID)
	}
	if len(rep.Cases) != 1 {
		t.Fatalf("cases: %d", len(rep.Cases))
	}
	c := rep.Cases[0]
	if c.HTTPOnly {
		t.Fatal("golden has trajectory")
	}
	if c.RunID != "run_a1b2c3d4e5f60718" {
		t.Fatalf("run_id: %q", c.RunID)
	}
	if c.DurationMs != 40 || c.ToolCalls != 2 || c.LLMCalls != 0 {
		t.Fatalf("stats: %+v", c)
	}
	if len(c.Tree) != 2 {
		t.Fatalf("tree len %d want 2 (root run omitted)", len(c.Tree))
	}
	if c.Tree[0].Name != "search_customer" || c.Tree[0].Depth != 0 {
		t.Fatalf("n0: %+v", c.Tree[0])
	}
	if c.Tree[1].Name != "refund" || c.Tree[1].Kind != SpanKindTool {
		t.Fatalf("n1: %+v", c.Tree[1])
	}
}

func TestInspectViewFromArtifact_httpOnly(t *testing.T) {
	art := RunArtifact{
		Version: RunArtifactVersion,
		SuiteID: "serving-smoke",
		Cases: []RunCase{
			{ID: "a", Model: "m", Passed: true, LatencyMs: 100},
			{ID: "b", Model: "m", Passed: false, Reason: "content_mismatch", LatencyMs: 80},
		},
	}
	rep := InspectViewFromArtifact(art)
	if len(rep.Cases) != 2 {
		t.Fatalf("cases: %d", len(rep.Cases))
	}
	if !rep.Cases[0].HTTPOnly || !rep.Cases[1].HTTPOnly {
		t.Fatal("expected HTTP-only")
	}
	if len(rep.Cases[0].Tree) != 0 {
		t.Fatalf("tree: %+v", rep.Cases[0].Tree)
	}
	if rep.Cases[1].Reason != "content_mismatch" {
		t.Fatalf("reason: %q", rep.Cases[1].Reason)
	}
}

func TestInspectViewFromArtifact_nestedAndDuplicate(t *testing.T) {
	root := validSpan()
	policy := validSpan()
	policy.SpanID = "cccccccccccccccc"
	policy.ParentSpanID = root.SpanID
	policy.Name = "search_policy"
	policy.Kind = SpanKindTool

	cache := validSpan()
	cache.SpanID = "dddddddddddddddd"
	cache.ParentSpanID = policy.SpanID
	cache.Name = "cache_get"
	cache.Kind = SpanKindTool
	cache.DurationMs = 2

	refund1 := validSpan()
	refund1.SpanID = "eeeeeeeeeeeeeeee"
	refund1.ParentSpanID = root.SpanID
	refund1.Name = "refund"
	refund1.Kind = SpanKindTool

	refund2 := validSpan()
	refund2.SpanID = "ffffffffffffffff"
	refund2.ParentSpanID = root.SpanID
	refund2.Name = "refund"
	refund2.Kind = SpanKindTool
	refund2.DurationMs = 200

	llm := validSpan()
	llm.SpanID = "1111111111111111"
	llm.ParentSpanID = root.SpanID
	llm.Name = "chat"
	llm.Kind = SpanKindLLM
	llm.PromptTokens = 10
	llm.CompletionTokens = 5

	art := RunArtifact{
		Version: RunArtifactVersion,
		SuiteID: "support-agent-complex",
		Cases: []RunCase{{
			ID:        "support-agent-complex",
			Model:     "support-agent-complex",
			Passed:    true,
			LatencyMs: 50,
			Trajectory: &Trajectory{
				RunID: root.RunID,
				Spans: []Span{root, policy, cache, refund1, refund2, llm},
			},
		}},
	}
	c := InspectViewFromArtifact(art).Cases[0]
	if c.ToolCalls != 4 || c.LLMCalls != 1 || c.Tokens != 15 {
		t.Fatalf("counts: tools=%d llm=%d tokens=%d", c.ToolCalls, c.LLMCalls, c.Tokens)
	}
	// display: policy (depth 0), cache (1), refund, refund, chat — root omitted
	if len(c.Tree) != 5 {
		t.Fatalf("tree: %d %+v", len(c.Tree), names(c.Tree))
	}
	if c.Tree[0].Name != "search_policy" || c.Tree[0].Depth != 0 {
		t.Fatalf("policy: %+v", c.Tree[0])
	}
	if c.Tree[1].Name != "cache_get" || c.Tree[1].Depth != 1 {
		t.Fatalf("cache: %+v", c.Tree[1])
	}
	if c.Tree[2].Name != "refund" || c.Tree[2].Duplicate || c.Tree[2].CallIndex != 1 {
		t.Fatalf("refund1: %+v", c.Tree[2])
	}
	if c.Tree[3].Name != "refund" || !c.Tree[3].Duplicate || c.Tree[3].CallIndex != 2 {
		t.Fatalf("refund2: %+v", c.Tree[3])
	}
}

func TestInspectViewFromArtifact_errorNode(t *testing.T) {
	root := validSpan()
	fail := validSpan()
	fail.SpanID = "cccccccccccccccc"
	fail.ParentSpanID = root.SpanID
	fail.Name = "send_email"
	fail.Kind = SpanKindTool
	fail.Status = SpanStatusError
	fail.Error = "ConnectionError: down"

	art := RunArtifact{
		Cases: []RunCase{{
			ID: "x", Model: "m", Passed: false, Reason: "agent_failed",
			Trajectory: &Trajectory{RunID: root.RunID, Spans: []Span{root, fail}},
		}},
	}
	n := InspectViewFromArtifact(art).Cases[0].Tree[0]
	if n.Status != SpanStatusError || n.Error == "" || n.Name != "send_email" {
		t.Fatalf("node: %+v", n)
	}
}

func names(nodes []InspectNode) []string {
	out := make([]string, len(nodes))
	for i, n := range nodes {
		out[i] = n.Name
	}
	return out
}

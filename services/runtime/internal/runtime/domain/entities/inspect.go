package entities

// InspectReport is a display-only view of a run.json (RFC 0003 §7).
// No gates: inspect without --expect never fails a case.
type InspectReport struct {
	SuiteID string
	Target  string
	Cases   []InspectCase
}

// InspectCase is one case row for connor inspect.
type InspectCase struct {
	ID        string
	Model     string
	Passed    bool
	Reason    string
	LatencyMs int64
	HTTPOnly  bool // no trajectory — HTTP connor run artifact

	RunID      string
	DurationMs int64
	LLMCalls   int
	ToolCalls  int
	Tokens     int
	Tree       []InspectNode
}

// InspectNode is one printable span after DFS by parent_span_id.
type InspectNode struct {
	Name       string
	Kind       string
	Status     string
	DurationMs int64
	Error      string
	Depth      int
	Duplicate  bool // 2nd+ tool span with the same name (RFC §7 example)
	CallIndex  int  // 1-based occurrence of this tool name in visit order
}

// InspectViewFromArtifact builds a deterministic inspect report (RFC 0003 §7).
func InspectViewFromArtifact(a RunArtifact) InspectReport {
	cases := make([]InspectCase, 0, len(a.Cases))
	for _, c := range a.Cases {
		cases = append(cases, inspectCaseFromRun(c))
	}
	return InspectReport{
		SuiteID: a.SuiteID,
		Target:  a.Target,
		Cases:   cases,
	}
}

func inspectCaseFromRun(c RunCase) InspectCase {
	out := InspectCase{
		ID:         c.ID,
		Model:      c.Model,
		Passed:     c.Passed,
		Reason:     c.Reason,
		LatencyMs:  c.LatencyMs,
		DurationMs: c.LatencyMs,
		HTTPOnly:   c.Trajectory == nil || len(c.Trajectory.Spans) == 0,
	}
	if out.HTTPOnly {
		return out
	}
	tr := c.Trajectory
	out.RunID = tr.RunID
	for _, s := range tr.Spans {
		switch s.Kind {
		case SpanKindTool:
			out.ToolCalls++
		case SpanKindLLM:
			out.LLMCalls++
		}
		out.Tokens += s.PromptTokens + s.CompletionTokens
		if s.Kind == SpanKindRun && s.ParentSpanID == "" {
			out.DurationMs = s.DurationMs
		}
	}
	out.Tree = trajectoryTree(*tr)
	return out
}

func trajectoryTree(tr Trajectory) []InspectNode {
	byID := make(map[string]struct{}, len(tr.Spans))
	for _, s := range tr.Spans {
		byID[s.SpanID] = struct{}{}
	}

	children := make(map[string][]Span)
	var roots []Span
	for _, s := range tr.Spans {
		if s.ParentSpanID == "" || !spanIDExists(byID, s.ParentSpanID) {
			roots = append(roots, s)
			continue
		}
		children[s.ParentSpanID] = append(children[s.ParentSpanID], s)
	}

	toolSeen := make(map[string]int)
	var nodes []InspectNode
	var walk func(s Span, depth int)
	walk = func(s Span, depth int) {
		callIndex := 0
		dup := false
		if s.Kind == SpanKindTool {
			toolSeen[s.Name]++
			callIndex = toolSeen[s.Name]
			dup = callIndex >= 2
		}
		nodes = append(nodes, InspectNode{
			Name:       s.Name,
			Kind:       s.Kind,
			Status:     s.Status,
			DurationMs: s.DurationMs,
			Error:      s.Error,
			Depth:      depth,
			Duplicate:  dup,
			CallIndex:  callIndex,
		})
		for _, child := range children[s.SpanID] {
			walk(child, depth+1)
		}
	}

	for _, r := range roots {
		if r.Kind == SpanKindRun {
			for _, child := range children[r.SpanID] {
				walk(child, 0)
			}
			continue
		}
		walk(r, 0)
	}
	return nodes
}

func spanIDExists(byID map[string]struct{}, id string) bool {
	_, ok := byID[id]
	return ok
}

package entities

import (
	"errors"
	"fmt"
	"math"
	"sort"
)

// RunArtifactVersion is the JSON schema version (RFC 0001 §4).
// connor compare will reject unknown versions.
const RunArtifactVersion = 1

// RunCaseMeta is input metadata from the YAML suite (id + model).
// Lives in domain only — no benchmark package import (ADR 0001).
type RunCaseMeta struct {
	ID    string
	Model string
}

// RunArtifact is the portable run.json root object.
type RunArtifact struct {
	Version int        `json:"version"`
	SuiteID string     `json:"suite_id"`
	Target  string     `json:"target"`
	Cases   []RunCase  `json:"cases"`
	Summary RunSummary `json:"summary"`
}

// RunCase is one exported case row (must include id + model for compare).
type RunCase struct {
	ID               string     `json:"id"`
	Model            string     `json:"model"`
	Passed           bool       `json:"passed"`
	Reason           string     `json:"reason"`
	LatencyMs        int64      `json:"latency_ms"`
	Attempts         int        `json:"attempts"`
	ToolCalls        []ToolCall `json:"tool_calls,omitempty"`
	PromptTokens     int        `json:"prompt_tokens,omitempty"`
	CompletionTokens int        `json:"completion_tokens,omitempty"`
}

// RunSummary holds suite-level KPIs for compare gates (RFC 0001 / 0002).
type RunSummary struct {
	Total            int     `json:"total"`
	Passed           int     `json:"passed"`
	Failed           int     `json:"failed"`
	PassRate         float64 `json:"pass_rate"`
	P50Ms            int64   `json:"p50_ms"`
	P95Ms            int64   `json:"p95_ms"`
	PromptTokens     int     `json:"prompt_tokens,omitempty"`
	CompletionTokens int     `json:"completion_tokens,omitempty"`
	TotalTokens      int     `json:"total_tokens,omitempty"`
}

var (
	ErrEmptyRunExport    = errors.New("entities: run export requires at least one case")
	ErrRunMetaMismatch   = errors.New("entities: case meta and result length mismatch")
	ErrRunCaseIDMismatch = errors.New("entities: case id does not match result case id")
)

// BuildRunArtifact maps execution results + YAML meta into run.json shape.
//
// Invariants:
//   - len(metas) == len(results)
//   - metas[i].ID == results[i].CaseID
//   - p50/p95 over ALL case latencies (passed + failed) — RFC 0001 §4
//   - pass_rate = passed / total * 100
//   - summary token totals = sum of per-case tokens — RFC 0002 §4.3
func BuildRunArtifact(
	suiteID string,
	target string,
	metas []RunCaseMeta,
	results []CaseResult,
) (RunArtifact, error) {
	if len(metas) == 0 {
		return RunArtifact{}, ErrEmptyRunExport
	}
	if len(metas) != len(results) {
		return RunArtifact{}, ErrRunMetaMismatch
	}

	cases := make([]RunCase, len(metas))
	latencies := make([]int64, len(metas))
	passed := 0
	promptTokens := 0
	completionTokens := 0

	for i := range metas {
		// Defense in depth: YAML row must align with execution result.
		if metas[i].ID != results[i].CaseID {
			return RunArtifact{}, fmt.Errorf("%w: index %d", ErrRunCaseIDMismatch, i)
		}
		if metas[i].ID == "" || metas[i].Model == "" {
			return RunArtifact{}, fmt.Errorf("entities: case %d missing id or model", i)
		}

		resp := results[i].Response
		rc := RunCase{
			ID:               metas[i].ID,
			Model:            metas[i].Model, // ADR 0001: required for future compare
			Passed:           results[i].Passed,
			LatencyMs:        resp.LatencyMs,
			Attempts:         resp.Attempts,
			ToolCalls:        resp.ToolCalls,
			PromptTokens:     resp.PromptTokens,
			CompletionTokens: resp.CompletionTokens,
		}
		if results[i].Passed {
			rc.Reason = ""
		} else {
			rc.Reason = string(results[i].Reason)
		}

		cases[i] = rc
		latencies[i] = rc.LatencyMs
		promptTokens += resp.PromptTokens
		completionTokens += resp.CompletionTokens
		if rc.Passed {
			passed++
		}
	}

	total := len(cases)
	summary := RunSummary{
		Total:            total,
		Passed:           passed,
		Failed:           total - passed,
		PassRate:         float64(passed) / float64(total) * 100,
		P50Ms:            percentileMs(latencies, 0.50),
		P95Ms:            percentileMs(latencies, 0.95),
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      promptTokens + completionTokens,
	}

	return RunArtifact{
		Version: RunArtifactVersion,
		SuiteID: suiteID,
		Target:  target,
		Cases:   cases,
		Summary: summary,
	}, nil
}

// percentileMs returns the nearest-rank percentile in milliseconds.
// Stable algorithm: sort asc, then pick index ceil(p*n)-1.
func percentileMs(values []int64, p float64) int64 {
	if len(values) == 0 {
		return 0
	}
	sorted := append([]int64(nil), values...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	// Nearest-rank: index = ceil(p * n) - 1, clamped to [0, n-1]
	rank := int(math.Ceil(p*float64(len(sorted)))) - 1
	if rank < 0 {
		rank = 0
	}
	if rank >= len(sorted) {
		rank = len(sorted) - 1
	}
	return sorted[rank]
}

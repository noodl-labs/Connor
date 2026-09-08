package entities

import (
	"errors"
	"fmt"
)

// Span kinds and statuses for agent trajectories (RFC 0003 §6.2).
const (
	SpanKindRun  = "run"
	SpanKindTool = "tool"
	SpanKindLLM  = "llm"
	SpanKindSpan = "span"

	SpanStatusOK    = "ok"
	SpanStatusError = "error"
)

var (
	ErrInvalidTrajectory = errors.New("entities: invalid trajectory")
	ErrInvalidSpan       = errors.New("entities: invalid span")
)

// Trajectory is the recorded agent execution tree for one case (RFC 0003).
type Trajectory struct {
	RunID string `json:"run_id"`
	Spans []Span `json:"spans"`
}

// Span is one node in a trajectory (OTel-shaped names, no OTel SDK — ADR 0003).
type Span struct {
	TraceID          string         `json:"trace_id"`
	RunID            string         `json:"run_id"`
	SpanID           string         `json:"span_id"`
	ParentSpanID     string         `json:"parent_span_id"`
	Name             string         `json:"name"`
	Kind             string         `json:"kind"`
	Status           string         `json:"status"`
	StartedAt        string         `json:"started_at"`
	EndedAt          string         `json:"ended_at"`
	DurationMs       int64          `json:"duration_ms"`
	Error            string         `json:"error,omitempty"`
	Input            any            `json:"input,omitempty"`
	Output           any            `json:"output,omitempty"`
	Attributes       map[string]any `json:"attributes,omitempty"`
	PromptTokens     int            `json:"prompt_tokens,omitempty"`
	CompletionTokens int            `json:"completion_tokens,omitempty"`
}

// Validate checks required trajectory fields (RFC 0003 P1).
func (t Trajectory) Validate() error {
	if t.RunID == "" {
		return fmt.Errorf("%w: run_id is required", ErrInvalidTrajectory)
	}
	if len(t.Spans) == 0 {
		return fmt.Errorf("%w: at least one span is required", ErrInvalidTrajectory)
	}
	for i, s := range t.Spans {
		if err := s.Validate(); err != nil {
			return fmt.Errorf("%w: spans[%d]: %w", ErrInvalidTrajectory, i, err)
		}
		if s.RunID != t.RunID {
			return fmt.Errorf("%w: spans[%d]: run_id mismatch", ErrInvalidTrajectory, i)
		}
	}
	return nil
}

// Validate checks required span fields.
func (s Span) Validate() error {
	if s.TraceID == "" {
		return fmt.Errorf("%w: trace_id is required", ErrInvalidSpan)
	}
	if s.RunID == "" {
		return fmt.Errorf("%w: run_id is required", ErrInvalidSpan)
	}
	if s.SpanID == "" {
		return fmt.Errorf("%w: span_id is required", ErrInvalidSpan)
	}
	if s.Name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidSpan)
	}
	if !validSpanKind(s.Kind) {
		return fmt.Errorf("%w: invalid kind %q", ErrInvalidSpan, s.Kind)
	}
	if !validSpanStatus(s.Status) {
		return fmt.Errorf("%w: invalid status %q", ErrInvalidSpan, s.Status)
	}
	if s.StartedAt == "" {
		return fmt.Errorf("%w: started_at is required", ErrInvalidSpan)
	}
	if s.EndedAt == "" {
		return fmt.Errorf("%w: ended_at is required", ErrInvalidSpan)
	}
	if s.DurationMs < 0 {
		return fmt.Errorf("%w: duration_ms must be >= 0", ErrInvalidSpan)
	}
	return nil
}

func validSpanKind(kind string) bool {
	switch kind {
	case SpanKindRun, SpanKindTool, SpanKindLLM, SpanKindSpan:
		return true
	default:
		return false
	}
}

func validSpanStatus(status string) bool {
	switch status {
	case SpanStatusOK, SpanStatusError:
		return true
	default:
		return false
	}
}

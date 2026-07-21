package entities

import (
	"errors"
	"strings"
	"unicode/utf8"
)

/*
This file defines the runtime observation returned after a provider call.
Response is a domain value object: raw outcome + timing + retry summary.
It does not perform validation (JSON), aggregation (p95), or QoS decisions.
*/

// FailureKind classifies how a provider call ultimately failed (UC2, UC3, UC9).
type FailureKind string

const (
	FailureNone           FailureKind = ""
	FailureTimeout        FailureKind = "timeout"
	FailureRetryExhausted FailureKind = "retry_exhausted"
	FailureProvider       FailureKind = "provider"
	FailureCancelled      FailureKind = "cancelled"
)

// Sentinel errors for errors.Is in tests and ExecuteCase (optional but useful).
var (
	ErrCallTimeout        = errors.New("entities: call timeout")
	ErrCallRetryExhausted = errors.New("entities: retry exhausted")
	ErrCallCancelled      = errors.New("entities: call cancelled")
)

// ToolCall is a tool invocation name from a provider response (RFC 0002).
// Args are intentionally omitted in v0.2 — name (+ order) only.
type ToolCall struct {
	Name string `json:"name"`
}

// Response is what the runtime observed after executing a Request (final attempt summary).
type Response struct {
	Body             string
	HTTPStatus       int
	LatencyMs        int64
	TTFTMs           int64 // Time to first token when Request.Stream == true; 0 if non-streaming.
	Attempts         int
	Kind             FailureKind // Type of failure
	Err              error
	ToolCalls        []ToolCall // from message.tool_calls (RFC 0002 PR-1)
	PromptTokens     int        // from usage.prompt_tokens
	CompletionTokens int        // from usage.completion_tokens
}

// --- P0: success / failure (UC1, UC7) ---

// Succeeded reports a successful provider call (2xx and no failure kind).
func (r Response) Succeeded() bool {
	return r.Kind == FailureNone && r.Err == nil && r.HTTPStatus >= 200 && r.HTTPStatus < 300
}

// Failed is the inverse of Succeeded.
func (r Response) Failed() bool {
	return !r.Succeeded()
}

// --- P0: timeout (UC2) ---

// IsTimeout reports whether the call failed due to deadline / timeout.
func (r Response) IsTimeout() bool {
	if r.Kind == FailureTimeout {
		return true
	}
	return errors.Is(r.Err, ErrCallTimeout)
}

// --- P1: retry exhausted (UC3) ---

// IsRetryExhausted reports whether all retry attempts were used without success.
func (r Response) IsRetryExhausted() bool {
	if r.Kind == FailureRetryExhausted {
		return true
	}
	return errors.Is(r.Err, ErrCallRetryExhausted)
}

// --- P2: streaming (UC4) ---

// HasTTFT reports whether a first-token time was recorded (streaming).
func (r Response) HasTTFT() bool {
	return r.TTFTMs > 0
}

// --- P2: reporting (UC7) ---

// BodyPreview returns a truncated body safe for logs and terminal output.
func (r Response) BodyPreview(maxRunes int) string {
	if maxRunes <= 0 || r.Body == "" {
		return ""
	}
	if utf8.RuneCountInString(r.Body) <= maxRunes {
		return r.Body
	}
	var b strings.Builder
	count := 0
	for _, ru := range r.Body {
		b.WriteRune(ru)
		count++
		if count >= maxRunes {
			break
		}
	}
	return b.String() + "…"
}

// --- P1: test / mock constructors (UC10) ---

// NewSuccessResponse builds a synthetic successful Response (unit tests, fake provider).
func NewSuccessResponse(body string, httpStatus int, latencyMs, ttftMs int64, attempts int) Response {
	if attempts < 1 {
		attempts = 1
	}
	if httpStatus == 0 {
		httpStatus = 200
	}
	return Response{
		Body:       body,
		HTTPStatus: httpStatus,
		LatencyMs:  latencyMs,
		TTFTMs:     ttftMs,
		Attempts:   attempts,
		Kind:       FailureNone,
		Err:        nil,
	}
}

// NewFailedResponse builds a synthetic failed Response (unit tests, fake provider).
func NewFailedResponse(kind FailureKind, httpStatus int, attempts int, latencyMs int64, err error) Response {
	if attempts < 1 {
		attempts = 1
	}
	r := Response{
		Body:       "",
		HTTPStatus: httpStatus,
		LatencyMs:  latencyMs,
		TTFTMs:     0,
		Attempts:   attempts,
		Kind:       kind,
		Err:        err,
	}
	// Ensure Err is set for errors.Is when callers pass nil.
	if r.Err == nil {
		switch kind {
		case FailureTimeout:
			r.Err = ErrCallTimeout
		case FailureRetryExhausted:
			r.Err = ErrCallRetryExhausted
		case FailureCancelled:
			r.Err = ErrCallCancelled
		case FailureProvider:
			r.Err = errors.New("entities: provider error")
		}
	}
	return r
}

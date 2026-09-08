package entities

// FailReason is benchmark-level (not HTTP FailureKind on Response).
type FailReason string

const (
	FailReasonNone            FailReason = ""
	FailReasonCallFailed      FailReason = "call_failed"     // timeout, 4xx/5xx, retry exhausted
	FailReasonInvalidJSON     FailReason = "invalid_json"    // 2xx but body fails JSON syntax check
	FailReasonSchemaMismatch  FailReason = "schema_mismatch" // 2xx JSON valid but fails expect_json_schema
	FailReasonContentMismatch FailReason = "content_mismatch"
	FailReasonAgentFailed     FailReason = "agent_failed" // trajectory root status=error (RFC 0003)
)

// CaseResult is the outcome of one benchmark case after ExecuteCase + optional checks.
type CaseResult struct {
	CaseID   string
	Response Response
	Passed   bool
	Reason   FailReason
}

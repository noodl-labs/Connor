# ADR 0003 — Connor JSON spans, not OpenTelemetry as V1 runtime

## Status
Accepted

## Context
RFC 0003 needs a trajectory model for agent CI (`inspect`, assertions, compare, later replay). OpenTelemetry is the industry default for traces. Using the OTel SDK (and OTLP) internally would make Connor look like an observability product and pull a large dependency graph into a CI gate.

Connor’s consumers in V1 are:

- a Python decorator/context manager in CI
- a Go CLI that must stay deterministic, offline-capable, and easy to fixture

They need a JSON document, not a collector.

## Decision
V1 tracing is a **small proprietary span JSON** with OpenTelemetry-*shaped* field names (`trace_id`, `span_id`, `parent_span_id`, `name`, `status`).

- No OpenTelemetry API/SDK/OTLP dependency in the Python SDK or the Go CLI.
- No requirement to run a collector in GitHub Actions.
- Mapping to OTel GenAI semantic conventions is allowed later as an **optional exporter** (Runtime / production tracing), not as the evaluation source of truth.

CI gates always read `run.json`, never a live OTLP stream.

## Consequences
- Inspect and gates stay unit-testable with golden JSON.
- We do not inherit OTel versioning, processors, or sampling.
- A future exporter can translate Connor spans → OTLP without changing fail reasons or `compare` exit codes.
- We accept some “yet another span schema” cost; it is smaller than making OTel a CI contract.

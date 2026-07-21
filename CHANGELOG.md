# Changelog

## Unreleased

### Docs
- RFC 0002 Accepted (tool & cost gates); ADR 0002 (token delta); tracking issue #9

### Unreleased (v0.2)
- Parse `tool_calls` + token `usage` from OpenAI-compatible responses into `run.json` (RFC 0002 PR-1)

---

## v0.1.0

**Theme:** Pass-rate gate + CI handbook + regression demo.

### Added
- `connor compare ... --min-pass-rate N` — fail if candidate `pass_rate` is below threshold (RFC 0001 §5.3)
- [CI regression handbook](docs/handbook/ci-regression.md) — baseline → candidate → compare in CI
- Offline [regression-demo](benchmarks/examples/regression-demo/) fixtures (p95 + pass rate FAIL)
- Themed compare output (Lipgloss, aligned with `connor run`)
- [Product vision](docs/vision.md) and [traceability](docs/traceability.md) docs
- RFC 0002 draft (tool & cost gates, v0.2)

---

## v0.1.0-beta.3

**Theme:** Regression compare (p95).

### Added
- `connor run suite.yaml --out run.json` — export run artifact (RFC 0001)
- `connor compare baseline.json candidate.json --max-p95-regression N` — p95 regression gate
- Compare FAIL output shows p95 driver case (id, model, latency delta)

---

## v0.1.0-beta.2

**Theme:** Agent output gates — text and JSON Schema.

### Added
- `expect_contains` — substring gate (`content_mismatch`)
- `expect_contains_ignore_case` — per-case or suite default
- `expect_json_schema` — inline JSON Schema validation (`schema_mismatch`)
- Example suites: `agent-json.yaml`, `agent-json-smoke.yaml`, `agent-json-compare.yaml`
- CLI: `schema ✓/✗` badge, hints for schema failures

### Dependencies
- `github.com/santhosh-tekuri/jsonschema/v6` for schema validation

---

## v0.1.0-beta.1

**Theme:** Serving smoke — "Does my endpoint respond?"

### Added
- `connor run` — single case (`--model`, `--prompt`) and YAML suites
- OpenAI-compatible provider (`CONNOR_BASE_URL`, `CONNOR_API_KEY`)
- Retry on 429 / 5xx / transient errors; per-attempt timeout
- `expect_json` — JSON syntax gate (`invalid_json`)
- `exit 0` / `exit 1` for CI
- `benchmarks/examples/serving-smoke.yaml`
- GitHub Actions CI + release binaries on tag

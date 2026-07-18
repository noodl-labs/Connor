# Changelog

## Unreleased

### Added
- `connor compare ... --min-pass-rate N` — fail if candidate `pass_rate` is below threshold (RFC 0001 §5.3)
- [CI regression handbook](docs/handbook/ci-regression.md) — baseline → candidate → compare in CI

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

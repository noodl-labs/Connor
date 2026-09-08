# ConnorLLM Roadmap

> **CI for AI Agents** — test agent behavior, tool usage, reliability, latency, and cost before shipping.

**Current release:** [v0.1.0](CHANGELOG.md#v010)  
**Status:** v0.1 shipped (HTTP smoke + JSON/schema + p95/pass-rate). RFC 0002 (HTTP tools/cost) in progress. RFC 0003 (agent tracing) in design.

**Longer-term vision:** reliability infrastructure for AI agents — **not** this MVP. See [docs/vision.md](docs/vision.md).

**North star:** [docs/vision.md](docs/vision.md) · **Agent CI RFC:** [docs/rfc/0003-agent-ci-tracing.md](docs/rfc/0003-agent-ci-tracing.md) · **Doc map:** [docs/traceability.md](docs/traceability.md)

---

## Vision

Block merges when an **LLM endpoint or AI agent** regresses — availability, structured output, **tool usage**, latency, pass rate, cost — using versioned artifacts and `exit 0` / `exit 1`.

**Today (v0.1):** Connor calls your OpenAI-compatible HTTP API (black-box).  
**Next (v0.3):** you keep your agent stack; a Python SDK records the trajectory; Connor inspects, asserts, and compares the same `run.json`.

**Product category:** AI Release Engineering (gatekeeper, not dashboard, not agent framework). See [vision.md](docs/vision.md).

**Not in this MVP:** production observability (Langfuse), orchestrating agents, MCP proxy, policy enforcement, LLM-as-judge RCA, MMLU.

---

## Release timeline

| Release | Theme | CI value |
|---------|-------|----------|
| **v0.1.0-beta.1** ✅ | Serving smoke | "Does my endpoint respond?" |
| **v0.1.0-beta.2** ✅ | Agent output gates | "Does output match the contract?" |
| **v0.1.0-beta.3** ✅ | Regression compare (p95) | "Did p95 regress vs baseline?" |
| **v0.1.0** ✅ | Pass-rate gate + handbook | Full v0.1 regression gates |
| **v0.2.0** 📋 | HTTP tool names + token cost | "Did this chat completion *request* the right tool?" |
| **v0.3.0** 📋 | Agent CI (trace → inspect → gates → compare → replay) | "What did the agent *do*, and did that regress?" |
| **v1.0.0** 📋 | Semantic eval + richer workflows | Soft judges (Python service), not a rewrite of Go gates |

Dates are indicative — ship when **exit criteria** below are met.

v0.3 is **phased** (RFC 0003). First PR is data model + Python `trace`/`@tool` only — not the whole table.

---

## Use-case matrix

| # | Use case | Question CI | Status | Example |
|---|----------|-------------|--------|---------|
| | **Availability & serving** | | | |
| 1 | Post-deploy smoke | Endpoint responds? | ✅ beta.1 | `serving-smoke.yaml` |
| 2 | Multi-model | All routed models OK? | ✅ beta.1 | 3 models in one suite |
| 3 | Single model gate | Prod model responds? | ✅ beta.1 | `connor run --model ...` |
| 4 | Staging vs prod | Staging gateway works? | ✅ beta.1 | Change `CONNOR_BASE_URL` |
| 5 | vLLM / LiteLLM local | Self-hosted server OK? | ✅ beta.1 | `CONNOR_BASE_URL=http://localhost:8000/v1` |
| 6 | Bad model (404) | Broken config detected? | ✅ beta.1 | Invalid slug → `call_failed` |
| 7 | Timeout | Latency within budget? | ✅ beta.1 | `--timeout-ms` |
| 8 | Retry / transient | 429/5xx handled? | ✅ beta.1 | Retry policy |
| | **Output quality** | | | |
| 9 | Structured JSON | Output is JSON? | ✅ beta.1 | `expect_json` |
| 10 | Block prose | Prose fails CI? | ✅ beta.1 | `bad-json-should-fail` in `agent-json.yaml` |
| 11 | Exact content | "pong" not "Tabletennis"? | ✅ beta.2 | `expect_contains` |
| 12 | JSON Schema | Required fields present? | ✅ beta.2 | `expect_json_schema` |
| | **Regression & budget** | | | |
| 13 | Latency regression | p95 vs baseline? | ✅ beta.3 | `connor compare` |
| 14 | Pass rate | Success rate ≥ threshold? | ✅ v0.1 | `--min-pass-rate` |
| 15 | Token cost | API budget exceeded? | 📋 v0.2 | `max_cost_regression` (tokens, ADR 0002) |
| | **Agent & tools (HTTP message)** | | | |
| 16 | Tool name present | Model requested `search`? | 📋 v0.2 | `expect_tool` (RFC 0002) |
| 17 | Tool name order | Requested names in order? | 📋 v0.2 | `expect_tool_calls` |
| 18 | Custom agent HTTP | Non-`/chat/completions` URL? | 📋 v0.2 | Agent provider (may slip) |
| | **Agent CI (executed trajectory)** | | | |
| 26 | Trace primitive | Record tool/LLM spans in CI? | 📋 v0.3 P1 | Python `trace` / `@tool` |
| 27 | Inspect | Explain a run without an LLM? | 📋 v0.3 P2 | `connor inspect` |
| 28 | Trajectory assertions | `refund` at most once? | 📋 v0.3 P3 | `inspect --expect` |
| 29 | Tool-volume regression | Calls/run exploded vs main? | 📋 v0.3 P4 | `compare --max-tool-calls-regression` |
| 20 | Replay (narrow) | Re-run with recorded tool outputs? | 📋 v0.3 P5 | `connor replay` — not prod time-travel |
| | **Workflow & semantic** | | | |
| 19 | Multi-step YAML workflows | Chained HTTP scenarios? | 📋 v1 | Still not an orchestrator |
| 21 | Semantic similarity | "Close enough" answer? | 📋 v1 | Python eval service |
| 22 | Groundedness | Answer anchored in docs? | 📋 v1 | Python eval service |
| | **DX & CI** | | | |
| 23 | Exit code CI | GitHub Actions PASS/FAIL? | ✅ beta.1 | `echo $?` |
| 24 | JSON artifact | Store results for compare? | ✅ v0.1 | `--out run.json` |
| 25 | Prompt A vs B | New prompt regresses? | 📋 v1 | Prompt diff |

---

## Shipped — v0.1.0-beta.1

### Execution Engine
- [x] `connor run --model --prompt` (single case)
- [x] `connor run suite.yaml` (multi-case)
- [x] OpenAI-compatible provider (`CONNOR_BASE_URL`, `CONNOR_API_KEY`)
- [x] Per-attempt timeout, retry on 429 / 5xx / transient network
- [x] Sequential suite execution

### Evaluation Engine
- [x] HTTP 2xx success check
- [x] JSON syntax gate (`expect_json`)

### Quality Gates & DX
- [x] `exit 0` / `exit 1`, fail reasons: `call_failed`, `invalid_json`
- [x] YAML parser, `serving-smoke.yaml`, CLI output

---

## Shipped — v0.1.0-beta.2

### Evaluation Engine
- [x] `expect_contains` + `expect_contains_ignore_case`
- [x] `expect_json_schema` (inline JSON Schema)
- [x] Fail reasons: `content_mismatch`, `schema_mismatch`

### Developer Experience
- [x] `agent-json.yaml`, `agent-json-smoke.yaml`, `agent-json-compare.yaml`
- [x] CLI schema badge and hints
- [x] [CHANGELOG.md](CHANGELOG.md), [docs/architecture.md](docs/architecture.md)

---

## Shipped — v0.1.0

**Theme:** Pass-rate gate + CI handbook + regression demo

### Quality Gates
- [x] `--min-pass-rate` on compare (RFC 0001 §5.3)
- [x] Themed compare output (Lipgloss)

### Developer Experience
- [x] [docs/handbook/ci-regression.md](docs/handbook/ci-regression.md)
- [x] [benchmarks/examples/regression-demo/](benchmarks/examples/regression-demo/)
- [x] [docs/vision.md](docs/vision.md), [docs/traceability.md](docs/traceability.md)
- [ ] LICENSE

### Exit criteria
- [x] All v0.1 RFC 0001 §8 items
- [x] Tag `v0.1.0` published

---

## Shipped — toward v0.1.0 (beta → v0.1)

**Theme:** Regression testing & budget gates

### Benchmark Engine
- [x] Export `run.json` after suite run
- [x] `connor compare baseline.json candidate.json`
- [x] Suite summary: p50 / p95 latency

### Quality Gates
- [x] `max_p95_regression` threshold
- [x] `min_pass_rate` threshold
- [x] `connor compare` exits 1 on gate failure

### Developer Experience
- [x] README + architecture + `.env.example`
- [x] Handbook (was `docs/ci-github-actions.md`)

---

## Planned — v0.2.0

**Theme:** Tool calls + cost gates (L3)  
**Tracking:** [#9](https://github.com/noodl-labs/Connor/issues/9) · [RFC 0002](docs/rfc/0002-tool-and-cost-gates.md) (Accepted) · [ADR 0002](docs/adr/0002-cost-regression-tokens.md)

### Evaluation & Execution
- [ ] Parse `tool_calls` from API response (PR-1)
- [ ] `expect_tool` / `expect_tool_calls` (name, order) (PR-2)
- [ ] Token usage in run.json + `--max-cost-regression` (PR-3)
- [ ] `benchmarks/examples/agent-support.yaml` + handbook (PR-4)
- [ ] Agent HTTP provider (custom URL) (PR-5, may slip to 0.2.1)

### Exit criteria for v0.2.0
- [ ] Wrong tool → `connor run` exit 1 (`tool_mismatch` / `tool_order_mismatch`)
- [ ] Cost regression demo with `compare --max-cost-regression`
- [ ] Example suite + CHANGELOG + ci-gates rule
- [ ] Tag `v0.2.0`

---

## Planned — v0.3.0

**Theme:** CI for AI Agents (white-box trajectory)  
**Tracking:** [RFC 0003](docs/rfc/0003-agent-ci-tracing.md) (Draft) · [ADR 0003](docs/adr/0003-trace-model-not-otel.md) · [ADR 0004](docs/adr/0004-run-artifact-additive-trajectory.md)

Does **not** replace v0.2 HTTP `tool_calls`. Complementary surface: Python SDK writes `run.json`; Go CLI stays the CI contract.

| Phase | User-visible | First-PR sized? |
|-------|----------------|-----------------|
| **P1** | `with trace():` + `@tool` → `run.json` trajectory | **Yes — start here** |
| **P2** | `connor inspect run.json` | After P1 |
| **P3** | `inspect --expect` trajectory gates | After P2 |
| **P4** | `compare --max-tool-calls-regression` | After P1 (needs summary counters) |
| **P5** | `connor replay` (tool stubs only) | After P1 + P4 |

### Exit criteria for v0.3.0
- [ ] Sync + async `@tool` spans correlated; exceptions re-raised
- [ ] `connor inspect` renders trajectory from `run.json` (no LLM)
- [ ] Forbidden / max tool-call gates fail `connor` with stable reasons
- [ ] Compare AND-composes tool-call regression with p95 / pass-rate
- [ ] Replay guarantees documented; HTTP-only artifact → exit 2
- [ ] Existing HTTP suites unchanged
- [ ] Tag `v0.3.0` (may ship as 0.3.0-alpha after P1–P2)

### First PR (P1 only — do not bundle P2–P5)

**Title:** `feat: agent trajectory data model and Python trace/@tool`

**Scope:** Go `entities.Span` / `Trajectory` (JSON tags only) + `sdk/python/connor` (`trace`, `@tool`, serialize to version-1 `run.json`) + golden fixture + tests.

**Non-goals:** `inspect`, YAML expect, compare flags, replay, OTel, framework plugins.

Design: [RFC 0003](docs/rfc/0003-agent-ci-tracing.md). Spans: [ADR 0003](docs/adr/0003-trace-model-not-otel.md). Artifact: [ADR 0004](docs/adr/0004-run-artifact-additive-trajectory.md).

---

#### First PR — detailed contract

**Problem:** Connor can gate a single HTTP response, not an agent loop. We need a serializable trajectory before any new CLI command.

**Public API (Python):**

```python
from connor import trace, tool

@tool
async def search_customer(customer_id: str): ...

@tool
async def refund(payment_id: str): ...

with trace("test-agent") as run:
    await search_customer("123")
    await refund("payment-456")
    run.write("run.json")
```

Must produce a tree `run → search_customer, refund` with stable `run_id`, parent/child span ids, `duration_ms`, `status`, original exceptions preserved.

**Tests:** sync + async; nested tools; concurrent tools (sibling spans); exception re-raise; unserializable I/O; golden JSON vs Go unmarshal.

---

## Planned — v1.0.0

**Theme:** Soft evaluation + release engineering depth (not a rewrite)

- [ ] `services/evaluation/` (Python): semantic similarity, groundedness — **feeds** `run.json`, does not replace Go gates
- [ ] Reliability score with explicit N/A dimensions
- [ ] Prompt diff, richer driver taxonomy on compare FAIL
- Replay of **executed** tools ships in v0.3 (narrow). v1 may add cassette/LLM replay — only if P5 proves useful.

---

## Six engines (status)

| Engine | Today | Target |
|--------|-------|--------|
| Execution | HTTP provider, retry, timeout | **KEEP.** Agents are not executed by Connor |
| Evaluation | JSON, schema, contains (Go) | + trajectory gates (Go); semantic eval (Python, v1) |
| Benchmark | YAML suites + `compare` | + tool-call / token regression |
| Quality Gates | `exit 0/1/2` | Same contract; more optional flags |
| Observability | `run.json` (HTTP) | + optional `trajectory` on the same file (no store) |
| Developer Experience | CLI, YAML, handbook | + Python SDK (`trace`, `@tool`); `inspect` |

Details: [docs/architecture.md](docs/architecture.md)

---

## Eight features (status)

| # | Feature | Status | Release |
|---|---------|--------|---------|
| 1 | Regression testing | compare p95 + pass rate | v0.1 ✅ |
| 2 | Tool call verification | HTTP names: v0.2; executed spans: v0.3 | v0.2 / v0.3 |
| 3 | Reliability score | — | v1 |
| 4 | Budget guard | Latency + tokens (v0.2) + tool volume (v0.3) | v0.1–v0.3 |
| 5 | Prompt diff | — | v1 |
| 6 | Replay | Narrow tool-stub replay | v0.3 (was v1) |
| 7 | Multi-model benchmark | HTTP suites | v0.1 ✅ |
| 8 | CI quality gates | `exit 0/1` | beta.2 ✅ |

---

## Integration levels

| Level | Description | Release |
|-------|-------------|---------|
| L1 Serving | `POST /chat/completions` | beta.1 ✅ |
| L2 Gateway | Staging OpenAI-compatible URL | beta.1 ✅ |
| L3 Agent HTTP | Custom endpoint + **requested** `tool_calls` | v0.2 |
| L3b Agent process | Python SDK trajectory of **executed** tools | v0.3 |
| L4 Workflow / semantic | Soft judges, richer workflows | v1 |
| Runtime | Policies, production enforcement | **After** v0.3 — not MVP |

---

## Non-goals (MVP + current product)

- Replacing Langfuse / LangSmith in **production**
- Becoming an agent framework, orchestrator, or MCP server
- OpenTelemetry as the CI contract (ADR 0003)
- New trace database (ADR 0004)
- Kubernetes operators / distributed runners
- MMLU and academic model leaderboards
- Explicit `retries: 0` in YAML (beta limitation)
- LLM-based root-cause analysis in `inspect`

---

## Changelog

See [CHANGELOG.md](CHANGELOG.md).

# ConnorLLM Roadmap

> **The CI/CD reliability toolkit for AI systems** — quality gates before merge.

**Current release:** [v0.1.0](CHANGELOG.md#v010)  
**Status:** v0.1 shipped — AI Testing + first AI Regression gates (Go).

**North star:** [docs/vision.md](docs/vision.md) · **Doc map:** [docs/traceability.md](docs/traceability.md)

---

## Vision

Block merges when LLM/agent runtime regresses — availability, structured output, latency, pass rate — using versioned YAML suites and `exit 0` / `exit 1`.

**Product category:** AI Release Engineering (gatekeeper, not dashboard). See [vision.md](docs/vision.md).

**Not in scope:** production observability (Langfuse), model intelligence benchmarks (MMLU).

---

## Release timeline

| Release | Theme | CI value |
|---------|-------|----------|
| **v0.1.0-beta.1** ✅ | Serving smoke | "Does my endpoint respond?" |
| **v0.1.0-beta.2** ✅ | Agent output gates | "Does output match the contract?" |
| **v0.1.0-beta.3** ✅ | Regression compare (p95) | "Did p95 regress vs baseline?" |
| **v0.1.0** ✅ | Pass-rate gate + handbook | Full v0.1 regression gates |
| **v0.2.0** 📋 | Tool calls + cost | "Did the agent call the right tool?" |
| **v1.0.0** 📋 | Full agent CI | Workflows, replay, semantic eval (Python) |

Dates are indicative — ship when **exit criteria** below are met.

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
| 15 | Token cost | API budget exceeded? | 📋 v0.2 | `max_cost_regression` |
| | **Agent & tools** | | | |
| 16 | Tool call present | Agent called `search`? | 📋 v0.2 | `expect_tool` |
| 17 | Tool order | Multi-step plan respected? | 📋 v0.2 | `expect_tool_calls` |
| 18 | Custom agent HTTP | Non-OpenAI agent API? | 📋 v0.2 | Agent provider URL |
| | **Workflow & semantic** | | | |
| 19 | Multi-step | Chained scenario? | 📋 v1 | Workflow YAML |
| 20 | Replay prod | Replay prod run in CI? | 📋 v1 | `connor replay` |
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
- [x] Parse `tool_calls` from API response (PR-1)
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

## Planned — v1.0.0

**Theme:** Full agent CI (L4)

- [ ] Multi-step workflows, `connor replay`
- [ ] `services/evaluation/` (Python): semantic similarity, groundedness
- [ ] Reliability score with explicit N/A dimensions
- [ ] Prompt diff, model leaderboard

---

## Six engines (status)

| Engine | Today | Target |
|--------|-------|--------|
| Execution | Partial — HTTP provider, retry, timeout | Agent runner, tools |
| Evaluation | JSON, schema, contains (Go) | + Python semantic eval |
| Benchmark | Multi-case YAML suites | `connor compare` |
| Quality Gates | `exit 0/1` | Latency, pass-rate, cost thresholds |
| Observability | — | `run.json`, replay store |
| Developer Experience | CLI, parser, docs | SDK, feature docs |

Details: [docs/architecture.md](docs/architecture.md)

---

## Eight features (status)

| # | Feature | Status | Release |
|---|---------|--------|---------|
| 1 | Regression testing | compare p95 + pass rate | v0.1 ✅ |
| 2 | Tool call verification | — | v0.2 |
| 3 | Reliability score | — | v1 |
| 4 | Budget guard | Latency display only | v0.1 / v0.2 |
| 5 | Prompt diff | — | v0.2–v1 |
| 6 | Replay | — | v1 |
| 7 | Multi-model benchmark | Started | v0.1 |
| 8 | CI quality gates | `exit 0/1` OK | beta.2 ✅ |

---

## Integration levels

| Level | Description | Release |
|-------|-------------|---------|
| L1 Serving | `POST /chat/completions` | beta.1 ✅ |
| L2 Gateway | Staging OpenAI-compatible URL | beta.1 ✅ |
| L3 Agent API | Custom endpoint + tool checks | v0.2 |
| L4 Workflow | Multi-step, replay, semantic eval | v1 |

---

## Non-goals

- Replacing Langfuse / LangSmith in production
- Kubernetes operators / distributed runners (for now)
- MMLU and academic model leaderboards
- Explicit `retries: 0` in YAML (beta limitation)

---

## Changelog

See [CHANGELOG.md](CHANGELOG.md).

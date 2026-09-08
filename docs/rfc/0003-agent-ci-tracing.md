# RFC 0003 — CI for AI Agents (tracing, inspect, trajectory gates, replay)

| | |
|---|---|
| **Status** | Draft |
| **Target** | v0.3.0 (phased; first PR = data model + Python SDK) |
| **Issues** | TBD |
| **Roadmap** | UC #16–#20 evolution; Agent CI MVP |
| **ADRs** | [0003](../adr/0003-trace-model-not-otel.md), [0004](../adr/0004-run-artifact-additive-trajectory.md) |
| **Does not supersede** | RFC 0001 (run.json / compare), RFC 0002 (HTTP `tool_calls` + token cost) |

---

## 1. Problem

Connor today is a **black-box HTTP CI gate**.

It already answers:

- does this OpenAI-compatible endpoint respond?
- is the body JSON / schema / substring OK?
- did p95 or pass rate regress vs baseline?

RFC 0002 will add: did **one** chat-completions response *request* the right tool names, and did token totals regress.

It does **not** answer:

- what did the **agent actually do** (LLM → tool → LLM → tool)?
- did `refund` run twice? did `delete_customer` run at all?
- why did this CI case fail (trajectory, not only `reason: schema_mismatch`)?
- did tool-call volume or cost-per-run regress on a real agent loop?

Teams do not want Connor to become their agent framework. They want to **keep LangGraph / OpenAI Agents / custom Python** and still block the PR.

---

## 2. Goal (Agent CI MVP)

Position:

> **Connor — CI for AI Agents** — test agent behavior, tool usage, latency and cost before shipping.

After this RFC is fully shipped, a team can:

```python
from connor import trace, tool

@tool
async def search_customer(customer_id: str) -> dict: ...

@tool
async def refund(payment_id: str) -> dict: ...

with trace("support-agent") as run:
    await agent.run(...)
    run.write("candidate.json")  # Connor run.json (version 1, additive trajectory)
```

```bash
connor inspect candidate.json
connor inspect candidate.json --expect support-agent.yaml   # trajectory gates → exit 0|1
connor compare baseline.json candidate.json \
  --max-p95-regression 20 \
  --max-tool-calls-regression 50
connor replay candidate.json --candidate ./agent_entry.py   # Phase 5; narrow semantics
```

CI contract unchanged in spirit:

```text
gate fail → Connor exit 1 → CI fails → PR blocked
```

---

## 3. Out of scope (this MVP)

Connor is **not**: an agent framework, orchestrator, LLM provider, MCP server, generic observability platform, scheduler, production runtime, or security platform.

Explicitly later (Runtime / v1+ vision) — do not implement:

- policy engine, runtime action blocking, human approval, permissions
- automatic retries / recovery in production
- MCP proxy, multi-language SDK, complex dashboard
- LLM-based root-cause analysis (`inspect` is deterministic)
- `services/evaluation/` semantic / groundedness (still v1)
- replacing RFC 0002 HTTP `expect_tool` (different signal)

---

## 4. How Connor works today (Phase 0 map)

Do **not** rewrite this. Extend it.

```text
CLI (cobra)
  connor run [suite.yaml] [--out run.json]
  connor compare baseline.json candidate.json
        │
        ▼
benchmark.Parse          YAML → Spec
        │
        ▼
application.ExecuteSuite → ExecuteCase → EvaluateCase
        │                      │
        │                      ▼
        │              openai_compatible.Client
        │              POST {CONNOR_BASE_URL}/chat/completions
        │                      │
        │                      ▼
        │              entities.Response
        │              (Body, LatencyMs, ToolCalls, tokens)
        │                      │
        ▼                      ▼
domain.validation.Evaluate(body, Expectations)
  contains → JSON syntax → JSON schema
        │
        ▼
entities.CaseResult + SuiteResult
        │
        ├─ cli/output.PrintRun     → exit 0 | 1
        └─ entities.BuildRunArtifact → run.json (version 1)
                    │
                    ▼
         entities.CompareRuns      → exit 0 | 1 | 2
```

### 4.1 Component inventory

| Component | Path | Responsibility | Reuse vs change |
|-----------|------|----------------|-----------------|
| CLI root | `services/runtime/internal/cli/root.go` | Registers `run`, `compare` | **EXTEND** — add `inspect`, later `replay` |
| `connor run` | `internal/cli/run.go` | Single-case flags or YAML suite; `--out` | **KEEP** HTTP path. Do not turn it into an agent runner |
| `connor compare` | `internal/cli/compare.go` | Load two artifacts; optional flags; `Flags().Changed` | **EXTEND** new optional flags; same exit codes |
| Artifact I/O | `internal/cli/run_export.go` | `json.MarshalIndent` → file | **KEEP**; inspect/replay read the same files |
| Human output | `internal/cli/output/` | Lipgloss PASS/FAIL, greppable lines | **EXTEND** inspect renderer; compare rows |
| YAML spec | `internal/benchmark/spec.go` | `id, model, prompt, expect_*` | **EXTEND** later with `expect:` trajectory block (Phase 3). Parser still validates required `model`+`prompt` for HTTP cases |
| YAML parse | `internal/benchmark/parser.go` | Unmarshal + validate | **EXTEND** when new fields land |
| Execute suite/case | `internal/runtime/application/` | Retry, timeout, `EvaluateCase` | **KEEP** for HTTP. Trajectory eval is a sibling pure function, not inside `ExecuteCase` |
| Provider port | `internal/runtime/domain/ports.go` | `ProviderExecutor` | **KEEP**. Tracing does not implement this port |
| HTTP client | `.../openai_compatible/` | Chat completions; parse `tool_calls` + usage | **KEEP** (RFC 0002). Distinct from `@tool` spans |
| `Request` | `entities/request.go` | Model + messages | **KEEP**. Comment already: correlation IDs do not belong on Request |
| `Response` | `entities/response.go` | HTTP observation + `ToolCall{Name}` | **KEEP**. HTTP `tool_calls` ≠ executed tool spans |
| `Expectations` | `entities/expectations.go` | Contains / JSON / schema | **EXTEND** Phase 3 with trajectory expect struct |
| `FailReason` | `entities/case_result.go` | Stable strings | **EXTEND** new reasons; never rename old ones |
| `validation.Evaluate` | `domain/validation/evaluate.go` | Sequential body gates | **EXTEND** via a **gate list** (Phase 3), not a giant `if` |
| `RunArtifact` | `entities/run_artifact.go` | `version: 1`, cases, summary | **EXTEND** additive fields (ADR 0004) |
| `CompareRuns` | `entities/compare.go` | p95 + pass rate; `Passed` = AND | **EXTEND** new `*CompareResult` structs + nil-skip |
| Reliability | `domain/reliability/` | Retry / timeout | **KEEP**. Not used by the Python SDK |
| Tests | 21 `*_test.go` files | Table-driven gates, fixtures | **EXTEND** fixtures under `benchmarks/examples/` |
| Examples | `benchmarks/examples/` | HTTP suites + offline compare | **EXTEND** trace fixtures (no live agent required) |
| CI | `.github/workflows/ci.yml` | Go test/lint/build | **EXTEND** Phase 1 with Python unit tests |
| Env | `CONNOR_BASE_URL`, `CONNOR_API_KEY` | HTTP only | **KEEP**. Tracing does not require them |

### 4.2 What already exists for “agents”

- Marketing / examples use “agent” for **structured LLM output** (`agent-json-smoke.yaml`), not an agent loop.
- RFC 0002 `ToolCall` is **model-requested** tools in one HTTP message (`function.name` only; no args, no duration, no error).
- `run.json` already stores per-case `tool_calls`, `prompt_tokens`, `completion_tokens`, suite `total_tokens`.
- Compare is already the regression engine: pure functions, optional flags, AND composition, driver case on p95 FAIL.

**HTTP `tool_calls` and `@tool` spans are different signals.** RFC 0002 stays. Tracing is additive.

---

## 5. Product boundary

| In MVP | Out |
|--------|-----|
| Instrument *their* agent (Python) | Own the agent runtime |
| Record trajectory + metrics into `run.json` | New database / trace store |
| Deterministic inspect | LLM judge / RCA |
| Trajectory assertions + compare gates | Production enforcement |
| Narrow replay (recorded tool outputs) | Distributed deterministic replay |

Frameworks (LangGraph, OpenAI Agents SDK, Anthropic, MCP-enabled agents) are **in-process Python**. V1 has **no** framework-specific plugins. If a tool is a Python callable, `@tool` wraps it. That is enough.

---

## 6. Feature 1 — Agent tracing (Python SDK)

### 6.1 Target DX

```python
from connor import trace, tool

@tool
async def search_customer(customer_id: str) -> dict:
    ...

with trace("support-agent") as run:
    await agent.run(...)
```

A run is a tree of spans:

```text
Run  support-agent
├── llm.chat          (optional if wrapped)
├── tool.search_customer
├── llm.chat
├── tool.refund
└── (run status = error if uncaught exception)
```

V1 does **not** require auto-instrumenting the LLM client. LLM spans are optional (`@llm` or a tiny helper). Required: **run context + tool spans**. Nested `@tool` and concurrent tools must still parent correctly.

### 6.2 Minimal span model (V1)

Align names with OpenTelemetry **without** taking a dependency (ADR 0003).

| Field | V1? | Why |
|-------|-----|-----|
| `trace_id` | **required** | Correlate spans in one `trace()` |
| `run_id` | **required** | Stable id in artifact / inspect (`run_128` style: `run_` + hex) |
| `span_id` | **required** | Unique span |
| `parent_span_id` | **required** (empty on root) | Tree / inspect indentation |
| `name` | **required** | Tool or span name (`search_customer`, `llm.chat`) |
| `kind` | **required** | `run` \| `tool` \| `llm` \| `span` |
| `status` | **required** | `ok` \| `error` |
| `started_at` | **required** | RFC 3339 UTC |
| `ended_at` | **required** | RFC 3339 UTC |
| `duration_ms` | **required** | Integer ms; inspect + budgets |
| `error` | optional | Exception type + message (not full traceback in default artifact) |
| `input` | optional | JSON-serializable args; size-capped |
| `output` | optional | JSON-serializable result; size-capped |
| `attributes` | optional | Small string/number/bool map (`model`, `attempt`) |

**Not in V1:** OTel `SpanContext` bits, baggage, resource attributes, events-on-span, sampling, traceparent headers, full exception stack by default.

Tokens / cost: attach on `llm` spans when the caller sets them (`prompt_tokens`, `completion_tokens`). Suite totals = sum of spans (do not invent USD in V1; reuse ADR 0002 token semantics; optional `cost_usd` only if the caller sets it).

### 6.3 Instrumentation mechanics (required for V1)

| Concept | Role |
|---------|------|
| `contextvars` | Current span / run; correct across `await` and nested tools |
| `contextmanager` (`trace()`) | Root run span; flush on exit; record status on exception |
| Decorator `@tool` | Wrap sync **and** async (`inspect.iscoroutinefunction`); `functools.wraps` |
| Exception propagation | Record `status=error`, then **re-raise** — never swallow |
| Concurrency | Each task must copy context; concurrent tools = sibling spans under the same parent |
| Serialization | `json.dumps` default=str; cap payload bytes; on failure store `"<unserializable>"` |

Sync and async wrappers share one `_record_span` helper. Nested `trace()` in the same context is V1-forbidden (raise); nested `@tool` is required.

### 6.4 Artifact mapping

Python SDK writes a **Connor `run.json`** (ADR 0004):

- `version` remains `1`
- one agent invocation → one `cases[]` row (or one row per pytest case if the test creates one `trace()`)
- `cases[].trajectory` = `{ "run_id", "spans": [ ... ] }`
- `cases[].latency_ms` = root span `duration_ms`
- `cases[].passed` / `reason` filled later by Go gates (SDK default: `passed=true` if root status ok, else `passed=false`, `reason=agent_failed`)
- `summary` keeps existing KPIs; additive counters when trajectory present (`tool_calls`, `llm_calls`, `total_tokens` already exist)

No new storage. Inspect/compare/replay read the same file as `--out`.

### 6.5 Package layout (new)

```text
sdk/python/connor/
  __init__.py          # trace, tool
  trace.py             # context manager + contextvars
  tool.py              # decorator
  span.py              # dataclasses
  serialize.py         # run.json writer
  limits.py            # payload cap
```

Go types (Phase 1, **no CLI yet**):

```text
entities/trajectory.go   # Span, Trajectory
```

JSON tags must match the Python writer 1:1. Golden fixture: `benchmarks/examples/agent-trace-demo/run.json`.

---

## 7. Feature 2 — `connor inspect`

```bash
connor inspect <run.json>
```

**Purpose:** deterministic explanation of a recorded run. **No LLM.**

Reads existing `run.json`. If `trajectory` is missing, print today’s case table (id, passed, reason, latency, tool_calls names) and a one-line note that this is an HTTP-only artifact — still useful, not an error.

Example (trajectory present):

```text
Connor  v0.3.0
RUN     run_a1b2c3    support-agent

Status       FAILED
Duration     8.2s
LLM calls    4
Tool calls   7
Tokens       1840

Trajectory
✓  customer.lookup           120ms
✓  payment.lookup             80ms
⚠  refund                   5000ms
   └─ error  timeout
⚠  refund                    200ms
   └─ duplicate name (2nd call)
✗  send_email
   └─ error  ConnectionError
```

CLI: **EXTEND** `root.go`. Output: **EXTEND** `cli/output/` (do not put formatting in domain). Domain: pure `InspectViewFromArtifact(RunArtifact) InspectReport`.

Exit: `0` valid file; `2` invalid JSON / unknown version (same as compare usage). Inspect without `--expect` never exits `1` (display is not a gate).

---

## 8. Feature 3 — Agent assertions

### 8.1 Target YAML

New optional block. HTTP fields (`expect_contains`, …) stay valid.

```yaml
suite: support-agent-ci
cases:
  - id: refund-happy
    # HTTP cases still need model+prompt. Trace-only cases:
    source: artifact   # Phase 3; parser exception to model/prompt required
    expect:
      success: true
      tools:
        refund:
          max_calls: 1
        customer_lookup:
          min_calls: 1
      forbidden_tools:
        - delete_customer
      max:
        duration_ms: 5000
        tool_calls: 10
        tokens: 8000
```

MVP CLI (avoid a fourth command):

```bash
connor inspect candidate.json --expect support-agent.yaml
```

`--expect` set → apply trajectory (+ existing body gates if `cases[].` body fields exist) → **exit 1** on failure. Flag omitted → display only.

### 8.2 Assertion classes

**Required for MVP**

| Assertion | Fail reason (stable) |
|-----------|----------------------|
| `expect.success` | `agent_failed` |
| per-tool `min_calls` / `max_calls` | `tool_count` |
| `forbidden_tools` | `forbidden_tool` |
| `max.duration_ms` | `duration_exceeded` |
| `max.tool_calls` | `tool_budget_exceeded` |
| `max.tokens` | `token_exceeded` |

**Useful immediately after MVP**

- tool ordering (`expect_tool_calls` already RFC 0002 for HTTP; same names on **executed** spans)
- `max.llm_calls`
- `max.cost` (only if `cost_usd` present; else skip or exit 2 if flag/field set)
- tool error count
- `expect_tool` on **spans** vs RFC 0002 on **HTTP message** — keep fail reasons distinct if both fail (`tool_mismatch` stays HTTP)

**Later**

- argument deep equality, idempotency (`refund` same payment_id twice), semantic judges, allowed-tool allowlists with wildcards, parallel-vs-serial constraints

### 8.3 Architecture (no giant `if`)

Today `Evaluate(body, exp)` is three sequential checks — fine for three gates. Trajectory adds more. Introduce a small list:

```text
type Observation struct { Body string; Trajectory *Trajectory; Response Response }

type Gate interface {
    Active(exp Expectations) bool
    Check(obs Observation) (bool, FailReason)
}

EvaluateAll(obs, exp, gates)  // first failure wins; order documented
```

Body gates remain first (contains → json → schema) so HTTP suites do not change. Trajectory gates append. Each gate is a file in `domain/validation/` (same checklist as `go-runtime.mdc`).

`Expectations` grows a nested `Agent *AgentExpect` (`omitempty`). Empty / nil → all agent gates inactive.

---

## 9. Feature 4 — Trajectory regression (`compare`)

Keep:

```bash
connor compare baseline.json candidate.json
```

ADR 0001 still applies (suite_id, case ids, models). For **trace-only** cases, `model` may be a stable label (`support-agent` / `local`) so compare still works — document that.

**Minimum new metrics** (pre-aggregated in `summary` at write time; compare **must not** re-walk spans if summary is present — same anti-pattern as recomputing `pass_rate`):

| Summary field | Gate flag | Shape |
|---------------|-----------|--------|
| `p95_ms` | `--max-p95-regression` | **KEEP** |
| `pass_rate` | `--min-pass-rate` | **KEEP** |
| `total_tokens` | `--max-cost-regression` | RFC 0002 (tokens, not USD) |
| `tool_calls` (count) | `--max-tool-calls-regression` | % delta vs baseline, p95-style |
| per-tool mean calls | `--max-named-tool-regression name=N` **or** post-MVP | Phase 4b |

MVP compare additions: **suite-level `tool_calls` count** + reuse token gate. Per-tool “search_documents +291%” is **Phase 4b**: store `summary.tools: { "search_documents": { "calls": 12 } }` at artifact build; FAIL prints the driver tool name (mirror p95 driver).

`CompareResult.Passed` = AND of enabled gates. Missing metric while flag is set → **exit 2** (ADR 0002 pattern).

Human output: extra greppable lines:

```text
PASS  tool calls +4%
FAIL  tool calls +291%  (threshold: 50%)
      driver  search_documents  1.2 → 4.7 /run
```

---

## 10. Feature 5 — Replay (narrow)

```bash
connor replay <run.json> --candidate <entrypoint>
```

**V1 semantics (explicit):**

```text
Recorded run
  → initial input (root span input)
  → recorded tool outputs keyed by (name, ordinal) or (name, canonical args)
  → candidate code runs with @tool stubs returning recorded outputs
  → new trajectory written
  → operator runs connor compare original.json new.json
```

`connor replay` may invoke compare internally **only** if `--compare` is passed; default is write `replay.json` and print a small table.

**Must persist for this to work:** root `input`; each tool span `name`, `input`, `output`, `status`. LLM outputs are **not** replayed in V1 (candidate may call a live model).

**V1 guarantees**

- same user/task input
- tools do not hit real side effects **if** they went through `@tool` and stub mode is on
- a new trajectory is produced

**V1 does not guarantee**

- same LLM tokens, tool order, or latency
- tools not wrapped by `@tool`
- concurrent scheduling order
- time/clock-dependent code
- HTTP-only `run.json` without trajectory (exit 2)

Not a distributed-system time-travel debugger.

---

## 11. Target architecture (smallest)

```text
                 Connor CLI (Go)
        run (HTTP)   inspect   compare   replay
                       │
                       ▼
              domain (pure)
        EvaluateAll · CompareRuns · InspectReport
                       │
                       ▼
                  run.json v1
              (HTTP fields + optional trajectory)

  Python SDK ──writes──► run.json
  HTTP ExecuteSuite ──writes──► run.json
```

No new service, queue, or OTel collector. Python is an **SDK that writes artifacts**. Go remains the **CI contract** (exit codes, compare, inspect).

---

## 12. Technical concepts — when they are needed

### Required for MVP

| Concept | Why |
|---------|-----|
| Decorators + `functools.wraps` | `@tool` without breaking pytest/inspect |
| `*args` / `**kwargs` | Wrap arbitrary tools |
| Sync vs async | Real agents mix both |
| Context managers | `trace()` lifetime |
| `contextvars` | Parent span across await / nesting |
| Event/span model | Inspect + gates + compare |
| Exception propagation | Record then re-raise |
| Serialization + size caps | Artifact must stay CI-friendly |
| Pure gate functions | Same as current `Evaluate` / `CompareRuns` |
| Optional CLI flags (`Changed`) | No implicit CI fail |

### Useful for V2

| Concept | Why |
|---------|-----|
| OpenTelemetry SDK / OTLP | Export to Jaeger/Langfuse later; not required to *evaluate* |
| Proxy / LLM client middleware | Auto `llm` spans |
| Tool ordering state machine | Prefix/sequence gates |
| Idempotency keys | Duplicate `refund` detection beyond counts |
| Cancellation / timeouts on spans | Richer than process-level timeout |
| Canonical JSON args | Replay matching by args, not ordinal |

### Runtime vision only

| Concept | Why |
|---------|-----|
| Policy engine / interceptors that **block** | Production enforcement |
| MCP proxy | Traffic-level control |
| Human approval | Runtime, not CI |
| Scheduler / orchestrator | Not Connor’s job |
| Multi-language SDK | After Python DX is proven |

---

## 13. Backwards compatibility

| Piece | Action |
|-------|--------|
| `connor run` HTTP + YAML v0.1 fields | **KEEP** |
| `connor compare` p95 / pass-rate / exit 0/1/2 | **KEEP** |
| Fail reasons `call_failed`, `invalid_json`, `schema_mismatch`, `content_mismatch` | **KEEP** |
| `run.json` `version: 1` | **KEEP** (ADR 0004) |
| RFC 0002 `tool_calls` on the HTTP message | **KEEP** (ship v0.2) |
| `Evaluate(body)` | **KEEP**; add `EvaluateAll` |
| `inspect` / `replay` / Python SDK | **NEW** |
| Trajectory fields / agent expect / compare flags | **EXTEND** |
| Giant rewrite, new store, OTel required | **REJECT** |

Old `connor compare` binaries ignore unknown JSON fields → new artifacts still compare on p95/pass_rate. New flags on old artifacts without summary counters → exit 2.

Parser: HTTP cases still require `model` + `prompt`. `source: artifact` is opt-in (Phase 3) so existing suites keep failing closed if those fields are missing.

---

## 14. Implementation plan (PRs)

| Phase | PR | Scope |
|-------|----|--------|
| **1** | PR-A | Go `Trajectory`/`Span` + Python `trace`/`@tool` + golden `run.json` + tests. **No new CLI command.** |
| **2** | PR-B | `connor inspect` display-only |
| **3** | PR-C | Gate list + agent expect YAML + `inspect --expect` exit 1 |
| **4** | PR-D | Compare: tool-call count regression (+ 4b named-tool driver) |
| **5** | PR-E | Replay stubs + documented non-guarantees |

RFC 0002 PRs (expect_tool, cost) continue in **parallel** on the HTTP path. Do not block Phase 1 on RFC 0002 PR-2.

---

## 15. Decisions (resolved for this RFC)

| Question | Decision |
|----------|----------|
| OTel as V1 runtime? | **No** — JSON spans, OTel-shaped names (ADR 0003) |
| New artifact format? | **No** — additive `run.json` v1 (ADR 0004) |
| Does `connor run` launch agents? | **No** — HTTP only |
| Inspect uses an LLM? | **No** |
| USD cost in V1? | **No** unless caller sets `cost_usd`; compare cost stays tokens (ADR 0002) |
| Framework plugins? | **No** in V1 |
| Absolute trajectory gates vs compare? | Both: `--expect` on inspect (absolute); compare flags (delta) |
| Swallow tool errors? | **Never** — record + re-raise |

---

## 16. Done when (MVP exit criteria)

- [ ] `@tool` + `trace()` produce a version-1 `run.json` with correlated spans (sync + async)
- [ ] Exceptions propagate; span `status=error`
- [x] `connor inspect run.json` prints trajectory from the artifact
- [ ] `inspect --expect` fails CI on `max_calls` / `forbidden_tools` / `success`
- [ ] `compare --max-tool-calls-regression` AND-composes with existing gates
- [ ] Replay docs state guarantees / non-guarantees; HTTP-only artifact → exit 2
- [ ] CHANGELOG + ci-gates rule + examples
- [ ] Existing HTTP suites and compare fixtures still pass unchanged

---

## 17. First PR (only)

### Title

`feat: agent trajectory data model and Python trace/@tool`

### Problem

Connor gates a single HTTP response. Teams cannot record what their agent *executed*. We need a serializable trajectory before `inspect` / YAML expect / replay.

### Scope

- Go: `entities.Span`, `entities.Trajectory` with JSON tags matching the SDK
- Python: `sdk/python/connor` — `trace()`, `@tool` (sync + async), `run.write(path)`
- Golden fixture `benchmarks/examples/agent-trace-demo/run.json`
- Tests listed below
- Docs: CHANGELOG Unreleased; this RFC stays Draft until P1 merges (then mark Phase 1 done)

### Non-goals

`connor inspect`, YAML `expect:`, compare flags, replay, OpenTelemetry, LangGraph plugins, cost USD, LLM auto-wrap.

### Proposed architecture

`contextvars` current span; `trace()` is the root context manager; `@tool` opens a child span, records I/O (capped), sets status, **re-raises**. Writer emits `run.json` version 1 with `cases[0].trajectory`.

### Files

**New:** `sdk/python/connor/*`, `services/runtime/internal/runtime/domain/entities/trajectory.go`, `trajectory_test.go`, `benchmarks/examples/agent-trace-demo/run.json`, `sdk/python/tests/*`

**Touched (minimal):** `CHANGELOG.md`; `entities/run_artifact.go` — add `Trajectory *Trajectory \`json:"trajectory,omitempty"\`` on `RunCase` so Go can unmarshal the fixture. `ParseRunArtifactJSON` stays version 1. No CLI.

### Public API

```python
from connor import trace, tool
```

### Data model

See §6.2. Root span `kind=run`; tools `kind=tool`.

### Tests

- Async two-tool happy path → parent/child ids, order, both `ok`
- Sync `@tool`
- Nested `@tool`
- Concurrent `asyncio.gather` → siblings, not parented to each other
- Tool raises `ValueError` → span `error`, exception identity preserved
- `trace()` after exception still writes artifact (`status=error` on run)
- Unserializable output → placeholder, no crash
- Go: unmarshal golden fixture; reject broken span (missing `span_id`)

### Acceptance criteria

```python
with connor.trace("test-agent") as run:
    await search_customer("123")
    await refund("payment-456")
    run.write("run.json")
```

produces `run → search_customer, refund` with stable `run_id`, correlated spans, durations, success/error, version-1 artifact. `connor compare` on two HTTP fixtures still passes (no CLI regression).

### Follow-up PRs

P2 inspect → P3 expect → P4 compare tool volume → P5 replay.

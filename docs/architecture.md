# ConnorLLM Architecture

> **Connor** is CI for AI Agents — quality gates **before merge**, not production observability.

Two complementary surfaces share one artifact (`run.json` version 1):

1. **HTTP black-box** (shipped): Go CLI calls an OpenAI-compatible `/chat/completions`.
2. **Agent white-box** (RFC 0003 P1–P2): Python SDK records a trajectory; `connor inspect` explains the same `run.json`.

Your agent stack stays yours. Connor instruments and evaluates.

---

## Six engines

| # | Engine | Role | Today | Target |
|---|--------|------|-------|--------|
| 1 | **Execution** | Providers, retry, timeout, suite runs | `services/runtime/` (Go HTTP) | **Not** an agent runner. Python SDK is out-of-process instrumentation |
| 2 | **Evaluation** | JSON, schema, contains; later trajectory gates | Go: syntax + schema + text | Gate **list** in Go (RFC 0003); `services/evaluation/` Python semantic (v1) |
| 3 | **Benchmark** | Multi-case suites, model comparison | YAML suites + `compare` | Tool-volume regression (v0.3) |
| 4 | **Quality Gates** | CI pass/fail, budgets | `exit 0/1/2`, fail reasons | More optional flags; same contract |
| 5 | **Observability** | Run artifacts (CI scope) | `run.json` v1 | Additive `trajectory` (ADR 0004). **No** collector / Langfuse |
| 6 | **Developer Experience** | CLI, parser, docs, GitHub Actions | `run`, `compare`, `inspect` | `replay`; trajectory `--expect` |

**Shipped today:** Execution (HTTP) + Evaluation (deterministic body gates) + Benchmark compare (p95, pass rate) + DX (`connor run` / `compare` / `inspect` display-only).

**Do not implement yet:** trajectory YAML `--expect`, compare tool-volume flags, replay (RFC 0003 P3–P5).

---

## Data flow (HTTP — shipped)

```text
YAML suite  →  benchmark.Parse  →  application.ExecuteSuite
                                        │
                                        ▼
                              application.ExecuteCase
                              (HTTP via openai_compatible)
                                        │
                                        ▼
                              domain.validation.Evaluate
                              (contains → JSON → schema)
                                        │
                                        ▼
                              entities.CaseResult
                                        │
                    ┌───────────────────┴───────────────────┐
                    ▼                                       ▼
          cli/output.PrintRun                    entities.BuildRunArtifact
          exit 0 | 1                             run.json version 1
                                                        │
                                                        ▼
                                              entities.CompareRuns
                                              exit 0 | 1 | 2
```

## Data flow (agent CI — RFC 0003)

```text
Python agent (LangGraph / custom / …)
        │  sdk: trace() + @tool
        ▼
run.json v1 + cases[].trajectory
        │
        ├─ connor inspect          → tree, timing, errors (no LLM)   [P2]
        ├─ inspect --expect YAML   → trajectory gates → exit 0 | 1  [P3]
        └─ connor compare          → p95 / pass rate / tool-call delta
```

`connor run` stays HTTP-only. The SDK does **not** go through `ProviderExecutor`.

---

## Layering (DDD)

| Layer | Path | Responsibility |
|-------|------|----------------|
| Benchmark (infra) | `internal/benchmark/` | YAML → `Spec`, parse-time validation |
| Application | `internal/runtime/application/` | Orchestrate cases, wire expectations |
| Domain | `internal/runtime/domain/` | `Request`, `Response`, `Expectations`, gates |
| Infrastructure | `internal/runtime/infrastructure/` | OpenAI-compatible HTTP client |
| CLI | `internal/cli/` | `connor run`, `compare`, `inspect`; planned `replay` |

Domain code does not import YAML or HTTP client types.

---

## Active gates (beta.2)

| YAML field | Fail reason | Question |
|------------|-------------|----------|
| *(implicit)* | `call_failed` | HTTP 2xx? Timeout? Retries exhausted? |
| `expect_json` | `invalid_json` | Valid JSON syntax? |
| `expect_json_schema` | `schema_mismatch` | Matches JSON Schema? (syntax implied) |
| `expect_contains` | `content_mismatch` | Body contains substring? |
| `expect_contains_ignore_case` | `content_mismatch` | Case-insensitive contains |

Evaluation order: **contains → JSON syntax → JSON schema**.

---

## Repository layout

```text
ConnorLLM/
├── services/runtime/           # Go CI contract (run, compare, inspect)

│   ├── cmd/connor/
│   └── internal/
│       ├── benchmark/          # YAML parser
│       ├── cli/                # Commands + output
│       └── runtime/
│           ├── application/    # ExecuteSuite, EvaluateCase (HTTP)
│           ├── domain/         # Entities, validation, reliability
│           └── infrastructure/ # openai_compatible provider
├── sdk/python/connor/          # RFC 0003 P1 — trace / @tool (writes run.json)
├── benchmarks/examples/        # Runnable demo suites + offline fixtures
├── docs/                       # RFC, ADR, architecture, vision
└── ROADMAP.md
```

**Shipped:** `sdk/python/connor` (instrumentation) + `connor inspect`. **Later:** `services/evaluation/` (semantic judges, v1) — different from the tracing SDK.

---

## Integration levels

| Level | Description | Status |
|-------|-------------|--------|
| L1 Serving | `POST /chat/completions` | ✅ beta.1 |
| L2 Gateway | Staging/prod OpenAI-compatible URL | ✅ config only |
| L3 Agent HTTP | Custom agent endpoint + **requested** tool names | 🔜 v0.2 (RFC 0002) |
| L3b Agent process | Python trajectory of **executed** tools | 🔜 v0.3 (RFC 0003) |
| L4 Workflow | Semantic eval, richer workflows | 🔜 v1 |

---

## Non-goals

- Production APM / tracing platforms (Langfuse territory). **CI** tracing of test runs is in RFC 0003.
- Academic model benchmarks (MMLU)
- Replacing your agent runtime (Python, LangGraph, …)
- OpenTelemetry as the evaluation runtime (ADR 0003)
- A new trace store (ADR 0004)

---

## Further reading

- [ROADMAP.md](../ROADMAP.md) — releases, use-case matrix, first PR
- [RFC 0003](rfc/0003-agent-ci-tracing.md) — Agent CI design
- [README.md](../README.md) — quick start and demo
- Example suites: `benchmarks/examples/`

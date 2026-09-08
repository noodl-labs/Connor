# Connor — Product vision

> **CI for AI Agents** — decide whether a new agent or LLM system version is safe to merge.

Longer-term: **reliability infrastructure for AI agents**. That is a later product. Ship CI gates first.

---

## One-liner

DeepEval measures answer quality. Langfuse observes production. Promptfoo compares prompts. Agent frameworks *run* agents.  
**Connor decides if you can merge** — via CI gates on HTTP responses **and** (v0.3) recorded agent trajectories. It does not become your runtime.

---

## Paradigm shift

| Traditional eval tools | Connor |
|------------------------|--------|
| Developer runs tests → reads metrics → decides | Git push → Connor → **PASS / FAIL** → merge blocked or allowed |
| Dashboard-first | **Gatekeeper-first** |
| "Is this response good?" | "Is this **release** acceptable?" |

---

## North-star workflow

```
Developer
    │
    ▼
Git push / PR
    │
    ▼
Connor
    ├── Run benchmark suite
    ├── Export run.json (candidate)
    ├── Compare vs baseline
    ├── Apply quality gates (latency, pass rate, cost, tools…)
    └── Report PASS / FAIL (exit 0 / 1)
    │
    ▼
Merge allowed or blocked
```

Future: `connor release` orchestrates run → compare → diff → risk → recommendation in one command.

---

## Phased product evolution

Each phase adds a layer. Ship incrementally — do not build V10 before V2 works.

| Phase | Name | Question | Connor today |
|-------|------|----------|--------------|
| **V0** | Engineering foundation | Is the engine reliable? | ✅ Go runtime, YAML, provider, retry, tests |
| **V1** | AI Testing | Does my system still work? | ✅ `connor run`, gates, exit 0/1 |
| **V2** | AI Regression | Did we degrade vs baseline? | ✅ `compare`, p95, pass rate (Go); tokens v0.2; tool volume v0.3 |
| **V2b** | CI for AI Agents | What did the agent *do*? | 📋 RFC 0003 — trace, inspect, trajectory gates, narrow replay |
| **V3** | AI Release Engineering | Can we merge this PR? | 🟡 Handbook + CLI flags; Action/PR comments later |
| **V4** | AI Reliability Platform | Why is this version worse? | 🟡 p95 driver only; **deterministic** inspect first; LLM RCA later |
| **V5** | AI Assets Platform | What assets define this run? | ❌ Versioned prompts, baselines, datasets |
| **V6–V10** | Runtime / Control Plane | Prod enforcement, policies, FinOps | ❌ **After** V2b works. Do not build Runtime in the Agent CI MVP. |

**Language split (intentional):**

- **Go** — execution (HTTP), hard gates, `run.json`, `compare` / `inspect`, exit codes (CI contract).
- **Python SDK** (v0.3) — `trace()` / `@tool` in the user’s process; writes `run.json`. Not an agent framework.
- **Python eval service** (v1+) — soft evaluators (semantic similarity, groundedness) feeding the same artifacts.

Do not rewrite regression compare in Python. Do not put semantic judges in the tracing SDK.

---

## Differentiation (future depth)

Capabilities to deepen V2→V4 without changing category:

| Capability | Value |
|------------|-------|
| Root cause analysis | Which prompt / model / metric / case failed |
| Explainable regression | "37 cases changed: 18 verbose, 9 wrong tool…" |
| Release risk score | Quality / latency / cost / tools → MEDIUM risk |
| SLO-driven gates | p95, cost, pass rate, schema as team SLOs |
| AI FinOps | Token + € impact of a PR |
| Quality diff | Side-by-side answers + why B differs from A |
| Regression timeline | When did pass rate start dropping (v18→v21) |

---

## What Connor is not

- Production observability (Langfuse, LangSmith)
- Academic model benchmarks (MMLU)
- A chatbot / agent framework
- An MCP server, scheduler, or production policy runtime
- A generic LLM playground

---

## Traceability

| Doc | Role |
|-----|------|
| [ROADMAP.md](../ROADMAP.md) | Releases, use cases, exit criteria |
| [traceability.md](traceability.md) | Vision phase ↔ release ↔ RFC ↔ ADR ↔ issue |
| [rfc/](rfc/) | Design before code |
| [adr/](adr/) | Irreversible decisions |
| [handbook/ci-regression.md](handbook/ci-regression.md) | User-facing CI guide |

**Workflow for new behavior:** GitHub Issue → RFC (design) → ADR (if decision) → PR (code) → CHANGELOG.

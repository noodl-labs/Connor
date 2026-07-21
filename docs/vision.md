# Connor — Product vision

> **AI Release Engineering** — decide whether a new AI system version is safe to deploy.

---

## One-liner

DeepEval measures answer quality. Langfuse observes production. Promptfoo compares prompts.  
**Connor decides if you can merge and deploy** — via CI gates, not dashboards.

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
| **V2** | AI Regression | Did we degrade vs baseline? | ✅ `compare`, p95, pass rate (Go) |
| **V3** | AI Release Engineering | Can we merge this PR? | 🟡 Handbook + CLI flags; Action/PR comments later |
| **V4** | AI Reliability Platform | Why is this version worse? | 🟡 p95 driver only; RCA taxonomy later |
| **V5** | AI Assets Platform | What assets define this run? | ❌ Versioned prompts, baselines, datasets |
| **V6–V10** | Deployments → Control Plane | Prod lifecycle, obs, FinOps | ❌ Out of scope until V2–V4 land |

**Language split (intentional):**

- **Go** — execution, hard gates, `run.json`, `compare`, exit codes (CI contract).
- **Python** (v1+) — soft evaluators (semantic similarity, groundedness) feeding the same artifacts.

Do not rewrite regression compare in Python — extend with eval scores later.

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
- A chatbot framework
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

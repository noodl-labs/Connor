# RFC 0002 — Tool & cost gates

| | |
|---|---|
| **Status** | Draft |
| **Target** | v0.2.0 |
| **Issues** | TBD |
| **Roadmap** | UC #15, #16, #17, #18 |

---

## 1. Problem

v0.1 gates cover HTTP output, JSON/schema, p95 latency, and pass rate.

Teams also need CI to block when:

- an agent **calls the wrong tool** (or skips a required tool)
- **token cost** regresses vs baseline or exceeds budget

---

## 2. Goal (v0.2)

After this RFC ships:

```yaml
cases:
  - id: book-flight
    prompt: ...
    expect_tool: search
    # expect_tool_calls: [{name: search}, {name: book}]
```

```bash
connor compare baseline.json candidate.json --max-cost-regression 15
```

Extend `run.json` with tool call metadata and token usage where available.

---

## 3. Out of scope (v0.2)

- Semantic similarity / LLM-as-judge (Python, v1)
- `connor release` unified pipeline
- Production observability
- YAML `regression:` block (CLI flags only, same as v0.1)

---

## 4. Proposed API

### YAML (per case)

| Field | Gate | Fail reason (TBD) |
|-------|------|-------------------|
| `expect_tool` | Named tool present in response | `tool_mismatch` |
| `expect_tool_calls` | Ordered tool names | `tool_order_mismatch` |

### run.json extensions (draft)

- Per case: `tool_calls[]` (name, args hash optional)
- Summary: `prompt_tokens`, `completion_tokens`, `estimated_cost_usd` (if pricing known)

### compare flags

| Flag | Rule |
|------|------|
| `--max-cost-regression N` | Fail if candidate cost delta % > N vs baseline |

---

## 5. Implementation plan (draft)

| PR | Scope |
|----|-------|
| PR-1 | Parse `tool_calls` from OpenAI-compatible response |
| PR-2 | `expect_tool` / `expect_tool_calls` + fail reasons |
| PR-3 | Token usage in run.json + cost compare gate |
| PR-4 | `benchmarks/examples/agent-support.yaml` |

---

## 6. Open questions

- [ ] Agent HTTP provider URL — separate RFC section or same?
- [ ] Cost without list price — store tokens only, compare token regression?
- [ ] ADR needed for pass-rate-style absolute vs delta on cost?

---

## 7. Done when (v0.2 exit criteria)

- [ ] Wrong tool → `connor run` exit 1
- [ ] Cost regression demo with `compare`
- [ ] Documented in handbook + CHANGELOG

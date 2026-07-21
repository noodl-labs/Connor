# RFC 0002 — Tool & cost gates

| | |
|---|---|
| **Status** | Accepted |
| **Target** | v0.2.0 |
| **Issues** | [#9](https://github.com/noodl-labs/Connor/issues/9) |
| **Roadmap** | UC #15, #16, #17, #18 |
| **ADRs** | [0002](../adr/0002-cost-regression-tokens.md) |

---

## 1. Problem

v0.1 gates cover HTTP output, JSON/schema, p95 latency, and pass rate.

Teams also need CI to block when:

- an agent **calls the wrong tool** (or skips a required tool)
- **token / cost** regresses vs baseline

---

## 2. Goal (v0.2)

After this RFC ships:

```yaml
cases:
  - id: book-flight
    model: openai/gpt-4o-mini
    prompt: |
      Use tools to find a flight to Paris.
    expect_tool: search
    # expect_tool_calls:
    #   - search
    #   - book
```

```bash
connor run suite.yaml --out candidate.json
connor compare baseline.json candidate.json --max-cost-regression 15
```

- Wrong / missing tool → `connor run` exit `1` with stable fail reason
- Cost (token) regression → `connor compare` exit `1` when flag set

---

## 3. Out of scope (v0.2)

- Semantic similarity / LLM-as-judge (Python, v1)
- `connor release` unified pipeline
- Production observability / FinOps € pricing tables
- YAML `regression:` block (CLI flags only, same as v0.1)
- Argument deep equality on tool calls (name + order only in v0.2)
- `p90` / `p99` latency flags

---

## 4. API

### 4.1 YAML (per case)

| Field | Rule | Fail reason |
|-------|------|-------------|
| `expect_tool` | At least one tool call whose **name** equals the string | `tool_mismatch` |
| `expect_tool_calls` | Ordered list of tool **names**; must match prefix of actual calls in order | `tool_order_mismatch` |

Evaluation order (extends v0.1): contains → json → schema → **tool gates**.

If both `expect_tool` and `expect_tool_calls` are set: evaluate `expect_tool_calls` first; skip `expect_tool` if order gate already failed (or require both — **decision:** both must pass when both set).

### 4.2 Provider response

Parse OpenAI-compatible `message.tool_calls` (name from `function.name` or equivalent).

Missing `tool_calls` when a tool gate is set → fail with the same reason as mismatch.

### 4.3 `run.json` extensions (`version` stays `1` for v0.2)

Additive fields only (unknown fields ignored by older compare if needed; v0.2 compare requires fields when cost gate is on).

Per case:

```json
"tool_calls": [{ "name": "search" }],
"prompt_tokens": 120,
"completion_tokens": 40
```

Summary:

```json
"prompt_tokens": 480,
"completion_tokens": 160,
"total_tokens": 640
```

`total_tokens` = sum of case tokens (or suite-level from API if available).  
No `estimated_cost_usd` in v0.2 (no pricing table).

### 4.4 Compare flag

| Flag | Rule | When omitted |
|------|------|--------------|
| `--max-cost-regression N` | Fail if token delta `%` **>** N | Gate skipped |

```
delta = (candidate.total_tokens - baseline.total_tokens) / baseline.total_tokens * 100
```

- baseline `total_tokens == 0` and candidate `> 0` → treat as `+100%` (same spirit as p95)
- Gate uses **token totals**, not USD (ADR 0002)

Output (greppable):

```text
PASS  cost +8%
FAIL  cost +34%  (threshold: 15%)
```

Compose with existing p95 / pass-rate gates via AND.

### 4.5 Custom agent HTTP (UC #18)

Same release, last PR: optional `CONNOR_AGENT_URL` (or flag) for non-`/chat/completions` agent endpoints that still return OpenAI-shaped tool_calls + usage. If capacity slips, ship tools+cost first and defer provider to a follow-up patch — but keep in RFC scope.

---

## 5. Exit codes & fail reasons

| Command | Behavior |
|---------|----------|
| `connor run` | exit `1` on `tool_mismatch` / `tool_order_mismatch` (and existing reasons) |
| `connor compare` | exit `1` if any enabled gate fails (incl. cost); exit `2` if incomparable / invalid |

Fail reasons are **stable strings** — CHANGELOG if renamed.

---

## 6. Implementation plan

| PR | Scope | Branch suggestion |
|----|-------|-------------------|
| **PR-1** | Parse `tool_calls` + tokens into domain / run.json | `feat/CON07-PARSE-TOOL-CALLS` |
| **PR-2** | `expect_tool` / `expect_tool_calls` + tests + badges | `feat/CON08-EXPECT-TOOL` |
| **PR-3** | `--max-cost-regression` + compare UX | `feat/CON09-COST-REGRESSION` |
| **PR-4** | `benchmarks/examples/agent-support.yaml` + handbook | `feat/CON10-AGENT-SUPPORT-EXAMPLE` |
| **PR-5** | Custom agent HTTP provider (if not slipped) | `feat/CON11-AGENT-PROVIDER` |

---

## 7. Decisions (resolved)

| Question | Decision |
|----------|----------|
| Cost without list price? | Compare **`total_tokens`** delta %; no USD in v0.2 |
| Absolute vs delta cost? | **Delta vs baseline** (like p95), not absolute — ADR 0002 |
| Agent HTTP provider? | In scope as PR-5; may slip to v0.2.1 |
| Tool arg validation? | Names (+ order) only |
| Flag omitted? | Gate skipped (no implicit fail) |
| run.json version bump? | Stay at `1`; additive fields |

---

## 8. Done when (v0.2 exit criteria)

- [ ] Wrong / missing tool → `connor run` exit `1` (`tool_mismatch` or `tool_order_mismatch`)
- [ ] `connor compare ... --max-cost-regression N` fails in a demo when tokens regress
- [ ] Example suite `benchmarks/examples/agent-support.yaml`
- [ ] Handbook + CHANGELOG + `ci-gates` rule updated
- [ ] Tag `v0.2.0` published

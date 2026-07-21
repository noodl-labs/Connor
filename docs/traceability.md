# Traceability — vision ↔ releases ↔ docs

Living map between product phases, shipped releases, and design docs.

Last updated: RFC 0002 Accepted + issue #9 (v0.2 kickoff).

---

## Release status (today)

| Release | Status | Tag | Theme |
|---------|--------|-----|-------|
| v0.1.0-beta.1 | ✅ Shipped | — | Serving smoke |
| v0.1.0-beta.2 | ✅ Shipped | — | Output gates |
| v0.1.0-beta.3 | ✅ Shipped | `v0.1.0-beta.3` | p95 compare |
| **v0.1.0** | ✅ Shipped | `v0.1.0` | pass rate + handbook + demo |
| **v0.2.0** | 🔜 In design | — | tools + cost ([#9](https://github.com/noodl-labs/Connor/issues/9)) |
| v1.0.0 | 📋 Planned | — | workflows + Python eval |

---

## Vision phase → release

| Vision phase | Maps to | Shipped? |
|--------------|---------|----------|
| V0 Foundation | beta.1 infra | ✅ |
| V1 AI Testing | beta.1–2 `connor run` | ✅ |
| V2 AI Regression | beta.3 + v0.1 `compare` | ✅ (hard gates); + cost in v0.2 |
| V3 Release Engineering | v0.1 handbook; Action later | 🟡 |
| V4+ Reliability / Assets / Control Plane | v1+ | ❌ |

Details: [vision.md](vision.md).

---

## RFC index

| RFC | Title | Status | Target | Issues | Shipped in |
|-----|-------|--------|--------|--------|------------|
| [0001](rfc/0001-run-artifact-and-compare.md) | Run artifact & compare | Accepted | v0.1.0 | #4 | v0.1.0-beta.3 / v0.1.0 |
| [0002](rfc/0002-tool-and-cost-gates.md) | Tool & cost gates | Accepted | v0.2.0 | [#9](https://github.com/noodl-labs/Connor/issues/9) | — |

**When to open a new RFC:** new CLI command, new YAML field, new gate, or exit-code contract change.

---

## ADR index

| ADR | Title | Status | RFC |
|-----|-------|--------|-----|
| [0001](adr/0001-comparable-runs.md) | Comparable run definition | Accepted | 0001 |
| [0002](adr/0002-cost-regression-tokens.md) | Cost gate = token delta vs baseline | Accepted | 0002 |

**When to open a new ADR:** one irreversible decision.

---

## PR / issue trace (v0.1)

| Work | Type | RFC § | Merged |
|------|------|-------|--------|
| run.json export | PR | 0001 §4, PR-1 | ✅ beta.3 |
| `connor compare` + p95 | PR | 0001 §5.2, PR-2 | ✅ beta.3 |
| `--min-pass-rate` | PR | 0001 §5.3, PR-3 | ✅ v0.1.0 |
| CI handbook | PR | 0001 §6, PR-4 | ✅ v0.1.0 |
| regression-demo + Lipgloss compare | PR | — | ✅ post-v0.1.0 |

---

## v0.2 track ([#9](https://github.com/noodl-labs/Connor/issues/9))

| Step | Status |
|------|--------|
| Issue opened | ✅ #9 |
| RFC 0002 Accepted | ✅ |
| ADR 0002 Accepted | ✅ |
| PR-1 Parse tool_calls + tokens | ✅ `feat/CON07-PARSE-TOOL-CALLS` |
| PR-2 expect_tool / expect_tool_calls | 📋 |
| PR-3 --max-cost-regression | 📋 |
| PR-4 agent-support.yaml + handbook | 📋 |
| PR-5 custom agent HTTP (may slip) | 📋 |
| Tag v0.2.0 | 📋 |

---

## Doc types (who reads what)

| Audience | Read |
|----------|------|
| User / DevOps | README, handbook |
| Contributor | RFC, ADR, architecture, go-runtime rule |
| Product / you | vision, traceability, ROADMAP |

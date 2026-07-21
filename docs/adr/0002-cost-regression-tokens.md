# ADR 0002 — Cost regression uses token delta vs baseline

## Status
Accepted

## Context
RFC 0002 adds a compare gate for “cost”. Teams may lack USD pricing. Pass-rate gate is absolute on the candidate; p95 gate is a **percent delta** vs baseline.

## Decision
`--max-cost-regression` compares **`summary.total_tokens`** between baseline and candidate as a percent change (same shape as p95). It does **not** use an absolute token threshold and does **not** require USD pricing in v0.2.

## Consequences
- run.json must export token totals when available
- Missing tokens on either side when the flag is set → exit `2` (usage / incomparable metrics)
- Future FinOps (€) can add a separate gate without changing this ADR

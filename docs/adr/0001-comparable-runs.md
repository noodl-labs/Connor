# ADR 0001 — Comparable run definition

## Status
Accepted

## Context
RFC 0001 requires comparing baseline vs candidate runs.

## Decision
Two runs are comparable only if suite_id, case ids, and per-case model match.
`connor compare` exits 2 if not comparable.

## Consequences
- GPT-4 baseline vs GPT-5 candidate is rejected
- run.json must include suite_id and model per case
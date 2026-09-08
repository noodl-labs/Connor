# ADR 0004 — Trajectory is additive on run.json version 1

## Status
Accepted

## Context
RFC 0001 defined `run.json` with `"version": 1`. Unknown versions make `connor compare` exit 2. RFC 0002 already extended version 1 additively (`tool_calls`, token totals) so older comparators keep working (unknown JSON fields are ignored).

RFC 0003 could bump to version 2 when `trajectory` is present. That would break any v0.1/v0.2 `connor compare` binary that rejects `version != 1`, even if the team only wanted p95/pass-rate gates.

## Decision
Keep `"version": 1`.

Add optional fields:

- `cases[].trajectory` — `{ "run_id", "spans": [ ... ] }`
- summary counters needed for compare (e.g. `tool_call_count`) written at artifact build time

Rules:

- HTTP-only artifacts omit `trajectory` (current exporters unchanged).
- Compare flags that need trajectory/summary counters: if the flag is set and the metric is missing → **exit 2** (same as ADR 0002 missing tokens).
- Inspect without trajectory: print the HTTP case table, do not fail.
- Do **not** recompute suite KPIs from spans when summary fields are already present.

A version bump (2) is reserved for a breaking change (rename/remove fields), not for adding agent CI.

## Consequences
- One artifact format for HTTP smoke and agent traces.
- Old CLIs still compare p95/pass_rate on new files.
- Python SDK and Go `BuildRunArtifact` must agree on JSON tags.
- Document that `version` is a **breaking-schema** counter, not a feature flag.

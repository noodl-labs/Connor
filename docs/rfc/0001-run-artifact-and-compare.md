# RFC 0001 — Run artifact & compare

| | |
|---|---|
| **Status** | Accepted |
| **Target** | v0.1.0 |
| **Issues** | #4 (latency regression) |
| **Roadmap** | UC #13, #14, #24 |

---

## 1. Problem

Connor blocks bad HTTP responses and bad output (JSON, schema, contains).

It does **not** block:

- latency getting worse (e.g. p95: 820ms → 2.8s)
- pass rate dropping below a team threshold (e.g. 95%)

There is no saved run file and no `compare` command.

---

## 2. Goal (v0.1)

After this RFC ships, a user can:

```bash
# 1. Save a run
connor run suite.yaml --out baseline.json

# 2. Save another run
connor run suite.yaml --out candidate.json

# 3. Compare and fail CI if gates break
connor compare baseline.json candidate.json \
  --max-p95-regression 20 \
  --min-pass-rate 95
```

Exit codes:

- `0` — compare passed
- `1` — gate failed (latency regression or pass rate)
- `2` — usage error (invalid file or runs not comparable)

---

## 3. Out of scope (v0.1)

- Token cost / `usage` parsing → v0.2
- `p90`, `p99`, tokens/sec
- Weighted tests (critical / high / medium)
- Streaming / TTFT (not shipped in beta.2)
- Run database or history service
- YAML `compare:` block in suite files (CLI flags only for v0.1)

---

## 4. `run.json` format

```json
{
  "version": 1,
  "suite_id": "serving-smoke",
  "target": "https://gateway.example/v1",
  "cases": [
    {
      "id": "ping-gemini",
      "model": "google/gemini-2.5-flash-lite",
      "passed": true,
      "reason": "",
      "latency_ms": 820,
      "attempts": 1
    }
  ],
  "summary": {
    "total": 3,
    "passed": 3,
    "failed": 0,
    "pass_rate": 100.0,
    "p50_ms": 600,
    "p95_ms": 820
  }
}
```

Rules:

- `version` is required; unknown version → compare rejects (exit 2)
- `pass_rate` = `passed / total * 100`
- `p50_ms` / `p95_ms` computed from **all cases** (passed + failed)
- `reason` uses existing fail reasons: `call_failed`, `invalid_json`, `schema_mismatch`, `content_mismatch`

---

## 5. Compare rules

### 5.1 Comparable runs

Two runs are comparable only if:

- same `suite_id`
- same case `id`s
- same `model` per case

Otherwise → exit `2` with a clear error.

Example: GPT-4 baseline vs GPT-5 candidate → **not comparable**.

### 5.2 Latency gate (`--max-p95-regression`)

Only checked when the flag is set.

```
delta_p95 = (candidate.p95_ms - baseline.p95_ms) / baseline.p95_ms * 100
```

Fail if `delta_p95 > threshold`.

Output:

```
PASS  p95 +8%
```

or

```
FAIL  p95 +87%  (threshold: 20%)
```

### 5.3 Pass rate gate (`--min-pass-rate`)

Only checked when the flag is set.

Fail if `candidate.pass_rate < threshold`.

Output:

```
PASS  pass rate 97%
```

or

```
FAIL  pass rate 89%  (threshold: 95%)
```

---

## 6. Implementation plan

| PR | Scope |
|----|-------|
| **PR-1** | `connor run --out run.json` + schema + tests |
| **PR-2** | `connor compare` + `--max-p95-regression` |
| **PR-3** | `--min-pass-rate` gate |
| **PR-4** | `docs/handbook/ci-regression.md` + CI example |

---

## 7. Decisions (resolved for v0.1)

| Question | Decision |
|----------|----------|
| Thresholds in YAML or CLI? | CLI flags only |
| Default threshold when flag omitted? | Check skipped (no implicit fail) |
| p95 over which cases? | All cases |
| Store `target` in run.json? | Yes (debugging) |

---

## 8. Done when (v0.1 exit criteria)

- [x] `connor run suite.yaml --out run.json` works
- [x] `connor compare` fails on p95 regression in a demo
- [x] `connor compare` respects `--min-pass-rate`
- [ ] Documented in handbook
- [ ] Tag `v0.1.0` published

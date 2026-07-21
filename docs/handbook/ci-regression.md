# CI regression handbook

Use Connor to block merges when an LLM / agent suite **regresses** — lower pass rate or higher p95 latency — versus a known-good baseline.

This guide covers the `run.json` → `compare` workflow (RFC 0001).

---

## What Connor measures

After `connor run suite.yaml --out run.json`, the artifact includes:

| Field | Meaning |
|-------|---------|
| `summary.pass_rate` | `(passed / total) × 100` |
| `summary.p95_ms` | p95 latency over **all** cases (pass + fail) |
| `cases[]` | Per-case id, model, passed, latency |

`connor compare` does **not** re-call the API. It reads two artifacts and applies optional gates.

---

## Local workflow (3 commands)

Use the **same suite** for both runs (same `suite_id`, case ids, and models).

```bash
export CONNOR_BASE_URL=https://your-gateway/v1   # must include /v1
export CONNOR_API_KEY=...                        # if required

mkdir -p runs

# 1. Known-good run (e.g. from main)
connor run benchmarks/examples/agent-json-smoke.yaml --out runs/baseline.json

# 2. Candidate run (e.g. this PR / staging)
connor run benchmarks/examples/agent-json-smoke.yaml --out runs/candidate.json

# 3. Gates
connor compare runs/baseline.json runs/candidate.json \
  --max-p95-regression 20 \
  --min-pass-rate 95
```

### Example output

```text
PASS  p95 +8%
PASS  pass rate 100%
```

or

```text
PASS  p95 +12%
FAIL  pass rate 75%  (threshold: 95%)
```

On p95 FAIL, Connor also prints the **driver** case (id, model, latency delta).

---

## Gates

| Flag | Rule | When omitted |
|------|------|--------------|
| `--max-p95-regression N` | Fail if p95 delta `%` **>** N | Gate skipped |
| `--min-pass-rate N` | Fail if `candidate.pass_rate` **<** N | Gate skipped |

- Both enabled → **AND** (any FAIL → exit 1).
- Thresholds are CLI-only in v0.1 (not in YAML).
- Pass rate uses the **candidate** absolute rate, not a delta vs baseline.

### Exit codes (`connor compare`)

| Code | Meaning |
|------|---------|
| `0` | All enabled gates passed |
| `1` | At least one gate failed |
| `2` | Invalid `run.json`, or runs not comparable |

### Comparable runs (ADR 0001)

Baseline and candidate must share:

- same `suite_id`
- same case `id`s (same order / count)
- same `model` per case

Otherwise → exit `2`. Different models (e.g. GPT-4 vs GPT-5) are **not** comparable.

---

## GitHub Actions

### Minimal smoke (no compare)

```yaml
name: LLM smoke

on: [pull_request]

jobs:
  connor-smoke:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install Connor
        run: |
          curl -L -o connor https://github.com/noodl-labs/Connor/releases/download/v0.1.0/connor-linux-amd64
          chmod +x connor
          sudo mv connor /usr/local/bin/

      - name: Smoke gate
        env:
          CONNOR_BASE_URL: ${{ secrets.CONNOR_BASE_URL }}
          CONNOR_API_KEY: ${{ secrets.CONNOR_API_KEY }}
        run: connor run benchmarks/examples/agent-json-smoke.yaml
```

`connor run` itself exits `1` if any case fails — enough for a basic PR gate.

### Regression compare (baseline + candidate)

**Idea:** save `baseline.json` from `main` (or nightly), download it on PRs, run the suite again, then `compare`.

```yaml
name: LLM regression

on:
  push:
    branches: [main]
  pull_request:

jobs:
  baseline:
    if: github.event_name == 'push' && github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Install Connor
        run: |
          curl -L -o connor https://github.com/noodl-labs/Connor/releases/download/v0.1.0/connor-linux-amd64
          chmod +x connor
          sudo mv connor /usr/local/bin/
      - name: Export baseline
        env:
          CONNOR_BASE_URL: ${{ secrets.CONNOR_BASE_URL }}
          CONNOR_API_KEY: ${{ secrets.CONNOR_API_KEY }}
        run: connor run benchmarks/examples/agent-json-smoke.yaml --out baseline.json
      - uses: actions/upload-artifact@v4
        with:
          name: connor-baseline
          path: baseline.json
          retention-days: 30

  compare:
    if: github.event_name == 'pull_request'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Install Connor
        run: |
          curl -L -o connor https://github.com/noodl-labs/Connor/releases/download/v0.1.0/connor-linux-amd64
          chmod +x connor
          sudo mv connor /usr/local/bin/
      - name: Download baseline
        uses: actions/download-artifact@v4
        with:
          name: connor-baseline
        # Prefer a committed baseline or release asset if cross-workflow
        # artifacts are unavailable — see "Baseline storage" below.
      - name: Export candidate
        env:
          CONNOR_BASE_URL: ${{ secrets.CONNOR_BASE_URL }}
          CONNOR_API_KEY: ${{ secrets.CONNOR_API_KEY }}
        run: connor run benchmarks/examples/agent-json-smoke.yaml --out candidate.json
      - name: Compare gates
        run: |
          connor compare baseline.json candidate.json \
            --max-p95-regression 20 \
            --min-pass-rate 95
```

### Baseline storage (practical options)

GitHub `upload-artifact` is scoped to **one workflow run**. For PR jobs, prefer one of:

1. **Commit** a golden `baselines/smoke.json` updated on `main` (simplest).
2. Upload to **release assets** / S3 / your artifact store; download on PRs.
3. Use a reusable workflow that always re-runs baseline from `main` checkout in the same job before compare.

Example (same job, no external artifact):

```yaml
- name: Baseline from main
  run: |
    git fetch origin main
    git show origin/main:baselines/smoke.json > baseline.json

- name: Candidate
  run: connor run benchmarks/examples/agent-json-smoke.yaml --out candidate.json

- name: Compare
  run: |
    connor compare baseline.json candidate.json \
      --max-p95-regression 20 \
      --min-pass-rate 95
```

---

## Choosing thresholds

| Gate | Starting point | Notes |
|------|----------------|-------|
| `--min-pass-rate` | `95` or `100` | Smoke suites with few cases often use `100` |
| `--max-p95-regression` | `20` | Percent increase vs baseline p95; tune after a week of data |

Omit a flag until you trust the metric — omitted = not enforced.

---

## Common failures

| Symptom | Cause | Fix |
|---------|-------|-----|
| `exit 2` / not comparable | Different suite, case ids, or models | Same YAML suite; don’t swap model slugs mid-compare |
| `exit 1` pass rate | Too many failed cases on candidate | Fix prompts / provider / suite; or lower threshold deliberately |
| `exit 1` p95 | Latency spike | Inspect **driver** line; check model / region / cold start |
| Gate never fails | Flag not passed | Flags are opt-in; add `--min-pass-rate` / `--max-p95-regression` |

---

## Offline dry-run (no API)

Comparable fixtures ship in the repo — no LLM calls:

```bash
connor compare \
  benchmarks/examples/regression-demo/baseline.json \
  benchmarks/examples/regression-demo/candidate.json \
  --max-p95-regression 20 \
  --min-pass-rate 95
echo $?
```

Expected: both gates **FAIL** (p95 ~+150%, pass rate 75%), exit `1`.

Details: [benchmarks/examples/regression-demo/README.md](../../benchmarks/examples/regression-demo/README.md).

---

## See also

- [README — Getting started](../../README.md)
- [RFC 0001 — Run artifact & compare](../rfc/0001-run-artifact-and-compare.md)
- [ADR 0001 — Comparable runs](../adr/0001-comparable-runs.md)
- [Architecture](../architecture.md)

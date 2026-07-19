# Regression demo (offline)

Comparable `run.json` fixtures — **no API required**.

| | Baseline | Candidate |
|-|----------|-----------|
| Pass rate | 100% | 75% |
| p95 | 160 ms | 400 ms (~+150%) |

## Detect regressions (both gates FAIL)

```bash
connor compare \
  benchmarks/examples/regression-demo/baseline.json \
  benchmarks/examples/regression-demo/candidate.json \
  --max-p95-regression 20 \
  --min-pass-rate 95
```

Expected:

```text
Connor  v0.1.0
Compare  regression-demo
        baseline  p95 160ms · pass rate 100%
       candidate  p95 400ms · pass rate 75%

FAIL  p95 +150%  (threshold: 20%)
      driver  d  m  160ms → 400ms  (+150%)
FAIL  pass rate 75%  (threshold: 95%)
────────────────────────────────────
GATE FAILED — do not merge
exit 1
```

Exit code: `1` (CI would block the merge).

## Wide thresholds (both PASS)

```bash
connor compare \
  benchmarks/examples/regression-demo/baseline.json \
  benchmarks/examples/regression-demo/candidate.json \
  --max-p95-regression 200 \
  --min-pass-rate 70
```

Exit code: `0`.

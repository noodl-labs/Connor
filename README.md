# Connor

**CI smoke tests for LLM endpoints and agents.**

Before you merge: does your gateway still respond? Is the JSON still valid? Does your schema still match?

```bash
connor run suite.yaml   # exit 0 = safe to merge · exit 1 = block the PR
```

Your agent stays Python (or any stack). Connor only needs an **OpenAI-compatible HTTP endpoint** (`/v1/chat/completions`).

**Not** a production observability platform (Langfuse, LangSmith). Connor runs **in CI** — like unit tests for your LLM layer.

---

## The problem

Most teams ship LLM features without automated gates in CI:

- Gateway returns HTTP 200 but **invalid JSON**
- Agent output **drifts** after a prompt or model change
- Latency **degrades silently** (p95 doubles) while functional tests still pass

Connor answers: **block the merge before users see it.**

---

## What Connor does

| Capability | Command | CI question |
|------------|---------|-------------|
| **Smoke gate** | `connor run suite.yaml` | Does every case pass right now? |
| **Run artifact** | `connor run suite.yaml --out run.json` | What were latencies and pass rate? |
| **Regression** | `connor compare baseline.json candidate.json` | Did p95 get worse vs baseline? |

**Gates today:** HTTP 2xx · JSON syntax · JSON Schema · text contains · p95 regression (compare).

---

## Getting started

### Requirements

- An OpenAI-compatible API (`CONNOR_BASE_URL` must include `/v1`)
- API key (required for OpenRouter/OpenAI; optional for some local vLLM setups)

### 1. Install

**Option A — Release binary (fastest)**

Download for your platform from [Releases](https://github.com/noodl-labs/Connor/releases) (`v0.1.0-beta.3`):

```bash
# macOS Apple Silicon example
curl -L -o connor https://github.com/noodl-labs/Connor/releases/download/v0.1.0-beta.3/connor-darwin-arm64
chmod +x connor
sudo mv connor /usr/local/bin/   # or keep in ./connor
```

**Option B — From source**

```bash
git clone https://github.com/noodl-labs/Connor.git
cd Connor
make install    # needs Go 1.22+
```

**Option C — `go install`**

```bash
go install github.com/noodl-labs/ConnorLLM/services/runtime/cmd/connor@latest
```

### 2. Configure

```bash
export CONNOR_BASE_URL=https://openrouter.ai/api/v1   # or your gateway / vLLM
export CONNOR_API_KEY=sk-your-key
```

Or copy [`.env.example`](.env.example) and `source` it.

**Self-hosted (vLLM / LiteLLM):**

```bash
export CONNOR_BASE_URL=http://localhost:8000/v1
# CONNOR_API_KEY optional
```

### 3. First run (recommended)

Start with the **agent JSON Schema** suite — stable and CI-friendly:

```bash
connor run benchmarks/examples/agent-json-smoke.yaml
echo $?   # 0 = pass · 1 = fail
```

Add `-v` to see response bodies on passed cases.

**Expected output:**

```text
Connor  v0.1.0-beta.3
Target  https://openrouter.ai/api/v1
Suite   agent-json-smoke (2 cases)

✓  flight-gpt4o-mini  openai/gpt-4o-mini           …  HTTP 200   schema ✓
✓  flight-gemini      google/gemini-2.5-flash-lite …  HTTP 200   schema ✓
────────────────────────────────────
2/2 passed · …
GATE PASSED — safe to merge
exit 0
```

### 4. Single-case mode (no YAML file)

```bash
connor run \
  --model openai/gpt-4o-mini \
  --prompt 'Return only valid JSON: {"status":"ok"}' \
  --expect-json
```

### 5. Regression compare (optional)

Save two runs of the **same suite**, then compare latency:

```bash
mkdir -p runs
connor run benchmarks/examples/agent-json-smoke.yaml --out runs/baseline.json
connor run benchmarks/examples/agent-json-smoke.yaml --out runs/candidate.json

connor compare runs/baseline.json runs/candidate.json --max-p95-regression 20
# PASS or FAIL  p95 +N%  (threshold: 20%)
# On FAIL: driver case id, model, and latency delta are shown
```
- exit 0: compare passed
- exit 1: latency regression failed
- exit 2: invalid or incomparable run files

On failure, Connor shows the driver case, model, and latency delta.
Baseline = your last known-good run (e.g. artifact from `main`). Candidate = current PR run.

---

## GitHub Actions

```yaml
name: LLM gate

on: [pull_request]

jobs:
  connor-smoke:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install Connor
        run: |
          curl -L -o connor https://github.com/noodl-labs/Connor/releases/download/v0.1.0-beta.3/connor-linux-amd64
          chmod +x connor
          sudo mv connor /usr/local/bin/

      - name: LLM smoke gate
        env:
          CONNOR_BASE_URL: ${{ secrets.CONNOR_BASE_URL }}
          CONNOR_API_KEY: ${{ secrets.CONNOR_API_KEY }}
        run: connor run benchmarks/examples/agent-json-smoke.yaml
```

Regression gate (optional second job): store `baseline.json` as a CI artifact from `main`, then `connor compare` on PRs.

---

## Write your own suite

Create `connor/ci/smoke.yaml` in your repo:

```yaml
suite: my-agent-smoke

defaults:
  timeout_ms: 30000
  retries: 2

cases:
  - id: intent-router
    model: gpt-4o-mini              # slug as exposed by YOUR gateway
    prompt: |
      Return only raw JSON, no markdown.
      User: "book a flight to Paris"
      Schema: {"intent":"string","confidence":"number"}
    expect_json_schema:
      type: object
      required: [intent, confidence]
      properties:
        intent: { type: string }
        confidence: { type: number }
```

```bash
export CONNOR_BASE_URL=https://your-staging-gateway/v1
connor run connor/ci/smoke.yaml
```

Point `CONNOR_BASE_URL` at **staging**, not prod. Use the same prompts and schemas your agent uses in production.

---

## Example suites

| File | Purpose | Use in CI? |
|------|---------|------------|
| [`agent-json-smoke.yaml`](benchmarks/examples/agent-json-smoke.yaml) | JSON Schema gate (2 models) | **Yes — start here** |
| [`agent-intent-gate.yaml`](benchmarks/examples/agent-intent-gate.yaml) | Intent router + health ping | Template for real workflows |
| [`serving-smoke.yaml`](benchmarks/examples/serving-smoke.yaml) | Multi-model serving smoke | Post-deploy / exploration |
| [`glm-qwen-smoke.yaml`](benchmarks/examples/glm-qwen-smoke.yaml) | 4-model schema smoke | Multi-provider demo |
| [`agent-json.yaml`](benchmarks/examples/agent-json.yaml) | 1 pass + 2 intentional fails | Local demo only |

---

## Exit codes

| Command | `0` | `1` | `2` |
|---------|-----|-----|-----|
| `connor run` | All cases passed | Any case failed | — |
| `connor compare` | All enabled gates passed | p95 regression failed | Invalid / incomparable `run.json` |

**Fail reasons (stable):** `call_failed` · `invalid_json` · `schema_mismatch` · `content_mismatch`

---

## FAQ

**Connor vs Langfuse / LangSmith?**  
They observe production traffic. Connor runs **before merge** in CI with deterministic gates and `exit 0/1`.

**Connor vs Promptfoo?**  
Promptfoo excels at eval and prompt comparison. Connor focuses on **CI gates**: block merges when HTTP/JSON/schema/latency regresses.

**Do I need to change my agent code?**  
No. Connor calls the same HTTP endpoint your app uses.

**Can I compare different models (GPT-4 vs GPT-5)?**  
No. Compare requires the same suite, case IDs, and models per case (baseline vs candidate = same config, different point in time).

---

## Architecture

Connor is organized into six engines (execution, evaluation, benchmark, quality gates, observability, DX).  
Details: [docs/architecture.md](docs/architecture.md) · [Roadmap](ROADMAP.md) · [Changelog](CHANGELOG.md)

---

## Status

**Current release:** [`v0.1.0-beta.3`](https://github.com/noodl-labs/Connor/releases/tag/v0.1.0-beta.3)

- `connor run` — YAML suites, JSON / schema / contains gates
- `connor run --out run.json` — run artifact
- `connor compare --max-p95-regression` — p95 regression + driver case on FAIL

**Next (`v0.1.0`):** `--min-pass-rate` on compare · CI handbook

---

## License

License TBD.

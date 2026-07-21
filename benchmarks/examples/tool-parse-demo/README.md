# Tool parse demo (RFC 0002 PR-1)

Offline fixture showing what `connor run … --out run.json` exports **after** PR-1:

- per-case `tool_calls[].name`
- per-case `prompt_tokens` / `completion_tokens`
- suite `summary.total_tokens` (= sum of case tokens)

No API call required — this is the **shape** of the artifact, not a live run.

## Inspect

```bash
cat benchmarks/examples/tool-parse-demo/run.json | jq '.summary'
# total_tokens: 440

cat benchmarks/examples/tool-parse-demo/run.json | jq '.cases[].tool_calls'
```

## Verify with unit tests (on branch `feat/CON07-PARSE-TOOL-CALLS`)

```bash
cd services/runtime
go test ./internal/runtime/infrastructure/providers/openai_compatible/ \
  -run 'ParseCompletion_toolCalls|parsesTool' -v
go test ./internal/runtime/domain/entities/ \
  -run 'BuildRunArtifact_toolCalls' -v
```

## Not yet (later PRs)

- YAML `expect_tool` → PR-2
- `connor compare --max-cost-regression` → PR-3

# Agent trace demo (RFC 0003 P1)

Offline fixture showing the **additive** `cases[].trajectory` field on `run.json` version 1.

No API call required. `connor inspect` is not shipped yet — this is the contract the Python SDK writes.

## Shape

```text
run  test-agent
├── tool.search_customer
└── tool.refund
```

```bash
cat benchmarks/examples/agent-trace-demo/run.json | jq '.cases[0].trajectory.spans[].name'
```

## Verify

```bash
cd services/runtime
go test ./internal/runtime/domain/entities/ -run Trajectory -v
```

tmp/ is gitignored. Demo output is a real `run.json` Connor will later `inspect`.

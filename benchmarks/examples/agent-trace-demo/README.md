# Agent trace demo (RFC 0003)

Offline fixture showing the **additive** `cases[].trajectory` field on `run.json` version 1.

No API call required.

```bash
connor inspect benchmarks/examples/agent-trace-demo/run.json
```

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
make demo-inspect
cd services/runtime && go test ./internal/runtime/domain/entities/ -run Trajectory -v
```

tmp/ is gitignored. `make demo-trace` / `demo-trace-complex` write live artifacts you can inspect the same way.

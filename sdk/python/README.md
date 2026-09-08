# Connor Python SDK (RFC 0003 P1)

Instrument your agent. Connor's Go CLI stays the CI contract.

```python
from connor import trace, tool

@tool
async def search_customer(customer_id: str) -> dict:
    return {"name": "Ada"}

with trace("test-agent") as run:
    await search_customer("123")
    run.write("run.json")  # version 1 + cases[].trajectory
```

`@tool` is a no-op without an active `trace()` — safe to leave on production callables.

## Tests

From the **repo root** (not from `sdk/python`):

```bash
make test-python
make demo-trace                 # writes tmp/support-agent.json
make demo-trace ARGS='--fail'   # email tool raises; still records the run

go test -C services/runtime ./internal/runtime/domain/entities/ -run Trajectory -v
```

If you already `cd sdk/python`, the Go path is `../../services/runtime`.

No OpenTelemetry dependency. Nested `trace()` is rejected in V1.

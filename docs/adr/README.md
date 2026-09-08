# ADRs (Architecture Decision Records)

Short, **immutable decisions** — not full designs (those live in RFCs).

## When to write an ADR

- Comparable runs rules (0001)
- Cost gate = token delta (0002)
- Trace model vs OpenTelemetry (0003)
- `run.json` stays version 1 with additive trajectory (0004)
- Go vs Python: Go = CI contract; Python SDK = instrumentation (RFC 0003); Python eval service remains v1 semantic judges

Skip ADR for: implementation details covered entirely by RFC.

## Template

```markdown
# ADR NNNN — Title

## Status
Accepted | Superseded by ADR NNNN

## Context
## Decision
## Consequences
```

## Index

See [traceability.md](../traceability.md#adr-index).

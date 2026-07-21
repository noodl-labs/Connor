# RFCs (Request for Comments)

Design documents **before** implementation. Status flow: `Draft` → `Accepted` → `Implemented` (via PR merge).

## When to write an RFC

- New CLI subcommand or flag with CI semantics
- New YAML case fields or gates
- Changes to `run.json` schema or exit codes
- New compare dimensions (cost, tools, semantic scores)

Do **not** RFC: bug fixes, refactors, docs-only, Lipgloss tweaks.

## Naming

```
docs/rfc/NNNN-short-kebab-title.md
```

Next number: **0003** (after 0002 ships or is superseded).

## Template

```markdown
# RFC NNNN — Title

| | |
|---|---|
| **Status** | Draft \| Accepted \| Superseded |
| **Target** | v0.x.0 |
| **Issues** | #N |
| **Roadmap** | UC #N |

## 1. Problem
## 2. Goal
## 3. Out of scope
## 4. API / YAML / CLI / schema
## 5. Exit codes & fail reasons
## 6. Implementation plan (PRs)
## 7. Decisions
## 8. Done when
```

## Index

See [traceability.md](../traceability.md#rfc-index).

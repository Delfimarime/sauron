# Data Model: Configuration — Sauron Settings (Persona)

**Spec**: [Import Persona](../spec.md)

Describes the persona definition file that the Import Persona feature reads and the persisted configuration it writes.

## Persona definition file (input)

A YAML document supplied by the user via `<path>`:

| Field | Type | Required | Constraints | Description |
|-------|------|----------|-------------|-------------|
| `name` | string | Yes | slug `^[a-z0-9]+(-[a-z0-9]+)*$`; unique across personas | Persona identity. Realizes [spec](../spec.md) FR-002, FR-013, FR-016. |
| `description` | string | Yes | non-empty | Who the persona is for. Realizes [spec](../spec.md) FR-014. |
| `tags` | list of string | No | — | Labels used for filtering. |
| `agents` | list of string | No* | artifact names | Agents the persona delivers. |
| `skills` | list of string | No* | artifact names | Skills the persona delivers. |

\* At least one entry is required across `agents` and `skills` combined. Realizes [spec](../spec.md) FR-015.

The file never carries a priority — priority comes from `--priority` at import and is managed afterwards by [set priority persona](../../0010-set-persona-priority/spec.md) (see [ADR-0001](../architecture/ADR-0001-persona-priority-model.md)).

Example:

```yaml
name: backend-developer
description: Backend developers working on Go services
tags:
  - backend
  - golang
agents:
  - software-engineer
skills:
  - design-oas3
  - code-review
```

## Persisted configuration — `~/.sauron/settings.yaml`

- **Path**: `~/.sauron/settings.yaml` (home directory resolved per platform).
- **Format**: a single YAML document.
- **Lifecycle**: the `personas` section is created on first successful import if absent. Realizes [spec](../spec.md) FR-008.

Top-level document:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `personas` | array of Persona | No | Imported personas. Absent or empty when none. |

Persona entry:

| Field | Type | Required | Constraints | Description |
|-------|------|----------|-------------|-------------|
| `name` | string | Yes | slug; unique across personas | Persona identity. Realizes [spec](../spec.md) FR-002. |
| `description` | string | Yes | non-empty | From the definition file. |
| `tags` | list of string | No | — | From the definition file. |
| `agents` | list of string | No* | — | From the definition file. |
| `skills` | list of string | No* | — | From the definition file. |
| `priority` | integer | No | `0` for the first persona; ≥ 1 and unique among defined values otherwise; absent = undefined | Precedence; lower = higher. Realizes [spec](../spec.md) FR-003, FR-006, FR-020, FR-007. |

## Identity

A persona is identified by its `name`, unique across all personas. Defined priorities are unique among personas; undefined priorities may repeat (see [ADR-0001](../architecture/ADR-0001-persona-priority-model.md)).

## Write semantics

- The whole document is loaded, the new persona appended, and the document written back only after all validation passes. The file is left untouched on any failure. Realizes [spec](../spec.md) FR-010.
- Writes are atomic: serialize to a temporary file in `~/.sauron/`, then rename over `settings.yaml`.

## Example

```yaml
personas:
  - name: backend-developer        # first import → priority 0
    description: Backend developers working on Go services
    priority: 0
    tags: [backend, golang]
    agents: [software-engineer]
    skills: [design-oas3, code-review]
  - name: qa-engineer              # imported without --priority
    description: QA engineers validating releases
    tags: [qa]
    skills: [test-plan]
```

# Describe Registry — state

This feature **reads** `settings.yaml` and writes nothing. The document schema is
owned by the [state data contract](../../contracts/state.md).

## Reads

- File: `settings.yaml` (which holds the single `Registry` document).
- Surfaces all fields of the `Registry`, with credential fields rendered as their
  stored environment reference.

## Field realization

| Field | Requirement |
|---|---|
| `spec.uri` | FR-001 |
| `spec.auth.*` | FR-002 |
| all `spec.*` | FR-003 |

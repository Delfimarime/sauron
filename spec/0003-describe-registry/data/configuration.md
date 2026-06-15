# Describe Registry — configuration

This feature **reads** `registries.yaml` and writes nothing. The document schema
is owned by the [configuration data contract](../../contracts/configuration.md).

## Reads

- File: `registries.yaml` (a stream of `Registry` documents).
- Surfaces all fields of the matched `Registry`, with credential fields rendered
  as their stored environment reference.

## Field realization

| Field | Requirement |
|---|---|
| `metadata.name` | FR-001 |
| `spec.auth.*` | FR-002 |
| all `spec.*` | FR-003 |

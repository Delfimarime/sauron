# List Registries — configuration

This feature **reads** `registries.yaml` and writes nothing. The document schema
is owned by the [configuration data contract](../../contracts/configuration.md).

## Reads

- File: `registries.yaml` (a stream of `Registry` documents).
- Fields surfaced as columns: `metadata.name`, `spec.transport`, `spec.uri`,
  `spec.timeout`.

## Field realization

| Field | Requirement |
|---|---|
| `metadata.name` | FR-001, FR-002, FR-003 |
| `spec.transport` | FR-002, FR-004 |

# List Catalogue — state

This feature **reads** `settings.yaml` to resolve the registry's connection and
writes nothing. The catalogue itself is not persisted. The document schema is
owned by the [state data contract](../../contracts/state.md).

## Reads

- File: `settings.yaml`.
- Uses the configured `Registry`'s `spec.transport`, `spec.uri`, `spec.auth`,
  `spec.tls`, `spec.sshKey`, and `spec.timeout` to fetch the live catalogue.

## Field realization

| Field | Requirement |
|---|---|
| `spec.transport`, `spec.uri`, connection fields | FR-001, FR-005 |
| `Registry` presence | FR-006 |

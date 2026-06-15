# List Catalogue — configuration

This feature **reads** `registries.yaml` to resolve the registry's connection and
writes nothing. The catalogue itself is not persisted. The document schema is
owned by the [configuration data contract](../../contracts/configuration.md).

## Reads

- File: `registries.yaml` (a stream of `Registry` documents).
- Uses the matched `Registry`'s `spec.transport`, `spec.uri`, `spec.auth`,
  `spec.tls`, `spec.sshKey`, and `spec.timeout` to fetch the live catalogue.

## Field realization

| Field | Requirement |
|---|---|
| `metadata.name` | FR-001, FR-006 |
| `spec.transport`, `spec.uri`, connection fields | FR-001, FR-005 |

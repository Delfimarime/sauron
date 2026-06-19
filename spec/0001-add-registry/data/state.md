# Add Registry — state

This feature **writes** `registries.yaml`, appending one `Registry` document. The
document schema is owned by the
[state data contract](../../contracts/state.md); only the
read/write semantics and the field→requirement realization are stated here.

## Writes

- File: `registries.yaml` (a stream of `Registry` documents).
- Operation: append a new `Registry`; the write is atomic and lock-guarded.

## Field realization

| Field | Requirement |
|---|---|
| `metadata.name` | FR-001, FR-008 |
| `spec.uri` | FR-001 |
| `spec.transport` | FR-001, FR-002 |
| `spec.auth.username` / `spec.auth.password` | FR-003, FR-011 |
| `spec.tls.*`, `spec.sshKey` | FR-011 |
| `spec.timeout` | FR-012 |
| `spec.ref` | FR-013 |

A registry of an existing `metadata.name` is rejected (FR-007); the existing
document is left unchanged until validation succeeds (FR-006).

# Set Registry — state

This feature **writes** `settings.yaml`, setting the single `Registry` document.
The document schema is owned by the
[state data contract](../../contracts/state.md); only the read/write semantics and
the field→requirement realization are stated here.

## Writes

- File: `settings.yaml` (which holds the single `Registry` document alongside the
  `Provider`).
- Operation: set the `Registry`, creating it or replacing the existing one; the
  write is atomic and lock-guarded.

## Field realization

| Field | Requirement |
|---|---|
| `spec.source` | FR-001 |
| `spec.transport` | FR-001, FR-002 |
| `spec.credentials.username` / `spec.credentials.password` | FR-003, FR-011 |
| `spec.tls.*`, `spec.sshKey` | FR-011 |
| `spec.timeout` | FR-012 |
| `spec.revision` | FR-013 |
| `metadata.createdAt` / `metadata.lastUpdatedAt` | FR-014 |

An existing `Registry` is replaced (FR-007); the existing document is left
unchanged until validation succeeds (FR-006).

# Set Provider — state

This feature **writes** `settings.yaml` and **writes** `track.yaml` on migration.
The document schema is owned by the
[state data contract](../../contracts/state.md).

## Writes

- `settings.yaml`: upsert the single `Provider` document.
- `track.yaml`: on a provider change, update each migrated artifact's `spec.path`
  to the new provider's directory. Writes are atomic and lock-guarded.

## Field realization

| Field | Requirement |
|---|---|
| `Provider.metadata.name` | FR-001, FR-003, FR-004 |
| `spec.path` (each artifact) | FR-002 |

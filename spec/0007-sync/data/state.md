# Sync — state

This feature **reads** `settings.yaml` and **writes** `track.yaml`. The document
schema is owned by the [state data contract](../../contracts/state.md).

## Reads

- `settings.yaml`: the registry's transport and connection, and the `Provider`
  document (must exist).

## Writes

- `track.yaml`: update changed artifacts and remove vanished ones; the write is
  atomic and lock-guarded. `--dry-run` writes nothing.

## Field realization

| Field | Requirement |
|---|---|
| `spec.digest` | FR-001, FR-002 |
| `spec.path` (drift repair / removal) | FR-002, FR-003 |
| `Provider` presence | FR-008 |

# Upgrade — state

This feature **reads** `settings.yaml` and **writes** `track.yaml`. The document
schema is owned by the [state data contract](../../contracts/state.md).

## Reads

- `settings.yaml`: the registry's transport and connection, and the `Provider`
  document (must exist).

## Writes

- `track.yaml`: update changed artifacts; never removes. The write is atomic and
  lock-guarded. `--dry-run` writes nothing.

## Field realization

| Field | Requirement |
|---|---|
| `spec.digest` | FR-001 |
| (no removal) | FR-002 |
| `Provider` presence | FR-007 |

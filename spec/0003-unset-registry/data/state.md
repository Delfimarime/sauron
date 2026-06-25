# Unset Registry — state

This feature **writes** `settings.yaml` and leaves `track.yaml` untouched. The
document schema is owned by the
[state data contract](../../contracts/state.md).

## Writes

- `settings.yaml`: remove the `Registry` document.

The write is atomic and lock-guarded; `--dry-run` performs none. Tracked artifacts
in `track.yaml` and in the provider are preserved (FR-002).

## Field realization

| Field | Requirement |
|---|---|
| `Registry` document | FR-001 |

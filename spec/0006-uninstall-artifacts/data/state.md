# Uninstall Artifacts — state

This feature **writes** `track.yaml` and reads nothing else. The document schema
is owned by the [state data contract](../../contracts/state.md).

## Writes

- `track.yaml`: remove the matched `Skill`/`Agent` documents; the write is atomic
  and lock-guarded. `--dry-run` writes nothing.

## Field realization

| Field | Requirement |
|---|---|
| `metadata.name`, `spec.path` | FR-001 |

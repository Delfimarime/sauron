# Sync — state

This feature **reads** `registries.yaml` and `settings.yaml` and **writes**
`track.yaml`. The document schema is owned by the
[state data contract](../../contracts/state.md).

## Reads

- `registries.yaml`: each tracked artifact's source connection.
- `settings.yaml`: the `Provider` document (must exist).

## Writes

- `track.yaml`: update changed artifacts, remove vanished ones, and update each
  persona's `members` snapshot; the write is atomic and lock-guarded. `--dry-run`
  writes nothing.

## Field realization

| Field | Requirement |
|---|---|
| `spec.digest` | FR-001, FR-002 |
| `spec.path` (drift repair / removal) | FR-002, FR-003 |
| `spec.members`, member `spec.provenance` | FR-004 |
| `Provider` presence | FR-008 |

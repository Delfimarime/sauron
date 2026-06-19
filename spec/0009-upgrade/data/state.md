# Upgrade — state

This feature **reads** `registries.yaml` and `settings.yaml` and **writes**
`track.yaml`. The document schema is owned by the
[state data contract](../../contracts/state.md).

## Reads

- `registries.yaml`: each tracked artifact's source connection.
- `settings.yaml`: the `Provider` document (must exist).

## Writes

- `track.yaml`: update changed artifacts and add new persona members; never
  removes. The write is atomic and lock-guarded. `--dry-run` writes nothing.

## Field realization

| Field | Requirement |
|---|---|
| `spec.digest` | FR-001 |
| (no removal) | FR-002 |
| `spec.members` (additions only), member `spec.provenance` | FR-003 |
| `Provider` presence | FR-007 |

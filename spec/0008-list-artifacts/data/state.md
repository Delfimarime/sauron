# List Artifacts — state

This feature **reads** `track.yaml` and writes nothing. The document schema is
owned by the [state data contract](../../contracts/state.md).

## Reads

- File: `track.yaml` (a stream of `Skill`/`Agent` documents), filtered to the
  chosen kind.
- Fields surfaced as columns: `metadata.name`, `spec.version`,
  `spec.updatedAt`.

## Field realization

| Field | Requirement |
|---|---|
| `metadata.name` | FR-001, FR-002, FR-004 |
| `spec.updatedAt` | FR-005 |

# Describe Artifact — state

This feature **reads** `track.yaml` and writes nothing. The document schema is
owned by the [state data contract](../../contracts/state.md).

## Reads

- File: `track.yaml`, matching the `Skill`/`Agent` document of the chosen kind and
  name.
- Surfaces all `spec` fields.

## Field realization

| Field | Requirement |
|---|---|
| all `spec.*` | FR-001, FR-002 |

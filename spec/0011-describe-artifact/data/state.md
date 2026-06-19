# Describe Artifact — state

This feature **reads** `track.yaml` and writes nothing. The document schema is
owned by the [state data contract](../../contracts/state.md).

## Reads

- File: `track.yaml`, matching the `Skill`/`Agent`/`Persona` document of the chosen
  kind and name.
- Surfaces all `spec` fields; for a persona, `spec.members`.

## Field realization

| Field | Requirement |
|---|---|
| all `spec.*` | FR-001, FR-003 |
| `spec.members` | FR-002 |

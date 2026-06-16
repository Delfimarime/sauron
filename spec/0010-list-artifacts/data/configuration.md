# List Artifacts — configuration

This feature **reads** `track.yaml` and writes nothing. The document schema is
owned by the [configuration data contract](../../contracts/configuration.md).

## Reads

- File: `track.yaml` (a stream of `Skill`/`Agent`/`Persona` documents), filtered
  to the chosen kind.
- Fields surfaced as columns: `metadata.name`, `spec.registry`, `spec.version`,
  `spec.provenance`, `spec.updatedAt`, and for personas `spec.members`.

## Field realization

| Field | Requirement |
|---|---|
| `metadata.name` | FR-001, FR-002, FR-003 |
| `spec.registry` | FR-002, FR-004 |
| `spec.updatedAt` | FR-004 |

# Uninstall Artifacts — configuration

This feature **writes** `track.yaml` and reads nothing else. The document schema
is owned by the [configuration data contract](../../contracts/configuration.md).

## Writes

- `track.yaml`: remove the matched `Skill`/`Agent`/`Persona` documents, or update a
  member's `spec.provenance.personas` when a persona is uninstalled but the member
  is retained; the write is atomic and lock-guarded. `--dry-run` writes nothing.

## Field realization

| Field | Requirement |
|---|---|
| `metadata.name`, `spec.registry`, `spec.path` | FR-001 |
| `spec.provenance.personas`, `spec.provenance.direct` | FR-002 |

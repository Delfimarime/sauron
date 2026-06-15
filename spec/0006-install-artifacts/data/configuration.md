# Install Artifacts — configuration

This feature **reads** `registries.yaml` and `settings.yaml` and **writes**
`track.yaml`. The document schema is owned by the
[configuration data contract](../../contracts/configuration.md).

## Reads

- `registries.yaml`: the source `Registry`'s transport and connection.
- `settings.yaml`: the `Provider` document (must exist) for the install
  destination.

## Writes

- `track.yaml`: add or update a `Skill`/`Agent`/`Persona` document per installed
  artifact; the write is atomic and lock-guarded.

## Field realization

| Field | Requirement |
|---|---|
| `metadata.name`, `spec.registry`, `spec.path` | FR-001 |
| `spec.digest`, `spec.version`, `spec.installedAt`, `spec.updatedAt` | FR-002 |
| `spec.provenance.direct` | FR-003 |
| `spec.members`, member `spec.provenance.personas` | FR-004 |
| existing-document reconcile | FR-005 |
| `Provider` presence | FR-007 |

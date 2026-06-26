# Install Artifacts — state

This feature **reads** `settings.yaml` and **writes** `track.yaml`. The document
schema is owned by the [state data contract](../../contracts/state.md).

## Reads

- `settings.yaml`: the configured `Registry`'s transport and connection, and the
  `Provider` document (must exist) for the install destination.

## Writes

- `track.yaml`: add or update a `Skill`/`Agent` document per installed
  artifact; the write is atomic and lock-guarded.

## Field realization

| Field | Requirement |
|---|---|
| `metadata.name`, `spec.path` | FR-001 |
| `spec.digest`, `spec.version`, `spec.installedAt`, `spec.updatedAt` | FR-002 |
| existing-document reconcile | FR-004 |
| `Provider` presence | FR-006 |

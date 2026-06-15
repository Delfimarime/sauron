# Delete Registry — configuration

This feature **writes** both `registries.yaml` and `track.yaml`. The document
schema is owned by the
[configuration data contract](../../contracts/configuration.md).

## Writes

- `registries.yaml`: remove the matched `Registry` document.
- `track.yaml`: remove every `Skill`, `Agent`, and `Persona` document whose
  `spec.registry` equals the deleted registry.

Both writes are atomic and lock-guarded; `--dry-run` performs neither.

## Field realization

| Field | Requirement |
|---|---|
| `metadata.name` (Registry) | FR-001, FR-005 |
| `spec.registry` (tracked artifacts) | FR-002, FR-003 |

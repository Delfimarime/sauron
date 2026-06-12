# Data Model: Configuration — List Artifacts (track.yaml + live registries)

**Spec**: [List Artifacts](../spec.md)

List Artifacts reads the managed artifacts from the
[track file](../../contracts/configuration.md#trackyaml) and, for `--available`,
reads registries live; it never writes. The schema is owned by the
[configuration data contract](../../contracts/configuration.md#trackyaml); this
document does not restate it.

## Reads

- `track.yaml` `items` — the managed artifacts of the requested type, each
  contributing `name`, source `registry`, `pinned`, `provider`, `persona`, `path`.
- `registries.yaml` `items` — for `--available`, the registries to query and the
  `priority` used to resolve the winning registry (pin then priority,
  [ADR-0001](../../0006-sync-artifacts/architecture/ADR-0001-conflict-resolution-by-registry-priority.md)),
  read live ([#registriesyaml](../../contracts/configuration.md#registriesyaml)).
- `personas.yaml` `items` — for the resolved `--available` catalog, to scope the
  desired set ([#personasyaml](../../contracts/configuration.md#personasyaml)).

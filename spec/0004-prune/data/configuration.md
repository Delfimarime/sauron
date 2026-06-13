# Data Model: Configuration — Prune (registries.yaml, track.yaml)

**Spec**: [Prune](../spec.md)

This feature reads the registered registry names from `registries.yaml` and the
installed artifacts from `track.yaml`'s `items`, then removes the orphaned
`track.yaml` entries (and their artifacts); it never writes `registries.yaml`.
The schema is owned by the
[configuration data contract](../../contracts/configuration.md#trackyaml); this
document does not restate it.

## Reads

- `registries.yaml` `items`: the `name` of each registered registry —
  the registered set against which orphans are detected
  ([#registriesyaml](../../contracts/configuration.md#registriesyaml)).
- `track.yaml` `items`: each installed artifact's `type`, `name`,
  `provider`, `path`, and `registry`; an entry is orphaned when its `registry`
  is not in the registered set. Realizes [spec](../spec.md) FR-004.

## Owns

- Nothing. `track.yaml` is owned by
  [sync artifacts](../../0006-sync-artifacts/spec.md).

## Writes

- `track.yaml` `items`: removes the orphaned entries and deletes their
  artifacts from `path`; with `--dry-run` nothing is deleted or written.
  Realizes [spec](../spec.md) FR-005, FR-012. Atomic single-file write per the
  [configuration data contract](../../contracts/configuration.md#cross-file-write-semantics).

# Data Model: Configuration — Pin Artifact (track.yaml)

**Spec**: [Pin Artifact](../spec.md)

Pin Artifact writes the `pinned` (and `registry`) fields of an artifact's
[track file](../../contracts/configuration.md#trackyaml) entry; it does not own
`track.yaml` (owned by [sync artifacts](../../0006-sync-artifacts/spec.md)). The
schema is owned by the
[configuration data contract](../../contracts/configuration.md#trackyaml); this
document does not restate it.

## Reads

- `track.yaml` `items` — to locate the entry for the named artifact and read its
  current `registry`/`pinned`.
- `registries.yaml` `items` — to validate the target registry exists and offers
  the named artifact ([#registriesyaml](../../contracts/configuration.md#registriesyaml)).

## Writes

- `track.yaml` `items[].pinned` and `items[].registry` — `pin` sets `registry` to
  the target and `pinned: true`; `unpin` clears `pinned`. Only the named entry is
  changed. With `--reconcile`, the affected artifact is reconciled to its source
  registry (and installed if absent). Realizes [spec](../spec.md) FR-002, FR-003,
  FR-006, FR-012.

## Notes

There is no separate pins file: the pin lives on the artifact's `track.yaml`
entry, so the track file alone shows which artifacts are pinned. Resolution of a
pin over priority during sync is governed by
[ADR-0002](../../0006-sync-artifacts/architecture/ADR-0002-pins-override-priority.md).

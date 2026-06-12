# Data Model: Configuration — Set Provider (settings.yaml, track.yaml)

**Spec**: [Set Provider](../spec.md)

This feature owns the active `provider` in `settings.yaml` and, when the provider
changes, rewrites the migrated entries of `track.yaml`. The schema of both files
is owned by the
[configuration data contract](../../contracts/configuration.md#settingsyaml);
this document does not restate it.

## Reads

- `settings.yaml` `provider` — the current active provider (absent means
  `claude`).
- `track.yaml` `items` — the artifacts to migrate are the entries whose
  `provider` is the previous active provider; entries already on other providers
  are not touched.

## Owns

- `settings.yaml` `provider` — the single active provider.

## Writes

- `track.yaml` `items` — on migration, each affected entry's `provider`
  and `path` are rewritten to the new provider (move, default); with
  `--copy-only`, a new entry is added for the new provider and the previous one
  is kept.
- `settings.yaml` `provider` — set to the new value.
- **Write order.** Per the contract's
  [cross-file write semantics](../../contracts/configuration.md#cross-file-write-semantics),
  `track.yaml` is written **before** `settings.yaml`, and the operation is
  idempotent: a run interrupted after `track.yaml` is fully repaired by
  re-running `set provider` with the same target, which migrates nothing further
  and completes `settings.yaml`. With no tracked artifacts, only `settings.yaml`
  is written. When the new provider equals the current one, neither file is
  written (no-op).

## Realizes

- `settings.yaml` `provider` write → [spec](../spec.md) FR-002, FR-006 (persist
  the single active provider), FR-008 (transactional), FR-011 (no-op when
  unchanged).
- `track.yaml` migration → [spec](../spec.md) FR-004, FR-005 (relocate affected
  entries), FR-014 (`--copy-only` adds a new entry and keeps the previous).

## Notes

Configuration is now split across files per the
[configuration data contract](../../contracts/configuration.md); file references
updated accordingly.

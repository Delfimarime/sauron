# Data Model: Configuration — Sync Personas

**Spec**: [Sync Personas](../spec.md)

Describes the parts of the [settings](../../0012-backend/spec.md)
this feature writes when refreshing the catalog. The `catalog` and
`personaRegistry` schema — field definitions, location, and format — is owned
by the
[backend data model](../../0012-backend/data/configuration.md);
this document does not redefine it, it specifies only the fields sync personas
writes and the semantics of those writes.

## Inputs

Read for the refresh:

- The full persona catalog offered by the configured
  [backend](../../0012-backend/spec.md) backend, including each
  definition's backend `lastModifiedAt`.
- The local `catalog` and `personaRegistry` settings, to compute which entries
  are new, changed, or absent upstream. See the
  [backend data model](../../0012-backend/data/configuration.md).

## Fields written

This feature writes only the following fields; every other field of the
`catalog`/`personaRegistry` schema is left untouched.

| Field | Owned by | Written when | Value |
|-------|----------|--------------|-------|
| `catalog[].lastModifiedAt` | [backend](../../0012-backend/data/configuration.md) | a catalog entry is added or updated | the backend's last-modified time for that definition. Realizes [spec](../spec.md) FR-003. |
| `catalog[].lastSyncedAt` | [backend](../../0012-backend/data/configuration.md) | a catalog entry is added or updated | the current time. Realizes [spec](../spec.md) FR-003. |
| `personaRegistry.lastSyncedAt` | [backend](../../0012-backend/data/configuration.md) | `sync personas` completes a full pull | the current time. Realizes [spec](../spec.md) FR-004. |

## Operation

- `sync personas` compares the registry's catalog against the local `catalog`;
  new definitions are added and changed ones updated, each stamping
  `lastModifiedAt` and `lastSyncedAt`, and `personaRegistry.lastSyncedAt` is set
  for the whole pull. Realizes [spec](../spec.md) FR-002, FR-003, FR-004.
- `sync persona <name>` performs the same write scoped to the one named entry;
  it does not set `personaRegistry.lastSyncedAt`, which records a *full* pull.
  Realizes [spec](../spec.md) FR-007.
- Without `--force`, an installed persona absent upstream is reported but its
  install is kept — no field is removed. With `--force`, such a persona is
  uninstalled and its catalog entry removed. Realizes [spec](../spec.md) FR-005,
  FR-006, FR-008.
- Without `--force`, when the registry's catalog already matches the local
  catalog, no field is written. Realizes [spec](../spec.md) FR-009.

## Write semantics

- This feature writes only the `catalog` and `personaRegistry` settings; it
  never writes the track file.
- Updates are atomic: serialize the settings to a temporary file in
  `~/.sauron/`, then rename over the settings file, so a refresh interrupted
  mid-write never leaves a partially updated catalog.

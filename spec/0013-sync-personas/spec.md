# Sync Personas

**Type:** feature
**Depends on:** [backend](../0012-backend/spec.md)

## Overview

A person responsible for a team's agentic-AI setup needs the local
[catalog](../0012-backend/spec.md) of persona definitions to reflect
what the configured [backend](../0012-backend/spec.md)
currently offers, so that decisions about which personas to install are made
against up-to-date definitions. This feature pulls persona *definitions* from
the registry into the local catalog; it neither installs nor delivers
artifacts. Artifact delivery remains the job of
[sync artifacts](../0006-sync-artifacts/spec.md), which is independent of this feature.

Two commands cover the need: `sync personas` refreshes the whole catalog, and
`sync persona <name>` refreshes a single definition by name.

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to refresh the local catalog of
  persona definitions from the configured
  [backend](../0012-backend/spec.md).

### Event-driven

- **FR-002**: When `sync personas` runs, Sauron shall pull the full catalog
  from the [backend](../0012-backend/spec.md), add catalog
  entries for definitions not yet present, and update catalog entries whose
  definitions changed upstream.
- **FR-003**: When a catalog entry is added or updated, Sauron shall record its
  `lastModifiedAt` from the registry backend and set its `lastSyncedAt` to the
  current time.
- **FR-004**: When `sync personas` completes, Sauron shall set
  `personaRegistry.lastSyncedAt` to the current time and report the catalog
  entries that were added, updated, and removed.
- **FR-005**: When an [installed persona](../0012-backend/spec.md) is
  no longer offered by the registry, Sauron shall report it and keep the
  install in place.
- **FR-006**: When `--force` is provided to `sync personas`, Sauron shall
  re-pull the full catalog authoritatively — ignoring any "unchanged"
  short-circuit — and hard-reconcile by uninstalling every
  [installed persona](../0012-backend/spec.md) no longer present in
  the registry.
- **FR-007**: When `sync persona <name>` runs, Sauron shall refresh only the
  named persona's definition, applying the same add/update and timestamp rules
  scoped to that one entry.
- **FR-008**: When `--force` is provided to `sync persona <name>`, Sauron shall
  re-pull that definition authoritatively and, if the registry no longer offers
  it, uninstall it when it is an
  [installed persona](../0012-backend/spec.md).

### State-driven

- **FR-009**: While `sync personas` runs without `--force` and the registry's
  catalog already matches the local catalog, Sauron shall make no change,
  report that the catalog is already up to date, and exit successfully.

### Unwanted behavior

- **FR-010**: If no [backend](../0012-backend/spec.md) is
  configured, then Sauron shall reject the request and report that a
  backend must be set first.
- **FR-011**: If the [backend](../0012-backend/spec.md)
  cannot be reached, then Sauron shall reject the request and report that it is
  unreachable.
- **FR-012**: If the settings or the
  [catalog](../0012-backend/spec.md) cannot be read or parsed, then
  Sauron shall reject the request and report that it cannot be read.
- **FR-013**: If `sync persona <name>` names a persona the registry does not
  offer, then Sauron shall report that the persona is not found and exit with
  an error.
- **FR-014**: If a command fails, then Sauron shall write exactly one
  human-readable message to stderr.

## Key Entities

- **Catalog Entry**: a persona definition mirrored locally from the
  [backend](../0012-backend/spec.md), carrying its
  `lastModifiedAt` (from the backend) and `lastSyncedAt` (when it was last
  pulled). The catalog schema is owned by the
  [backend data model](../0012-backend/data/configuration.md).
- **Backend**: the configured backend that owns persona definitions
  (singleton per instance); this feature records its `lastSyncedAt` after a
  full pull.
- **Installed Persona**: a catalog persona activated locally; under `--force`
  it is uninstalled when no longer present in the registry, otherwise kept and
  reported (FR-005, FR-006, FR-008).

## Notes

- `sync personas` (catalog definitions) and
  [sync artifacts](../0006-sync-artifacts/spec.md) (artifact delivery) are independent
  operations: one refreshes persona *definitions* in the catalog, the other
  delivers *artifacts* to the active provider. Neither triggers the other.
- `--dry-run` is intentionally **not** offered by either command; the only flag
  is `--force`. The catalog is a read-only mirror, so a non-`--force` pull is
  already safe — it never uninstalls anything.

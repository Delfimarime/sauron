# Sync Personas

**Type:** feature

**Depends on:** [backend](../0012-backend/spec.md)

## Overview

A person responsible for a team's agentic-AI setup needs the stored definitions
of the [installed personas](../0014-select-personas/spec.md) to reflect what the
configured [backend](../0012-backend/spec.md) currently offers, so that artifact
sync and persona decisions run against up-to-date definitions. This feature
refreshes the stored definition of each installed persona from the backend; it
neither installs nor delivers artifacts. There is **no persisted catalog** — the
set of *available* personas is a
[live view](../contracts/configuration.md#live-persona-view) assembled at command
time, so this feature only updates what is already installed. Artifact delivery
remains the job of [sync artifacts](../0006-sync-artifacts/spec.md), which is
independent of this feature.

Two commands cover the need: `sync personas` refreshes every installed persona's
stored definition, and `sync persona <name>` refreshes one installed persona's
definition by name.

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to refresh the stored definitions
  of the [installed personas](../0014-select-personas/spec.md) from the
  configured [backend](../0012-backend/spec.md).

### Event-driven

- **FR-002**: When `sync personas` runs, Sauron shall, for each
  [installed persona](../0014-select-personas/spec.md), fetch its current
  definition from the [backend](../0012-backend/spec.md) and update its stored
  `description`, `tags`, `skills`, and `agents` in
  [personas.yaml](../contracts/configuration.md#personasyaml) when the upstream
  definition has changed.
- **FR-003**: When an [installed persona](../0014-select-personas/spec.md)'s
  stored definition is refreshed, Sauron shall record its `last_modified_at`
  from the [backend](../0012-backend/spec.md) and set its `last_synced_at` to the
  current time.
- **FR-004**: When `sync personas` completes a full refresh, Sauron shall set
  [backend.yaml](../contracts/configuration.md#backendyaml)'s `last_synced_at` to
  the current time and report the
  [installed personas](../0014-select-personas/spec.md) that were refreshed and
  those that were left unchanged.
- **FR-005**: When an [installed persona](../0014-select-personas/spec.md) is no
  longer offered by the [backend](../0012-backend/spec.md), Sauron shall report
  it and keep the install in place.
- **FR-006**: When `--force` is provided to `sync personas`, Sauron shall re-pull
  every [installed persona](../0014-select-personas/spec.md)'s definition
  authoritatively — ignoring any "unchanged" short-circuit — and hard-reconcile
  by uninstalling every installed persona the
  [backend](../0012-backend/spec.md) no longer offers.
- **FR-007**: When `sync persona <name>` runs, Sauron shall refresh only the
  named [installed persona](../0014-select-personas/spec.md)'s stored definition,
  applying the same update and timestamp rules scoped to that one persona, and
  shall not set [backend.yaml](../contracts/configuration.md#backendyaml)'s
  `last_synced_at`, which records a full refresh.
- **FR-008**: When `--force` is provided to `sync persona <name>`, Sauron shall
  re-pull that [installed persona](../0014-select-personas/spec.md)'s definition
  authoritatively and, if the [backend](../0012-backend/spec.md) no longer offers
  it, uninstall it.

### State-driven

- **FR-009**: While `sync personas` runs without `--force` and every
  [installed persona](../0014-select-personas/spec.md)'s stored definition
  already matches the [backend](../0012-backend/spec.md), Sauron shall make no
  change, report that the installed personas are already up to date, and exit
  successfully.

### Unwanted behavior

- **FR-010**: If no [backend](../0012-backend/spec.md) is configured, then Sauron
  shall reject the request and report that a backend must be set first.
- **FR-011**: If the [backend](../0012-backend/spec.md) cannot be reached, then
  Sauron shall reject the request and report that it is unreachable.
- **FR-012**: If [personas.yaml](../contracts/configuration.md#personasyaml) or
  [backend.yaml](../contracts/configuration.md#backendyaml) cannot be read or
  parsed, then Sauron shall reject the request and report that it cannot be read.
- **FR-013**: If `sync persona <name>` names a persona that is not an
  [installed persona](../0014-select-personas/spec.md), then Sauron shall report
  that the persona is not installed and exit with an error.
- **FR-014**: If a command fails, then Sauron shall write exactly one
  human-readable message to stderr.
- **FR-015**: If required arguments or flags are missing or invalid, then Sauron
  shall exit with code 2 without executing the command.

## Key Entities

- **Installed Persona**: a persona activated locally via
  [set persona](../0014-select-personas/spec.md) and stored with its full
  definition in [personas.yaml](../contracts/configuration.md#personasyaml). This
  feature refreshes its stored `description`, `tags`, `skills`, `agents`,
  `last_modified_at`, and `last_synced_at` from the backend; under `--force` it
  is uninstalled when the backend no longer offers it, otherwise kept and
  reported (FR-005, FR-006, FR-008).
- **Backend**: the configured [backend](../0012-backend/spec.md) that owns
  persona definitions (singleton per instance); this feature records its
  `last_synced_at` after a full refresh.

## Notes

- **Redefined for the catalog-free model.** Sync Personas no longer maintains a
  persisted catalog mirror. There is no persisted catalog; the *available*
  personas are a [live view](../contracts/configuration.md#live-persona-view)
  computed at command time. This feature now refreshes the stored definitions of
  the [installed personas](../0014-select-personas/spec.md) in
  [personas.yaml](../contracts/configuration.md#personasyaml) instead of pulling
  definitions into a catalog. FR ids are preserved and redefined in place:
  FR-002–FR-004 now operate on installed personas' stored definitions and
  `backend.yaml` rather than on catalog entries and a persona registry; FR-009
  short-circuits when installed personas already match the backend; FR-012/FR-013
  refer to `personas.yaml`/`backend.yaml` and installed-persona membership rather
  than to a catalog.
- `sync personas` (installed-persona definitions) and
  [sync artifacts](../0006-sync-artifacts/spec.md) (artifact delivery) are
  independent operations: one refreshes the stored *definitions* of installed
  personas, the other delivers *artifacts* to the active provider. Neither
  triggers the other.
- `--dry-run` is intentionally **not** offered by either command; the only flag
  is `--force`. A non-`--force` refresh is already safe — it only updates stored
  definition fields and never uninstalls anything.

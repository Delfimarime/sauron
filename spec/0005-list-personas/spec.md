# List Personas

**Type:** feature

**Depends on:** [backend](../0012-backend/spec.md), [sync personas](../0013-sync-personas/spec.md), [set personas](../0014-select-personas/spec.md)

## Overview

A person responsible for a team's agentic-AI setup needs to see the available
personas and which of them are installed on this instance, so that they can
decide what to install, what to uninstall, and how installed personas are
prioritized. The available personas form the
[catalog](../0012-backend/spec.md): the live view assembled at command time from
the [installed personas](../0014-select-personas/spec.md) — always shown, with
their stored definitions — merged with the personas the
[backend](../0012-backend/spec.md) offers, fetched live when the command runs.
Listing assembles this view and marks which entries are installed; it is
read-only and never persists a catalog. When the backend is unreachable the live
fetch is skipped, so only the installed personas are listed and the command
still succeeds.

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to list the available
  [personas](../0012-backend/spec.md) — the
  [installed](../0014-select-personas/spec.md) ones merged with those the
  backend offers — marking which are installed.

### Event-driven

- **FR-002**: When a user lists personas, Sauron shall assemble the
  [catalog](../0012-backend/spec.md) by merging the
  [installed personas](../0014-select-personas/spec.md) with the personas the
  [backend](../0012-backend/spec.md) offers, fetched live, and display, for each
  persona, its name, whether it is installed, its priority, tags, number of
  skills, number of agents, when it was last updated in the backend, and when it
  was last synced locally.
- **FR-003**: When a persona is not [installed](../0014-select-personas/spec.md),
  Sauron shall show `-` for its priority and its last-synced columns, since those
  values exist only for installed personas.
- **FR-004**: When ordering by `priority`, Sauron shall rank
  [installed personas](../0014-select-personas/spec.md) by priority ascending
  (`0` first) and place not-installed personas — which have no priority — last;
  descending reverses the installed personas' order while still placing
  not-installed personas last.

### Unwanted behavior

- **FR-005**: If the [backend](../0012-backend/spec.md) is unreachable, then
  Sauron shall list only the [installed](../0014-select-personas/spec.md)
  personas and exit successfully.
- **FR-006**: If the filters match no personas, then Sauron shall report that
  none match and exit successfully.
- **FR-007**: If no personas are [installed](../0014-select-personas/spec.md) and
  the [backend](../0012-backend/spec.md) offers none or is unreachable, then
  Sauron shall report that there are no personas and exit successfully.
- **FR-008**: If [personas.yaml](../0014-select-personas/spec.md) exists but
  cannot be read or parsed, then Sauron shall reject the request and report that
  it cannot be read.
- **FR-009**: If `--sort` is not one of `name`, `installed`, `priority`,
  `last-updated`, or `last-synced`, then Sauron shall reject the request and
  report the allowed sort attributes.
- **FR-010**: If `--order` is not `asc` or `desc`, then Sauron shall reject the
  request and report the allowed order values.
- **FR-011**: If `--installed` is not `true` or `false`, then Sauron shall
  reject the request and report the allowed values.
- **FR-012**: If `--fields` names a column outside the valid set, then Sauron
  shall exit with code 2 without executing the command and report the allowed
  fields.

### Optional

- **FR-013**: Where `--search` is provided, Sauron shall include only personas
  whose name or description contains the term, matched case-insensitively.
- **FR-014**: Where `--tag` is provided (repeatable), Sauron shall include only
  personas that carry every given tag.
- **FR-015**: Where `--installed` is provided, Sauron shall include only
  [installed personas](../0014-select-personas/spec.md) when `true` and only
  not-installed personas when `false`; when omitted, both are included.
- **FR-016**: Where more than one of `--search`, `--tag`, and `--installed` is
  provided, Sauron shall include only personas that satisfy all of them.
- **FR-017**: Where `--fields` is provided, Sauron shall display the named
  columns in the given order, always keeping `name` present and first.
- **FR-018**: Where `--sort` is provided, Sauron shall order the personas by the
  chosen attribute — `name`, `installed`, `priority`, `last-updated`, or
  `last-synced` — and shall order by `priority` when it is omitted.
- **FR-019**: Where `--order` is provided, Sauron shall order the personas
  ascending or descending accordingly, and shall order ascending when it is
  omitted.

## Key Entities

- **Catalog persona**: an entry in the
  [catalog](../0012-backend/spec.md) — the live view of available personas
  assembled at command time, never persisted — shown by its name, tags,
  skill/agent counts, and the backend's last-updated time. An available persona
  the backend offers but that is not installed contributes a not-installed row.
- **Installed persona**: a persona activated locally by
  [set personas](../0014-select-personas/spec.md) and stored with its
  definition in [personas.yaml](../0014-select-personas/spec.md); it carries a
  priority and a local last-synced time, both shown only when the persona is
  installed, and is always listed, including offline.

## Notes

- **Catalog redesign (no persisted catalog).** This spec previously treated the
  catalog as a persisted, read-only local mirror pulled from the backend by
  [sync personas](../0013-sync-personas/spec.md), and read the catalog plus the
  installed set from a single settings file. Under the redesigned persona model
  there is no persisted catalog: the [catalog](../0012-backend/spec.md) is the
  [live view](../contracts/configuration.md#live-persona-view) computed at
  command time from the installed personas (from
  [personas.yaml](../0014-select-personas/spec.md)) merged with a live fetch from
  the [backend](../0012-backend/spec.md). The FR ids are unchanged in meaning for
  sorting, filtering, and field selection; the behavior changes recorded here
  are intentional:
  - FR-002 now assembles the live view (installed ∪ live backend fetch) rather
    than reading a persisted catalog and installed set from the settings.
  - FR-003 no longer shows `-` for last-updated on not-installed personas:
    last-updated comes from the backend's definition and is available for the
    personas the backend offers live; only priority and last-synced remain
    installed-only.
  - FR-005 was the empty-catalog case; it is redefined as the offline
    unwanted-behavior requirement (backend unreachable ⇒ list only installed
    personas and exit successfully).
  - FR-007 was "settings file does not exist ⇒ treat catalog as empty"; it is
    redefined as the no-personas case (nothing installed and the backend offers
    none or is unreachable ⇒ report no personas, exit successfully). A missing
    [personas.yaml](../0014-select-personas/spec.md) is read as its empty state
    (no installed personas) per the
    [configuration data contract](../contracts/configuration.md).
  - FR-008 now refers to [personas.yaml](../0014-select-personas/spec.md) being
    unreadable rather than the settings file.

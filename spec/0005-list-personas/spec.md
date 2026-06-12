# List Personas

**Type:** feature
**Depends on:** [backend](../0012-backend/spec.md), [sync personas](../0013-sync-personas/spec.md), [select personas](../0014-select-personas/spec.md)

## Overview

A person responsible for a team's agentic-AI setup needs to see the whole
[catalog](../0012-backend/spec.md) of persona definitions and which of
them are installed on this instance, so that they can decide what to install,
what to uninstall, and how installed personas are prioritized. The catalog is
the local read-only mirror pulled from the
[backend](../0012-backend/spec.md) by
[sync personas](../0013-sync-personas/spec.md); which catalog entries are
[installed](../0014-select-personas/spec.md) is recorded separately. Listing
reads both and joins them; it is read-only and works offline against the local
mirror.

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to list the whole
  [catalog](../0012-backend/spec.md) of personas, marking which are
  [installed](../0014-select-personas/spec.md).

### Event-driven

- **FR-002**: When a user lists personas, Sauron shall read the
  [catalog](../0012-backend/spec.md) and the
  [installed set](../0014-select-personas/spec.md) from the settings and
  display, for each catalog persona, its name, whether it is installed, its
  priority, tags, number of skills, number of agents, when it was last updated
  in the backend, and when it was last synced locally.
- **FR-003**: When a persona is not [installed](../0014-select-personas/spec.md),
  Sauron shall show `-` for its priority, its last-updated, and its last-synced
  columns, since those values exist only for installed personas.
- **FR-004**: When ordering by `priority`, Sauron shall rank
  [installed personas](../0014-select-personas/spec.md) by priority ascending
  (`0` first) and place not-installed personas — which have no priority — last;
  descending reverses the installed personas' order while still placing
  not-installed personas last.

### Unwanted behavior

- **FR-005**: If the [catalog](../0012-backend/spec.md) is empty, then
  Sauron shall report that the catalog is empty and exit successfully.
- **FR-006**: If the filters match no personas, then Sauron shall report that
  none match and exit successfully.
- **FR-007**: If the settings file does not exist, then Sauron shall treat the
  catalog as empty.
- **FR-008**: If the settings exist but cannot be read or parsed, then Sauron
  shall reject the request and report that the settings cannot be read.
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
  [catalog](../0012-backend/spec.md) — the local read-only mirror of
  persona definitions — shown by its name, tags, skill/agent counts, and the
  backend's last-updated time. Its schema is owned by the
  [backend data model](../0012-backend/data/configuration.md).
- **Installed persona**: a catalog persona activated locally by
  [select personas](../0014-select-personas/spec.md); it carries a priority
  and a local last-synced time, both shown only when the persona is installed.

# Select Personas

**Type:** feature
**Depends on:** [backend](../0012-backend/spec.md), [sync personas](../0013-sync-personas/spec.md)

## Overview

A person responsible for a team's agentic-AI setup needs to declare which
catalog personas are installed on this instance, so that only the intended
personas participate in artifact sync and in the right precedence order. The
catalog — the local read-only mirror of persona definitions — is owned by the
[backend](../0012-backend/spec.md) and refreshed by
[sync personas](../0013-sync-personas/spec.md); this feature does not change the
catalog, only which of its entries are active locally.

This feature owns two commands. `set persona <name>...` declares the exact set
of installed personas in one shot — the listed names *become* the installed set,
and their position in the argument list fixes their priority (the first listed
is highest precedence) under the unified priority model
([priority model](../AUTHORING.md#priority-model)).
`unset persona [<name>...]` uninstalls named personas, or all of them when no
name is given, leaving the catalog definitions in place. Adjusting a persona's
priority after installation is a separate concern, owned by
[set priority persona](../0007-set-persona-priority/spec.md).

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to declare the installed set of
  catalog personas with `set persona <name>...`.
- **FR-002**: Sauron shall provide the ability to uninstall personas with
  `unset persona [<name>...]`, leaving their catalog definitions available.

### Event-driven

- **FR-003**: When a user runs `set persona`, Sauron shall require at least one
  persona name.
- **FR-004**: When a user runs `set persona` with one or more names, Sauron
  shall make the given names the exact installed set, uninstalling any persona
  that was previously installed but is not listed (full replacement, not a
  partial add).
- **FR-005**: When a user runs `set persona`, Sauron shall assign each installed
  persona a priority from its position in the argument list — the first listed
  persona is `0`, the next `1`, and so on — per the unified priority model
  ([priority model](../AUTHORING.md#priority-model)).
- **FR-006**: When `set persona` succeeds, Sauron shall report the full
  resulting installed set with each persona's priority, and separately report
  the personas that were dropped.
- **FR-007**: When `set persona` re-declares the installed set, Sauron shall
  reset every installed persona's priority from the new argument order,
  discarding any prior [set priority persona](../0007-set-persona-priority/spec.md)
  adjustment (intended).
- **FR-008**: When a user runs `unset persona` with one or more names, Sauron
  shall uninstall each named persona while leaving its catalog definition
  available, and report the personas that were uninstalled.
- **FR-009**: When a user runs `unset persona` with no name, Sauron shall
  uninstall every installed persona while leaving their catalog definitions
  available, and report the personas that were uninstalled.

### State-driven

- **FR-010**: While a persona name is being validated, Sauron shall leave the
  existing configuration unchanged until validation succeeds.
- **FR-011**: While applying `set persona`, Sauron shall validate that every
  given name exists in the catalog before changing anything, and shall apply the
  new installed set only when all names are valid (transactional, all-or-nothing).

### Unwanted behavior

- **FR-012**: If a user runs `set persona` with no name, then Sauron shall exit
  with code 2 without executing the command and report that
  [unset persona](contracts/command-line.md) clears the installed set.
- **FR-013**: If any name given to `set persona` is not present in the catalog,
  then Sauron shall reject the whole command, leave the configuration unchanged,
  and report that [sync personas](../0013-sync-personas/spec.md) should be run
  first to refresh the catalog.
- **FR-014**: If a user runs `unset persona` for a persona that is not
  installed, then Sauron shall exit successfully and report that nothing was
  deleted.
- **FR-015**: If the settings cannot be read or parsed, then Sauron shall reject
  the request and report that the settings cannot be read.
- **FR-016**: If a command fails, then Sauron shall write exactly one
  human-readable message to stderr.

## Key Entities

- **Catalog persona**: an entry in the catalog — the local read-only mirror of
  persona definitions pulled from the
  [backend](../0012-backend/spec.md) and refreshed by
  [sync personas](../0013-sync-personas/spec.md). It is *available* until
  installed.
- **Installed persona**: a catalog persona activated locally by `set persona`.
  It participates in artifact sync and carries a priority assigned positionally
  by `set persona` and adjustable afterward only through
  [set priority persona](../0007-set-persona-priority/spec.md). Its priority
  follows the unified model
  ([priority model](../AUTHORING.md#priority-model))
  — a non-negative integer, unique within its kind, where the first installed
  persona is `0` and a lower value means higher precedence. The installed set is
  persisted as the `installed` block of `settings.yaml` (see
  [configuration](data/configuration.md)).

## Notes

- `set persona` is a full replacement, not a partial add: there is no way to add
  a single persona while keeping the rest untouched — every install command
  states the complete desired set. Re-running it therefore resets positional
  priorities and discards prior
  [set priority persona](../0007-set-persona-priority/spec.md) adjustments; this
  is intended (FR-007).
- Installing or uninstalling a persona never alters the catalog. Catalog content
  is owned by the [backend](../0012-backend/spec.md) and
  changes only through [sync personas](../0013-sync-personas/spec.md).

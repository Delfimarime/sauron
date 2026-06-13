# Set Personas

**Type:** feature

**Depends on:** [backend](../0012-backend/spec.md), [sync personas](../0013-sync-personas/spec.md)

## Overview

A person responsible for a team's agentic-AI setup needs to declare which
personas are installed on this instance, so that only the intended personas
participate in artifact sync and in the right precedence order. Sauron persists
no catalog: the set of *available* personas is the live view computed at command
time from the installed personas plus a live fetch from the
[backend](../0012-backend/spec.md) (see the
[live persona view](../contracts/configuration.md#live-persona-view)). This
feature declares which of those personas are installed locally, and at install
time stores each installed persona's full definition into
[personas.yaml](../contracts/configuration.md#personasyaml) so that
[sync artifacts](../0006-sync-artifacts/spec.md) works offline afterward.

This feature owns two commands. `set persona <name>...` declares the exact set
of installed personas in one shot — the listed names *become* the installed set,
and their position in the argument list fixes their priority (the first listed
is highest precedence) under the unified priority model
([priority model](../AUTHORING.md#priority-model)). It validates every given
name against a live fetch from the
[backend](../0012-backend/spec.md) before any write, and stores each installed
persona's full definition from that fetch.
`unset persona [<name>...]` uninstalls named personas, or all of them when no
name is given, without contacting the backend. Adjusting a persona's priority
after installation is a separate concern, owned by
[set priority persona](../0007-set-persona-priority/spec.md).

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to declare the installed set of
  personas with `set persona <name>...`.
- **FR-002**: Sauron shall provide the ability to uninstall personas with
  `unset persona [<name>...]`.

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
  shall uninstall each named persona without contacting the backend, and report
  the personas that were uninstalled.
- **FR-009**: When a user runs `unset persona` with no name, Sauron shall
  uninstall every installed persona without contacting the backend, and report
  the personas that were uninstalled.
- **FR-017**: When `set persona` succeeds, Sauron shall store each installed
  persona's full definition fetched from the
  [backend](../0012-backend/spec.md) — its description, tags, skills, agents,
  `last_modified_at`, and `last_synced_at` — into
  [personas.yaml](../contracts/configuration.md#personasyaml), so that
  [sync artifacts](../0006-sync-artifacts/spec.md) works offline afterward.

### State-driven

- **FR-010**: While a persona name is being validated, Sauron shall leave the
  existing configuration unchanged until validation succeeds.
- **FR-011**: While applying `set persona`, Sauron shall validate that every
  given name is offered by a live fetch from the
  [backend](../0012-backend/spec.md) before changing anything, and shall apply
  the new installed set only when all names are valid (transactional,
  all-or-nothing).

### Unwanted behavior

- **FR-012**: If a user runs `set persona` with no name, then Sauron shall exit
  with code 2 without executing the command and report that
  [unset persona](contracts/command-line.md) clears the installed set.
- **FR-013**: If any name given to `set persona` is not offered by the live
  fetch from the [backend](../0012-backend/spec.md), then Sauron shall reject the
  whole command, leave the installed personas unchanged, and report which name is
  not available.
- **FR-018**: If the backend is unreachable, then Sauron shall not change the
  installed personas and exit with a runtime error.
- **FR-014**: If a user runs `unset persona` for a persona that is not
  installed, then Sauron shall exit successfully and report that nothing was
  deleted.
- **FR-015**: If the installed personas cannot be read or parsed, then Sauron
  shall reject the request and report that the installed personas cannot be read.
- **FR-016**: If a command fails, then Sauron shall write exactly one
  human-readable message to stderr.

## Key Entities

- **Available persona**: a persona offered by the live fetch from the
  [backend](../0012-backend/spec.md) but not installed. It is part of the
  [live persona view](../contracts/configuration.md#live-persona-view), which
  Sauron never persists.
- **Installed persona**: a persona activated locally by `set persona`, stored
  with its full definition in
  [personas.yaml](../contracts/configuration.md#personasyaml). It participates in
  artifact sync and carries a priority assigned positionally by `set persona` and
  adjustable afterward only through
  [set priority persona](../0007-set-persona-priority/spec.md). Its priority
  follows the unified model
  ([priority model](../AUTHORING.md#priority-model))
  — a non-negative integer, unique within its kind, where the first installed
  persona is `0` and a lower value means higher precedence.

## Notes

- `set persona` is a full replacement, not a partial add: there is no way to add
  a single persona while keeping the rest untouched — every `set persona`
  states the complete desired set. Re-running it therefore resets positional
  priorities and discards prior
  [set priority persona](../0007-set-persona-priority/spec.md) adjustments; this
  is intended (FR-007).
- **Redesign (no persisted catalog):** this feature previously validated names
  against a persisted catalog (a local read-only mirror of persona definitions)
  and stored the installed set in `settings.yaml`. There is now no persisted
  catalog: `set persona` validates names against a live fetch from the
  [backend](../0012-backend/spec.md) and the installed set lives in
  [personas.yaml](../contracts/configuration.md#personasyaml). FR-011 and FR-013
  were redefined from catalog lookup to live-fetch validation, and FR-015 from
  reading settings to reading the installed personas.
- **Redesign (store definition at install):** `set persona` now stores each
  installed persona's full definition fetched from the backend into
  [personas.yaml](../contracts/configuration.md#personasyaml) (FR-017), which is
  why installing requires the backend reachable (FR-018) and why
  [sync artifacts](../0006-sync-artifacts/spec.md) works offline afterward.
  `unset persona` contacts no backend (FR-008, FR-009).

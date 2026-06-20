# Uninstall Artifacts

**Type:** feature

**Depends on:** [install](../0006-install-artifacts/spec.md)

## Overview

A developer removes named artifacts they previously installed. `uninstall` removes
each named skill, agent, or persona from the provider and the track file.
Uninstalling a persona removes the members it brought in, by provenance, while
keeping any member that was also installed directly or brought in by another
persona. A dry run previews the removal.

## Requirements

### Ubiquitous

- FR-001: Sauron shall remove each named installed artifact of the given kind that
  was sourced from the named registry, deleting it from the provider at its
  recorded `path` and from `track.yaml`.

### Event-driven

- FR-002: When a persona is uninstalled, Sauron shall drop the persona from each
  member's provenance `personas` list and remove a member only when it then has
  `direct: false` and no remaining personas.
- FR-003: When the uninstall is applied, Sauron shall print the plan of removed
  artifacts grouped under `skills:`, `agents:`, and `personas:`, prefixed `-`,
  followed by a summary count.

### Optional

- FR-004: Where `--dry-run` is provided, Sauron shall print the plan without
  changing the environment or the track file.

### Unwanted behavior

- FR-005: When a user uninstalls an artifact that is not installed, Sauron shall
  exit successfully and report that nothing was removed.
- FR-006: If required arguments or flags are missing or invalid, then Sauron shall
  exit with code 2 without executing the command.
- FR-007: If an individual artifact removal fails, then Sauron shall report it and
  continue.

## Key Entities

- **Artifact** — the removed skill, agent, or persona; tracked per the
  [state data contract](../contracts/state.md).
- **Provenance** — `direct` plus `personas`, which governs whether a persona's
  member is removed or retained.

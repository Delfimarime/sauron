# Delete Registry

**Type:** feature

**Depends on:** [uninstall](../0007-uninstall-artifacts/spec.md)

## Overview

A developer needs to remove a source they no longer use. `delete registry`
unregisters the named source and **cascade-uninstalls every artifact sourced from
it** — there is no option to keep installed artifacts: removing the source removes
what it delivered. A dry run previews the cascade.

## Requirements

### Ubiquitous

- FR-001: Sauron shall remove the named registry's `Registry` document from
  `registries.yaml`.
- FR-002: Sauron shall uninstall every tracked artifact whose `spec.registry` is
  the named registry — skills, agents, and personas alike — removing them from the
  provider and from `track.yaml`.

### Event-driven

- FR-003: When the cascade is applied, Sauron shall print the plan of removed
  artifacts grouped under `skills:`, `agents:`, and `personas:`, followed by a
  summary count.

### Optional

- FR-004: Where `--dry-run` is provided, Sauron shall print the plan without
  changing the environment or the track file.

### Unwanted behavior

- FR-005: If no registry of that name exists, then Sauron shall exit successfully
  and report that nothing was deleted.
- FR-006: If required arguments or flags are missing or invalid, then Sauron shall
  exit with code 2 without executing the command.
- FR-007: If an individual artifact removal fails, then Sauron shall report it and
  continue the cascade.

## Key Entities

- **Registry** — the removed source; see the
  [state data contract](../contracts/state.md).
- **Provenance** — the `spec.registry` on each tracked artifact, which selects the
  cascade set.

## Notes

Removal is by source, not by provenance reason: a skill that a persona also
brought in is removed too, because its source registry is gone.

The artifact cascade is realized by a single shared cleaning step that both this
feature and [uninstall](../0007-uninstall-artifacts/spec.md) invoke. This feature
introduces that step as a **seam** and ships the registry-document removal (FR-001),
the not-found-as-success outcome (FR-005), the `--dry-run` preview (FR-004), the
grouped report shape (FR-003), and the usage rejection (FR-006). The cascade's
**body** — uninstalling every tracked artifact whose `spec.registry` matches and
removing it from the provider and `track.yaml` (FR-002), the resulting non-empty
plan content (FR-003), and the continue-on-individual-failure behavior (FR-007) —
is realized only once [uninstall](../0007-uninstall-artifacts/spec.md) fills the
shared step. Until then the cascade is a no-op: the plan is empty, so an applied
removal reports `registry "X" removed; 0 artifacts removed`.

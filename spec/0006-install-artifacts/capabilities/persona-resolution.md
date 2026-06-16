# Persona Resolution

**Type:** capability

**Enables:** [install](../spec.md)

**Enables:** [uninstall](../../0007-uninstall-artifacts/spec.md)

**Enables:** [sync](../../0008-sync/spec.md)

**Enables:** [upgrade](../../0009-upgrade/spec.md)

## Overview

A persona references a set of skills and agents within its own registry. Resolution
reads a persona definition and produces its **membership** — the concrete set of
member skills and agents — which install expands, uninstall contracts, and
sync/upgrade re-resolve against the recorded snapshot.

## Requirements

### Ubiquitous

- FR-001: Sauron shall resolve a persona's membership to the skills and agents it
  references within the persona's own registry.
- FR-002: Sauron shall record the resolved membership as the persona's `members`
  snapshot, and add the persona to each member's provenance `personas` list.

### Event-driven

- FR-003: When a persona's members are reconciled, Sauron shall remove a member
  only when its provenance has `direct: false` and no personas remain — so a
  member also installed directly, or brought in by another persona, is retained.

### Unwanted behavior

- FR-004: If a persona references a member the registry does not offer, then
  Sauron shall report the unresolved member and continue with the rest.

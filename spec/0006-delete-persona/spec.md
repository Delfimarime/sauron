# Delete Persona

**Type:** feature
**Depends on:** [import persona](../0005-import-persona/spec.md), [set priority persona](../0010-set-persona-priority/spec.md)

## Overview

A person responsible for a team's agentic-AI setup needs to remove a persona
Sauron no longer serves, so that it stops shaping what is delivered, while the
artifacts already installed remain in place. Personas are registered by
[import persona](../0005-import-persona/spec.md); deletion removes the persona
definition only.

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to delete a registered persona
  by name.

### Event-driven

- **FR-002**: When a user deletes a persona by name, Sauron shall remove the
  matching persona from the settings.
- **FR-003**: When a persona is deleted, Sauron shall leave any artifacts
  already installed in place; deletion removes the persona definition only,
  and the next [sync](../0009-sync/spec.md) reconciles what is in scope (see
  [ADR-0001](architecture/ADR-0001-unregister-keeps-installed-artifacts.md)).
- **FR-004**: When the deletion succeeds, Sauron shall persist the updated
  settings and report success.
- **FR-005**: When a user deletes a persona that does not exist, Sauron shall
  exit successfully and report that nothing was deleted.

### State-driven

- **FR-006**: While deleting, Sauron shall leave the existing configuration
  unchanged until the removal is persisted; the settings are left untouched on
  any failure.

### Unwanted behavior

- **FR-007**: If no name is provided, then Sauron shall reject the request and
  report that a name is required.
- **FR-008**: If the settings cannot be read or parsed, then Sauron shall
  reject the request and report that the settings cannot be read.
- **FR-009**: If the deletion leaves personas behind, then Sauron shall not
  renumber their priorities; in particular, when exactly one persona remains
  it keeps its current priority, and priority changes stay blocked until
  another persona is imported (see
  [set priority persona](../0010-set-persona-priority/spec.md)).

## Key Entities

- **Persona**: a registered named set of artifacts (see
  [import persona](../0005-import-persona/spec.md)), identified by its name.
  Deletion removes only the definition; artifacts already delivered to the
  target are out of its reach.

## Decision Records

- [Unregister keeps installed artifacts](architecture/ADR-0001-unregister-keeps-installed-artifacts.md)
  — deleting a persona removes the definition only; installed artifacts are
  kept.

# Delete Registry

**Type:** feature

**Depends on:** [add registry](../0001-add-registry/spec.md)

## Overview

A person responsible for a team's agentic-AI setup needs to remove a
registry that Sauron is watching, so that it is no longer kept in sync,
while the artifacts already installed from it remain in place. Registries
are registered by [add registry](../0001-add-registry/spec.md); deletion
unregisters the registry only.

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to delete a registered
  registry by name.

### Event-driven

- **FR-002**: When a user deletes a registry by name, Sauron shall remove
  the matching registry from `registries.yaml`.
- **FR-003**: When a registry is deleted, Sauron shall leave any artifacts
  previously installed from it in place; deletion unregisters the registry
  only (see
  [ADR-0001](architecture/ADR-0001-unregister-keeps-installed-artifacts.md)).
- **FR-004**: When the deletion succeeds, Sauron shall persist the updated
  `registries.yaml` and report success.
- **FR-005**: When a user deletes a registry that does not exist, Sauron
  shall exit successfully and report that nothing was deleted.

### State-driven

- **FR-006**: While deleting, Sauron shall leave the existing configuration
  unchanged until the removal is persisted; `registries.yaml` is left untouched
  on any failure.

### Unwanted behavior

- **FR-007**: If no name is provided, then Sauron shall reject the request and
  report that a name is required.
- **FR-008**: If `registries.yaml` cannot be read or parsed, then Sauron shall
  reject the request and report that it cannot be read.

## Key Entities

- **Registry**: a registered source of artifacts (see
  [add registry](../0001-add-registry/spec.md)), identified by its name.
  Deletion removes only the registration entry; artifacts already delivered to
  the provider are out of its reach.

## Decision Records

- [Unregister keeps installed artifacts](architecture/ADR-0001-unregister-keeps-installed-artifacts.md)
  — deleting a registry unregisters the registry only; installed artifacts
  are kept.

## Notes

- Configuration is now split across files per the
  [configuration data contract](../contracts/configuration.md); file
  references updated accordingly.

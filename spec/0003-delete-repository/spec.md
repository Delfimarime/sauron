# Delete Repository

**Type:** feature
**Depends on:** [add repository](../0001-add-repository/spec.md)

## Overview

A person responsible for a team's agentic-AI setup needs to remove a
repository that Sauron is watching, so that it is no longer kept in sync,
while the artifacts already installed from it remain in place. Repositories
are registered by [add repository](../0001-add-repository/spec.md); deletion
unregisters the repository only.

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to delete a registered
  repository by name.

### Event-driven

- **FR-002**: When a user deletes a repository by name, Sauron shall remove
  the matching repository from the settings.
- **FR-003**: When a repository is deleted, Sauron shall leave any artifacts
  previously installed from it in place; deletion unregisters the repository
  only (see
  [ADR-0001](architecture/ADR-0001-unregister-keeps-installed-artifacts.md)).
- **FR-004**: When the deletion succeeds, Sauron shall persist the updated
  settings and report success.
- **FR-005**: When a user deletes a repository that does not exist, Sauron
  shall exit successfully and report that nothing was deleted.

### State-driven

- **FR-006**: While deleting, Sauron shall leave the existing configuration
  unchanged until the removal is persisted; the settings are left untouched on
  any failure.

### Unwanted behavior

- **FR-007**: If no name is provided, then Sauron shall reject the request and
  report that a name is required.
- **FR-008**: If the settings cannot be read or parsed, then Sauron shall
  reject the request and report that the settings cannot be read.

## Key Entities

- **Repository**: a registered source of artifacts (see
  [add repository](../0001-add-repository/spec.md)), identified by its name.
  Deletion removes only the registration entry; artifacts already delivered to
  the target are out of its reach.

## Decision Records

- [Unregister keeps installed artifacts](architecture/ADR-0001-unregister-keeps-installed-artifacts.md)
  — deleting a repository unregisters the repository only; installed artifacts
  are kept.

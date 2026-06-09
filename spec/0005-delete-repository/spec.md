# Feature Specification: Delete Repository

**Created**: 2026-06-09

**Status**: Draft

**Input**: "Allow a user to delete a repository registered with Sauron without removing the skills and agents already installed from it."

## Overview

A person responsible for a team's agentic-AI setup needs to remove a repository that Sauron is watching, so that it is no longer kept in sync, while the skills and agents already installed from it remain in place.

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to delete a registered repository by name.

### Event-driven (*When*)

- **FR-002**: When the user deletes a repository by name, Sauron shall remove the matching repository from its configuration.
- **FR-003**: When a repository is deleted, Sauron shall leave any skills and agents previously installed from it in place; deletion unregisters the source only. (See ADR-0001.)
- **FR-004**: When the deletion succeeds, Sauron shall persist the updated configuration and report success.

### State-driven (*While*)

- **FR-005**: While deleting, Sauron shall leave the existing configuration unchanged until the removal is persisted; the file is left untouched on any failure.

### Unwanted-behavior (*If / then*)

- **FR-006**: If no name is provided, then Sauron shall reject the request and report that a name is required.
- **FR-007**: If no repository with the given name exists, then Sauron shall make no change and report that there was nothing to delete (treated as success).
- **FR-008**: If the configuration cannot be read or parsed, then Sauron shall reject the request and report that the configuration cannot be read.

## Key Entities

- **Repository**: a registered source identified by its name. Deletion removes only the registration entry; artifacts already delivered to targets are out of its reach.

## Decision Records

- `architecture/ADR-0001-unregister-keeps-installed-artifacts.md` — deleting a repository unregisters the source only; installed skills and agents are kept.

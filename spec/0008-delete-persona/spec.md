# Feature Specification: Delete Persona

**Created**: 2026-06-10

**Status**: Draft

**Input**: "Allow a user to delete a persona registered with Sauron without removing the skills and agents already installed."

## Overview

A person responsible for a team's agentic-AI setup needs to remove a persona Sauron no longer serves, so that it stops shaping what is delivered, while the skills and agents already installed remain in place.

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to delete a registered persona by name.

### Event-driven (*When*)

- **FR-002**: When the user deletes a persona by name, Sauron shall remove the matching persona from its configuration.
- **FR-003**: When a persona is deleted, Sauron shall leave any skills and agents already installed in place; deletion removes the persona definition only, and the next sync reconciles what is in scope. (See ADR-0001.)
- **FR-004**: When the deletion succeeds, Sauron shall persist the updated configuration and report success.

### State-driven (*While*)

- **FR-005**: While deleting, Sauron shall leave the existing configuration unchanged until the removal is persisted; the file is left untouched on any failure.

### Unwanted-behavior (*If / then*)

- **FR-006**: If no name is provided, then Sauron shall reject the request and report that a name is required.
- **FR-007**: If no persona with the given name exists, then Sauron shall make no change and report that there was nothing to delete (treated as success).
- **FR-008**: If the configuration cannot be read or parsed, then Sauron shall reject the request and report that the configuration cannot be read.
- **FR-009**: If the deletion leaves personas behind, then Sauron shall not renumber their priorities; in particular, when exactly one persona remains it keeps its current priority, and priority changes stay blocked until another persona is imported (see `0012-set-persona-priority`).

## Key Entities

- **Persona**: a registered named set of agents and skills, identified by its name. Deletion removes only the definition; artifacts already delivered to targets are out of its reach.

## Decision Records

- `architecture/ADR-0001-unregister-keeps-installed-artifacts.md` — deleting a persona removes the definition only; installed skills and agents are kept.

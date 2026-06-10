# Feature Specification: Update Persona

**Created**: 2026-06-10

**Status**: Draft

**Input**: "Allow a user to update an existing persona from the same YAML definition file format used by import; the name in the file is the key."

## Overview

A person responsible for a team's agentic-AI setup needs to revise a persona's definition, so that the group it describes receives an updated set of skills and agents.

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to update an existing persona from a YAML definition file.

### Event-driven (*When*)

- **FR-002**: When a user updates a persona, Sauron shall require a path to a persona definition file in the same format as import (`0007-import-persona`).
- **FR-003**: When a persona definition file is submitted, Sauron shall validate the same shape as import: a `name` (slug), a `description`, optional `tags`, and at least one entry across `agents` and `skills`.
- **FR-004**: When the definition is valid, Sauron shall locate the persona by the `name` in the file — the name is the key.
- **FR-005**: When the persona is found, Sauron shall replace its `description`, `tags`, `agents`, and `skills` with the values from the file.
- **FR-006**: When updating, Sauron shall preserve the persona's priority unchanged, whether defined or undefined; the definition file never carries a priority, and priority is managed by `set priority persona` (`0012-set-persona-priority`; see `0007-import-persona` ADR-0001).
- **FR-007**: When the update succeeds, Sauron shall persist the updated configuration and report success.

### State-driven (*While*)

- **FR-008**: While a persona is being validated and updated, Sauron shall leave the existing configuration unchanged until the update is persisted; the file is left untouched on any failure.

### Unwanted-behavior (*If / then*)

- **FR-009**: If no path is provided, then Sauron shall reject the request and report that a path is required.
- **FR-010**: If the definition file does not exist, cannot be read, or is not valid YAML, then Sauron shall reject the request and report that the persona definition cannot be read.
- **FR-011**: If the name is missing or is not a valid slug, then Sauron shall reject the request and report that the name format is invalid.
- **FR-012**: If the description is missing, then Sauron shall reject the request and report that a description is required.
- **FR-013**: If the definition contains neither at least one agent nor at least one skill, then Sauron shall reject the request and report that a persona needs at least one skill or agent.
- **FR-014**: If no persona with the file's name exists, then Sauron shall reject the request and report that the persona is not found and that import creates new personas.
- **FR-015**: If the configuration cannot be read or parsed, then Sauron shall reject the request and report that the configuration cannot be read.

## Key Entities

- **Persona**: a registered named set of agents and skills, identified by its name. Update replaces its description, tags, agents, and skills; its name and priority are untouched.

## Notes

- Like import, update validates the definition's shape only. Whether the listed agents and skills exist in the registered repositories is resolved by the sync feature (`0011-sync`).

# Update Persona

**Type:** feature
**Depends on:** [import persona](../0005-import-persona/spec.md), [set priority persona](../0010-set-persona-priority/spec.md)

## Overview

A person responsible for a team's agentic-AI setup needs to revise a persona's
definition, so that the group it describes receives an updated set of
artifacts. The definition file format is the same one used by
[import persona](../0005-import-persona/spec.md); the `name` in the file is
the key.

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to update an existing persona
  from a YAML definition file.

### Event-driven

- **FR-002**: When a user updates a persona, Sauron shall require a path to a
  persona definition file in the same format as
  [import persona](../0005-import-persona/spec.md).
- **FR-003**: When a persona definition file is submitted, Sauron shall
  validate the same shape as import: a `name` (slug), a `description`,
  optional `tags`, and at least one entry across `agents` and `skills`.
- **FR-004**: When the definition is valid, Sauron shall locate the persona by
  the `name` in the file — the name is the key.
- **FR-005**: When the persona is found, Sauron shall replace its
  `description`, `tags`, `agents`, and `skills` with the values from the file.
- **FR-006**: When updating, Sauron shall preserve the persona's priority
  unchanged; the definition file never carries a priority, and priority is
  managed by
  [set priority persona](../0010-set-persona-priority/spec.md) (see
  [import persona ADR-0002](../0005-import-persona/architecture/ADR-0002-unified-priority-model.md)).
- **FR-007**: When the update succeeds, Sauron shall persist the updated
  settings and report success.

### State-driven

- **FR-008**: While a persona is being validated and updated, Sauron shall
  leave the existing configuration unchanged until the update is persisted;
  the settings are left untouched on any failure.

### Unwanted behavior

- **FR-009**: If no path is provided, then Sauron shall reject the request and
  report that a path is required.
- **FR-010**: If the definition file does not exist, cannot be read, or is not
  valid YAML, then Sauron shall reject the request and report that the persona
  definition cannot be read.
- **FR-011**: If the name is missing or is not a valid slug, then Sauron shall
  reject the request and report that the name format is invalid.
- **FR-012**: If the description is missing, then Sauron shall reject the
  request and report that a description is required.
- **FR-013**: If the definition contains neither at least one agent nor at
  least one skill, then Sauron shall reject the request and report that a
  persona needs at least one skill or agent.
- **FR-014**: If no persona with the file's name exists, then Sauron shall
  reject the request and report that the persona is not found and that
  [import persona](../0005-import-persona/spec.md) creates new personas.
- **FR-015**: If the settings cannot be read or parsed, then Sauron shall
  reject the request and report that the settings cannot be read.

## Key Entities

- **Persona**: a registered named set of artifacts (see
  [import persona](../0005-import-persona/spec.md)), identified by its name.
  Update replaces its description, tags, agents, and skills; its name and
  priority are untouched.

## Notes

- Like import, update validates the definition's shape only. Whether the
  listed agents and skills exist in the registered repositories is resolved by
  [sync](../0009-sync/spec.md).

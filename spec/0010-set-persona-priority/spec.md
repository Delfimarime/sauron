# Set Persona Priority

**Type:** feature
**Depends on:** [import persona](../0005-import-persona/spec.md)

## Overview

A person responsible for a team's agentic-AI setup needs to change a persona's
priority, so that the precedence order among personas reflects the team's
intent. Persona priority follows the model decided in
[import persona ADR-0001](../0005-import-persona/architecture/ADR-0001-persona-priority-model.md).

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to set a persona's priority by
  name.

### Event-driven

- **FR-002**: When a user sets a priority, Sauron shall require a persona name
  and a priority value.
- **FR-003**: When the request is valid, Sauron shall assign the value as the
  persona's defined priority and persist the updated settings.
- **FR-004**: When the value equals the persona's current priority, Sauron
  shall make no change and report success (no-op).
- **FR-005**: When the persona's priority is undefined, Sauron shall accept
  the request and give it the defined value (per
  [import persona ADR-0001](../0005-import-persona/architecture/ADR-0001-persona-priority-model.md),
  this moves it from the undefined tier into the defined order).

### State-driven

- **FR-006**: While setting a priority, Sauron shall leave the existing
  configuration unchanged until the change is persisted; the settings are left
  untouched on any failure.

### Unwanted behavior

- **FR-007**: If the name or the value is missing, then Sauron shall reject
  the request and report what is required.
- **FR-008**: If the value is not a non-negative integer, then Sauron shall
  reject the request and report that a non-negative integer is required.
- **FR-009**: If only one persona exists, then Sauron shall reject the request
  and report that priority cannot be changed while a single persona exists —
  it keeps priority `0` (see
  [import persona ADR-0001](../0005-import-persona/architecture/ADR-0001-persona-priority-model.md)).
- **FR-010**: If no persona with the given name exists, then Sauron shall
  reject the request and report that the persona is not found.
- **FR-011**: If the value is already used by another persona, then Sauron
  shall reject the request, leave the configuration unchanged, and report that
  the priority must be unique (`0` is assignable only when no persona holds
  it).
- **FR-012**: If the settings cannot be read or parsed, then Sauron shall
  reject the request and report that the settings cannot be read.

## Key Entities

- **Persona**: a registered named set of artifacts (see
  [import persona](../0005-import-persona/spec.md)), identified by its name.
  Its priority follows the model in
  [import persona ADR-0001](../0005-import-persona/architecture/ADR-0001-persona-priority-model.md)
  — zero-anchored, optional, defined values unique, undefined ranks last; this
  feature is the only way to change it after import.

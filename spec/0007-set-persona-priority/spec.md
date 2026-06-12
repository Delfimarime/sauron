# Set Persona Priority

**Type:** feature
**Depends on:** [select personas](../0014-select-personas/spec.md)

## Overview

A person responsible for a team's agentic-AI setup needs to change an installed
persona's priority, so that the precedence order among installed personas
reflects the team's intent. Persona priority is assigned subscribe-time:
positionally by [select personas](../0014-select-personas/spec.md) when
`set persona` declares the installed set (the catalog persona's position fixes
its priority). This command adjusts the priority of one already-installed
persona afterward; the override persists until the next `set persona`
redeclaration, which resets positional priorities. Persona priority follows the
unified model decided in
[priority model](../AUTHORING.md#priority-model):
an installed persona's priority is always defined and unique within its kind, a
lower value wins, and the first or only installed persona holds `0`.

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to set an installed persona's
  priority by name.

### Event-driven

- **FR-002**: When a user sets a priority, Sauron shall require an installed
  persona name and a priority value.
- **FR-003**: When the request is valid, Sauron shall assign the value as the
  installed persona's defined priority and persist the updated settings.
- **FR-004**: When the value equals the installed persona's current priority,
  Sauron shall make no change and report success (no-op).
- **FR-005**: When the request changes the installed persona's priority, Sauron
  shall replace its current value with the new value and re-establish the unique
  precedence order (per
  [priority model](../AUTHORING.md#priority-model)).

### State-driven

- **FR-006**: While setting a priority, Sauron shall leave the existing
  configuration unchanged until the change is persisted; the settings are left
  untouched on any failure.

### Unwanted behavior

- **FR-007**: If the name or the value is missing, then Sauron shall reject
  the request and report what is required.
- **FR-008**: If the value is not a non-negative integer, then Sauron shall
  reject the request and report that a non-negative integer is required.
- **FR-009**: If only one persona is installed, then Sauron shall reject the
  request and report that priority cannot be changed while a single persona is
  installed — it keeps priority `0` (see
  [priority model](../AUTHORING.md#priority-model)).
- **FR-010**: If no installed persona with the given name exists, then Sauron
  shall reject the request and report that the persona is not found.
- **FR-011**: If the value is already used by another installed persona, then
  Sauron shall reject the request, leave the configuration unchanged, and report
  that the priority must be unique (`0` is assignable only when no installed
  persona holds it).
- **FR-012**: If the settings cannot be read or parsed, then Sauron shall
  reject the request and report that the settings cannot be read.

## Key Entities

- **Installed persona**: a catalog persona activated locally by `set persona`
  (see [select personas](../0014-select-personas/spec.md)), identified by its
  name and participating in artifact sync. Its priority is assigned positionally
  at install time and follows the unified model in
  [priority model](../AUTHORING.md#priority-model)
  — always defined and unique within its kind, a lower value wins, and the first
  or only installed persona holds `0`. This feature is the only way to adjust an
  installed persona's priority after installation; the override persists until
  the next [select personas](../0014-select-personas/spec.md) `set persona`
  redeclaration resets positional priorities.

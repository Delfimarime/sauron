# Import Persona

**Type:** feature

## Overview

A person responsible for a team's agentic-AI setup needs to import a persona —
a named set of artifacts shared by a group of people — so that Sauron can
deliver the same artifacts to everyone in that group.

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to import a persona from a YAML
  definition file.
- **FR-002**: Sauron shall require every persona to have a name that is unique
  across all personas.
- **FR-003**: Sauron shall require every persona to have a priority that is
  unique across all personas (see
  [ADR-0002](architecture/ADR-0002-unified-priority-model.md)).

### Event-driven

- **FR-004**: When a user imports a persona, Sauron shall require a path to a
  persona definition file.
- **FR-005**: When a persona definition file is submitted, Sauron shall read
  it as YAML and validate that it carries a `name` (slug
  `^[a-z0-9]+(-[a-z0-9]+)*$`), a `description`, optional `tags`, and
  `agents`/`skills` lists containing at least one entry across the two.
- **FR-006**: When the first persona is imported (no personas registered yet),
  Sauron shall assign it priority `0` (see
  [ADR-0002](architecture/ADR-0002-unified-priority-model.md)).
- **FR-007**: When `--priority` is omitted and personas already exist, Sauron
  shall assign the next priority one greater than the highest existing persona
  priority (`max + 1`) (see
  [ADR-0002](architecture/ADR-0002-unified-priority-model.md)).
- **FR-008**: When a persona passes validation, Sauron shall persist it to the
  settings (`~/.sauron/settings.yaml`).
- **FR-009**: When a persona is successfully imported, Sauron shall report
  success.

### State-driven

- **FR-010**: While a persona is being validated, Sauron shall leave the
  existing configuration unchanged until validation succeeds.

### Unwanted behavior

- **FR-011**: If no path is provided, then Sauron shall reject the request and
  report that a path is required.
- **FR-012**: If the definition file does not exist, cannot be read, or is not
  valid YAML, then Sauron shall reject the request and report that the persona
  definition cannot be read.
- **FR-013**: If the name is missing or is not a valid slug, then Sauron shall
  reject the request and report that the name format is invalid.
- **FR-014**: If the description is missing, then Sauron shall reject the
  request and report that a description is required.
- **FR-015**: If the definition contains neither at least one agent nor at
  least one skill, then Sauron shall reject the request and report that a
  persona needs at least one skill or agent.
- **FR-016**: If the name is already used by another persona, then Sauron
  shall reject the request, leave the configuration unchanged, and report that
  the name must be unique.
- **FR-017**: If `--priority` is provided when no personas exist, then Sauron
  shall reject the request and report that the first persona always takes
  priority `0`.
- **FR-018**: If a provided `--priority` is not a non-negative integer, then
  Sauron shall reject the request and report that a valid priority is required.
- **FR-019**: If `--priority` is already used by another persona, then Sauron
  shall reject the request, leave the configuration unchanged, and report that
  the priority must be unique.

### Optional

- **FR-020**: Where `--priority` is provided and personas already exist,
  Sauron shall require a non-negative integer that is not used by another
  persona.

## Key Entities

- **Persona**: a named set of artifacts shared by a group of people,
  identified by its **name**. Carries:
  - **name** — a unique slug (`^[a-z0-9]+(-[a-z0-9]+)*$`).
  - **description** — a human-readable account of who the persona is for.
  - **tags** (optional) — labels used for filtering.
  - **agents** / **skills** — names of the artifacts the persona delivers; at
    least one entry across the two lists.
  - **priority** — an optional non-negative integer, always defined and unique
    across personas; lower value wins. `0` for the first persona; an omitted
    `--priority` on a later import appends `max + 1` (one greater than the
    highest existing persona priority) (see
    [ADR-0002](architecture/ADR-0002-unified-priority-model.md)).

## Decision Records

- [Unified priority model](architecture/ADR-0002-unified-priority-model.md) —
  repositories and personas share one priority model: `--priority` is optional,
  the first resource is `0`, an omitted value appends `max + 1`, and every
  priority is defined and unique within its kind.
- [Persona priority model](architecture/ADR-0001-persona-priority-model.md) —
  superseded by [ADR-0002](architecture/ADR-0002-unified-priority-model.md).

## Notes

- Import validates the definition's shape only. Whether the listed agents and
  skills exist in the registered repositories is resolved by
  [sync](../0009-sync/spec.md).

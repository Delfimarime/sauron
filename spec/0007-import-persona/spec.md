# Feature Specification: Import Persona

**Created**: 2026-06-10

**Status**: Draft

**Input**: "Allow a user to import a persona — a named set of agents and skills shared by a group of people — from a YAML definition file."

## Overview

A person responsible for a team's agentic-AI setup needs to import a persona — a named set of agents and skills shared by a group of people — so that Sauron can deliver the same artifacts to everyone in that group.

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to import a persona from a YAML definition file.
- **FR-002**: Every persona shall have a name that is unique across all personas.
- **FR-003**: Every persona with a defined priority shall have a priority that is unique across all personas. (See ADR-0001.)

### Event-driven (*When*)

- **FR-004**: When a user imports a persona, Sauron shall require a path to a persona definition file.
- **FR-005**: When a persona definition file is submitted, Sauron shall read it as YAML and validate that it carries a `name` (slug `^[a-z0-9]+(-[a-z0-9]+)*$`), a `description`, optional `tags`, and `agents`/`skills` lists containing at least one entry across the two.
- **FR-006**: When the first persona is imported (no personas registered yet), Sauron shall assign it priority `0`. (See ADR-0001.)
- **FR-007**: When `--priority` is provided and personas already exist, Sauron shall require an integer of `1` or greater that is not used by another persona.
- **FR-008**: When `--priority` is omitted and personas already exist, Sauron shall register the persona without a priority (undefined). (See ADR-0001.)
- **FR-009**: When a persona passes validation, Sauron shall persist it to its configuration (`~/.sauron/settings.yaml`).
- **FR-010**: When a persona is successfully imported, Sauron shall report success.

### State-driven (*While*)

- **FR-011**: While a persona is being validated, Sauron shall leave the existing configuration unchanged until validation succeeds.

### Unwanted-behavior (*If / then*)

- **FR-012**: If no path is provided, then Sauron shall reject the request and report that a path is required.
- **FR-013**: If the definition file does not exist, cannot be read, or is not valid YAML, then Sauron shall reject the request and report that the persona definition cannot be read.
- **FR-014**: If the name is missing or is not a valid slug, then Sauron shall reject the request and report that the name format is invalid.
- **FR-015**: If the description is missing, then Sauron shall reject the request and report that a description is required.
- **FR-016**: If the definition contains neither at least one agent nor at least one skill, then Sauron shall reject the request and report that a persona needs at least one skill or agent.
- **FR-017**: If the name is already used by another persona, then Sauron shall reject the request, leave the configuration unchanged, and report that the name must be unique.
- **FR-018**: If `--priority` is provided when no personas exist, then Sauron shall reject the request and report that the first persona always takes priority `0`.
- **FR-019**: If `--priority` is not an integer of `1` or greater, then Sauron shall reject the request and report that a valid priority is required (`0` belongs to the first persona).
- **FR-020**: If `--priority` is already used by another persona, then Sauron shall reject the request, leave the configuration unchanged, and report that the priority must be unique.

## Key Entities

- **Persona**: a named set of agents and skills shared by a group of people, identified by its **name**. Carries:
  - **name** — a unique slug (`^[a-z0-9]+(-[a-z0-9]+)*$`).
  - **description** — a human-readable account of who the persona is for.
  - **tags** (optional) — labels used for filtering.
  - **agents** / **skills** — names of the artifacts the persona delivers; at least one entry across the two lists.
  - **priority** — `0` for the first persona (forced), an integer ≥ 1 unique among personas when set, or undefined when imported without `--priority`. Undefined ranks after all defined priorities. (See ADR-0001.)

## Notes

- Import validates the definition's shape only. Whether the listed agents and skills exist in the registered repositories is resolved by the sync feature (`0011-sync`).

## Decision Records

- `architecture/ADR-0001-persona-priority-model.md` — persona priority is zero-anchored, optional, and undefined ranks last.

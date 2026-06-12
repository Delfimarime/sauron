# List Personas

**Type:** feature
**Depends on:** [import persona](../0005-import-persona/spec.md)

## Overview

A person responsible for a team's agentic-AI setup needs to see which personas
are defined, so that they can review how artifacts are grouped for delivery.
Personas are registered by [import persona](../0005-import-persona/spec.md);
listing is read-only.

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to list the registered
  personas.

### Event-driven

- **FR-002**: When a user lists personas, Sauron shall read the registered
  personas from the settings and display each one's name, priority, tags, and
  number of skills and agents.
- **FR-003**: When ordering by priority, Sauron shall rank personas by priority
  ascending (`0` first), since every persona has a defined, unique priority (per
  [import persona ADR-0002](../0005-import-persona/architecture/ADR-0002-unified-priority-model.md));
  descending reverses that order.

### Unwanted behavior

- **FR-004**: If no personas are registered, then Sauron shall report that
  none are registered and exit successfully.
- **FR-005**: If the filters match no personas, then Sauron shall report that
  none match and exit successfully.
- **FR-006**: If the settings file does not exist, then Sauron shall treat it
  as no personas registered.
- **FR-007**: If the settings exist but cannot be read or parsed, then Sauron
  shall reject the request and report that the settings cannot be read.
- **FR-008**: If `--sort` is not one of `name` or `priority`, then Sauron
  shall reject the request and report the allowed sort attributes.
- **FR-009**: If `--order` is not `asc` or `desc`, then Sauron shall reject
  the request and report the allowed order values.

### Optional

- **FR-010**: Where `--search` is provided, Sauron shall include only personas
  whose name or description contains the term, matched case-insensitively.
- **FR-011**: Where `--tag` is provided (repeatable), Sauron shall include
  only personas that carry every given tag.
- **FR-012**: Where both `--search` and `--tag` are provided, Sauron shall
  include only personas that satisfy both.
- **FR-013**: Where `--sort` is provided, Sauron shall order the personas by
  the chosen attribute — `name` or `priority` — and shall order by `priority`
  when it is omitted.
- **FR-014**: Where `--order` is provided, Sauron shall order the personas
  ascending or descending accordingly, and shall order ascending when it is
  omitted.

## Key Entities

- **Persona**: a registered named set of artifacts (see
  [import persona](../0005-import-persona/spec.md)), shown by its name,
  priority, tags, and skill/agent counts.

# List Registries

**Type:** feature
**Depends on:** [add registry](../0001-add-registry/spec.md)

## Overview

A person responsible for a team's agentic-AI setup needs to see which
registries Sauron is watching, so that they can review the team's configured
sources of artifacts. Registries are registered by
[add registry](../0001-add-registry/spec.md); listing is read-only and
spans all kinds.

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to list the registered
  registries.

### Event-driven

- **FR-002**: When a user lists registries, Sauron shall read the registered
  registries from the settings and display each one's name, kind, priority,
  and location.
- **FR-003**: When displaying a registry's location, Sauron shall use the
  kind-appropriate locator field — `path` for `filesystem`, `url` for `http`,
  `uri` for `git`.

### Unwanted behavior

- **FR-004**: If no registries are registered, then Sauron shall report that
  none are registered and exit successfully.
- **FR-005**: If `--search` matches no registries, then Sauron shall report
  that none match and exit successfully.
- **FR-006**: If the settings file does not exist, then Sauron shall treat it
  as no registries registered.
- **FR-007**: If the settings exist but cannot be read or parsed, then Sauron
  shall reject the request and report that the settings cannot be read.
- **FR-008**: If `--sort` is not one of `name`, `priority`, or `kind`, then
  Sauron shall reject the request and report the allowed sort attributes.
- **FR-009**: If `--order` is not `asc` or `desc`, then Sauron shall reject the
  request and report the allowed order values.
- **FR-014**: If `--fields` names a column other than `name`, `kind`,
  `priority`, or `location`, then Sauron shall reject the request and report the
  allowed fields.

### Optional

- **FR-010**: Where `--search` is provided, Sauron shall include only
  registries whose name or location contains the term, matched
  case-insensitively.
- **FR-011**: Where `--sort` is provided, Sauron shall order the registries
  by the chosen attribute — `name`, `priority`, or `kind` — and shall order by
  `priority` when it is omitted.
- **FR-012**: Where `--order` is provided, Sauron shall order the registries
  ascending or descending accordingly, and shall order ascending when it is
  omitted.
- **FR-013**: Where `--fields` is provided, Sauron shall display only the named
  columns in the given order, with `name` always present and first.

## Key Entities

- **Registry**: a registered source of artifacts (see
  [add registry](../0001-add-registry/spec.md)), shown by its name, kind,
  priority, and location. The location is the kind's locator
  (`path`/`url`/`uri`). Listing spans all kinds.

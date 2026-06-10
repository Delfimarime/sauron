# List Repositories

**Type:** feature
**Depends on:** [add repository](../0001-add-repository/spec.md)

## Overview

A person responsible for a team's agentic-AI setup needs to see which
repositories Sauron is watching, so that they can review the team's configured
sources of artifacts. Repositories are registered by
[add repository](../0001-add-repository/spec.md); listing is read-only and
spans all kinds.

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to list the registered
  repositories.

### Event-driven

- **FR-002**: When a user lists repositories, Sauron shall read the registered
  repositories from the settings and display each one's name, kind, priority,
  and location.
- **FR-003**: When displaying a repository's location, Sauron shall use the
  kind-appropriate locator field — `path` for `filesystem`, `url` for `http`,
  `uri` for `git`.

### Unwanted behavior

- **FR-004**: If no repositories are registered, then Sauron shall report that
  none are registered and exit successfully.
- **FR-005**: If `--search` matches no repositories, then Sauron shall report
  that none match and exit successfully.
- **FR-006**: If the settings file does not exist, then Sauron shall treat it
  as no repositories registered.
- **FR-007**: If the settings exist but cannot be read or parsed, then Sauron
  shall reject the request and report that the settings cannot be read.
- **FR-008**: If `--sort` is not one of `name`, `priority`, or `kind`, then
  Sauron shall reject the request and report the allowed sort attributes.
- **FR-009**: If `--order` is not `asc` or `desc`, then Sauron shall reject the
  request and report the allowed order values.

### Optional

- **FR-010**: Where `--search` is provided, Sauron shall include only
  repositories whose name or location contains the term, matched
  case-insensitively.
- **FR-011**: Where `--sort` is provided, Sauron shall order the repositories
  by the chosen attribute — `name`, `priority`, or `kind` — and shall order by
  `priority` when it is omitted.
- **FR-012**: Where `--order` is provided, Sauron shall order the repositories
  ascending or descending accordingly, and shall order ascending when it is
  omitted.

## Key Entities

- **Repository**: a registered source of artifacts (see
  [add repository](../0001-add-repository/spec.md)), shown by its name, kind,
  priority, and location. The location is the kind's locator
  (`path`/`url`/`uri`). Listing spans all kinds.

# Set Repository Priority

**Type:** feature
**Depends on:** [add repository](../0001-add-repository/spec.md)

## Overview

A person responsible for a team's agentic-AI setup needs to change a
repository's priority, so that the precedence order among repositories
reflects the team's intent. Repositories and their priorities are registered
by [add repository](../0001-add-repository/spec.md); priority drives conflict
resolution during [sync](../0009-sync/spec.md).

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to set a repository's priority
  by name.

### Event-driven

- **FR-002**: When a user sets a priority, Sauron shall require a repository
  name and a priority value.
- **FR-003**: When the request is valid, Sauron shall assign the value as the
  repository's priority and persist the updated settings.
- **FR-004**: When the value equals the repository's current priority, Sauron
  shall make no change and report success (no-op).

### State-driven

- **FR-005**: While setting a priority, Sauron shall leave the existing
  configuration unchanged until the change is persisted; the settings are left
  untouched on any failure.

### Unwanted behavior

- **FR-006**: If the name or the value is missing, then Sauron shall reject
  the request and report what is required.
- **FR-007**: If the value is not a non-negative integer, then Sauron shall
  reject the request and report that a non-negative integer is required.
- **FR-008**: If no repository with the given name exists, then Sauron shall
  reject the request and report that the repository is not found.
- **FR-009**: If the value is already used by another repository, then Sauron
  shall reject the request, leave the configuration unchanged, and report that
  the priority must be unique (`0` is assignable only when no repository holds
  it).
- **FR-010**: If the settings cannot be read or parsed, then Sauron shall
  reject the request and report that the settings cannot be read.
- **FR-011**: If only one repository exists, then Sauron shall reject the
  request and report that priority cannot be changed while a single repository
  exists — it keeps priority `0` (see
  [unified priority model](../0005-import-persona/architecture/ADR-0002-unified-priority-model.md)).

## Key Entities

- **Repository**: a registered source of artifacts (see
  [add repository](../0001-add-repository/spec.md)), identified by its name.
  Its priority follows the unified model in
  [unified priority model](../0005-import-persona/architecture/ADR-0002-unified-priority-model.md)
  — an optional non-negative integer, unique across all repositories
  regardless of kind, where the first repository is `0` and a lower value means
  higher precedence; this feature is the only way to change it after the
  repository is added.

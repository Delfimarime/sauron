# Set Registry Priority

**Type:** feature

**Depends on:** [add registry](../0001-add-registry/spec.md)

## Overview

A person responsible for a team's agentic-AI setup needs to change a
registry's priority, so that the precedence order among registries
reflects the team's intent. Registries and their priorities are registered
by [add registry](../0001-add-registry/spec.md); priority drives conflict
resolution during [sync](../0006-sync-artifacts/spec.md).

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to set a registry's priority
  by name.

### Event-driven

- **FR-002**: When a user sets a priority, Sauron shall require a registry
  name and a priority value.
- **FR-003**: When the request is valid, Sauron shall assign the value as the
  registry's priority and persist the updated `registries.yaml`.
- **FR-004**: When the value equals the registry's current priority, Sauron
  shall make no change and report success (no-op).

### State-driven

- **FR-005**: While setting a priority, Sauron shall leave the existing
  configuration unchanged until the change is persisted; `registries.yaml` is
  left untouched on any failure.

### Unwanted behavior

- **FR-006**: If the name or the value is missing, then Sauron shall reject
  the request and report what is required.
- **FR-007**: If the value is not a non-negative integer, then Sauron shall
  reject the request and report that a non-negative integer is required.
- **FR-008**: If no registry with the given name exists, then Sauron shall
  reject the request and report that the registry is not found.
- **FR-009**: If the value is already used by another registry, then Sauron
  shall reject the request, leave the configuration unchanged, and report that
  the priority must be unique (`0` is assignable only when no registry holds
  it).
- **FR-010**: If `registries.yaml` cannot be read or parsed, then Sauron shall
  reject the request and report that the registries cannot be read.
- **FR-011**: If only one registry exists, then Sauron shall reject the
  request and report that priority cannot be changed while a single registry
  exists — it keeps priority `0` (see
  [unified priority model](../AUTHORING.md#priority-model)).

## Key Entities

- **Registry**: a registered source of artifacts (see
  [add registry](../0001-add-registry/spec.md)), identified by its name.
  Its priority follows the unified model in
  [unified priority model](../AUTHORING.md#priority-model)
  — an optional non-negative integer, unique across all registries
  regardless of kind, where the first registry is `0` and a lower value means
  higher precedence; this feature is the only way to change it after the
  registry is added.

## Notes

Configuration is now split across files per the
[configuration data contract](../contracts/configuration.md); file references
updated accordingly.

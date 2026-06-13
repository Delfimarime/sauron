# Filesystem Registry Support

**Type:** capability

**Enables:** [add registry](../spec.md)

## Overview

Registries of kind `filesystem` are local directories holding artifacts
under `.skills/` and `.agents/`. This capability defines the
filesystem-specific behavior of [add registry](../spec.md): directory
existence and accessibility checks, artifact-presence validation, and stable
path resolution so re-registering the same directory yields the same recorded
`uri`. Common registration behavior (name, priority, persistence,
transactionality) is owned by the [feature spec](../spec.md).

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to register a filesystem
  directory as a registry source of artifacts.

### Event-driven

- **FR-002**: When a user submits a `uri` (a directory path), Sauron shall
  verify that the directory exists and is accessible before registering it.
- **FR-003**: When a user submits a `uri` (a directory path), Sauron shall
  verify that the directory contains at least one artifact under `.skills/` or
  `.agents/` before registering it.
- **FR-004**: When a user submits a `uri` (a directory path), Sauron shall
  resolve it to an absolute, symlink-resolved path and record that resolved
  path as the `uri`, so that registering the same directory again yields the
  same recorded `uri`.

### Unwanted behavior

- **FR-005**: If the directory does not exist or cannot be accessed, then
  Sauron shall reject the request, leave the configuration unchanged, and
  report that the registry cannot be accessed.
- **FR-006**: If the directory contains neither a populated `.skills/` nor a
  populated `.agents/`, then Sauron shall reject the request and report that
  no skills or agents were found.

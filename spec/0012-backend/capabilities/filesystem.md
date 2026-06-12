# Filesystem Backend Support

**Type:** capability
**Enables:** [set backend](../spec.md)

## Overview

A backend of kind `filesystem` is a local directory holding persona
definitions. This capability defines the filesystem-specific behavior of
[set backend](../spec.md): directory existence and accessibility
checks, stable path resolution so re-configuring the same directory yields the
same recorded location, and how the backend exposes a per-persona last-modified
timestamp. Common configuration behavior (singleton upsert, persistence,
transactionality, teardown) is owned by the [feature spec](../spec.md).

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to configure a local filesystem
  directory as the backend.

### Event-driven

- **FR-002**: When a user submits a directory path, Sauron shall verify that the
  directory exists and is accessible before configuring it.
- **FR-003**: When a user submits a directory path, Sauron shall resolve it to
  an absolute, symlink-resolved path and record that resolved path, so that
  configuring the same directory again yields the same recorded location.
- **FR-004**: When the catalog is refreshed, Sauron shall derive each persona's
  last-modified timestamp from the modification time (mtime) of that persona's
  definition file.

### Unwanted behavior

- **FR-005**: If the directory does not exist or cannot be accessed, then Sauron
  shall reject the request, leave the existing configuration unchanged, and
  report that the backend cannot be accessed.

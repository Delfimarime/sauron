# Filesystem Transport

**Type:** capability

**Enables:** [set registry](../spec.md)

**Enables:** [list catalogue](../../0004-list-catalogue/spec.md)

**Enables:** [install](../../0007-install-artifacts/spec.md)

## Overview

The filesystem transport reaches a registry rooted at a local directory. It
validates the source when set and reads artifact content for browsing,
installing, and reconciling, organized under `.skills/` and `.agents/`.

## Requirements

### Ubiquitous

- FR-001: Sauron shall reach filesystem registries at the URI's path, treating
  each directory under `.skills/` or `.agents/` as one skill or agent.
- FR-002: Sauron shall compute an artifact's `digest` from a hash of its directory
  content.

### Event-driven

- FR-003: When validating a filesystem registry, Sauron shall confirm the path
  exists, is readable, and hosts at least one skill or agent.

### Unwanted behavior

- FR-004: If the path does not exist or is not readable, then Sauron shall fail
  with a runtime error.

### Notes

A filesystem registry exposes no version history, so a filesystem artifact's
optional `version` is recorded only when the artifact declares one.

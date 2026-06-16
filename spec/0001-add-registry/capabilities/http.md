# HTTP Transport

**Type:** capability

**Enables:** [add registry](../spec.md)

**Enables:** [list catalogue](../../0005-list-catalogue/spec.md)

**Enables:** [install](../../0006-install-artifacts/spec.md)

## Overview

The http transport reaches a registry served over HTTP(S). It validates the source
at add time and fetches artifact content for browsing, installing, and
reconciling, organized under `.skills/` and `.agents/`.

## Requirements

### Ubiquitous

- FR-001: Sauron shall reach http registries over HTTP(S), supporting credentials
  passed as environment references and TLS options (`--skip-tls-verify`,
  `--ca-cert`, `--client-cert`, `--client-key`).
- FR-002: Sauron shall compute an artifact's `digest` from its content (the
  server's ETag when offered, otherwise a content hash).

### Event-driven

- FR-003: When validating an http registry, Sauron shall confirm the source is
  reachable and hosts at least one skill or agent.

### Optional

- FR-004: Where the server declares an artifact version, Sauron shall record it as
  the artifact's optional `version`.

### Unwanted behavior

- FR-005: If the source is unreachable, returns an error status, or fails TLS
  verification, then Sauron shall fail with a runtime error.

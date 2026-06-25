# HTTP Transport

**Type:** capability

**Enables:** [set registry](../spec.md)

**Enables:** [list catalogue](../../0004-list-catalogue/spec.md)

**Enables:** [install](../../0005-install-artifacts/spec.md)

## Overview

The http transport reaches a registry that implements the
[Sauron HTTP Registry API](../../contracts/registry-http-api.oas3.yaml) — a JSON
REST API served over HTTP(S). It validates the source when set and lists,
describes, and downloads artifacts through that API rather than browsing a raw
directory tree. The API exposes skills under `/skills` and agents under `/agents`.

## Requirements

### Ubiquitous

- FR-001: Sauron shall reach http registries over HTTP(S) using the
  [HTTP Registry API](../../contracts/registry-http-api.oas3.yaml), supporting
  HTTP Basic credentials passed as environment references and TLS options
  (`--skip-tls-verify`, `--ca-cert`, `--client-cert`, `--client-key`).
- FR-002: Sauron shall compute an artifact's `digest` from its downloaded content.

### Event-driven

- FR-003: When validating an http registry, Sauron shall confirm the API is
  reachable and lists at least one skill or agent.

### Optional

- FR-004: Where the API declares an artifact's version (the `Artifact-Version`
  response header), Sauron shall record it as the artifact's optional `version`.

### Unwanted behavior

- FR-005: If the source is unreachable, returns an error status, or fails TLS
  verification, then Sauron shall fail with a runtime error.

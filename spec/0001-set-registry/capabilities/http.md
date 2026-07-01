# HTTP Transport

**Type:** capability

**Enables:** [set registry](../spec.md)

**Enables:** [list catalogue](../../0004-list-catalogue/spec.md)

**Enables:** [install](../../0007-install-artifacts/spec.md)

## Overview

The http transport reaches a registry that implements the
[Sauron HTTP Registry API](../../contracts/registry-http-api.oas3.yaml) — a JSON
REST API served over HTTP(S). It validates the source when set and lists and
downloads artifacts through that API rather than browsing a raw
directory tree. The API exposes skills under `/skills` and agents under `/agents`.

## Requirements

### Ubiquitous

- FR-001: Sauron shall reach http registries over HTTP(S) using the
  [HTTP Registry API](../../contracts/registry-http-api.oas3.yaml), supporting
  HTTP Basic credentials passed as environment references and TLS options
  (`--skip-tls-verify`, `--ca-cert`, `--client-cert`, `--client-key`).
- FR-002: Sauron shall set an artifact's `version` to the version the registry
  declares for it in the artifact listing.

### Event-driven

- FR-003: When validating an http registry, Sauron shall confirm the API is
  reachable and lists at least one skill or agent.

### Unwanted behavior

- FR-004: If the source is unreachable, fails TLS verification, or returns a
  non-authentication error status, then Sauron shall fail with a runtime error.
- FR-005: If the source returns an authentication error status (401/403), then
  Sauron shall fail with a usage error.

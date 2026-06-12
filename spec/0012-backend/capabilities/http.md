# HTTP Backend Support

**Type:** capability
**Enables:** [set backend](../spec.md)

## Overview

A backend of kind `http` is an HTTP(S) server serving persona
definitions. This capability defines the HTTP-specific behavior of
[set backend](../spec.md): URL validation, the `HEAD` reachability
probe, HTTP Basic credentials supplied as environment references only, TLS and
mutual-TLS options, the network timeout, and how the backend exposes a
per-persona last-modified timestamp. Common configuration behavior (singleton
upsert, persistence, transactionality, teardown) is owned by the
[feature spec](../spec.md).

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to configure an HTTP(S) server
  as the backend.

### Event-driven

- **FR-002**: When a user submits a URL, Sauron shall verify it is a
  syntactically valid URL whose scheme is `http` or `https` before configuring
  it.
- **FR-003**: When a user submits a URL, Sauron shall verify the server is
  reachable by issuing an HTTP `HEAD` request — honoring the supplied
  authentication and TLS options — before configuring it.
- **FR-004**: When a username or password is supplied, Sauron shall accept only
  the `${env:VAR}` form, persist the reference literally, and resolve it from
  the environment at use time (see
  [Credentials via environment variables only](../../0001-add-registry/architecture/ADR-0001-credentials-via-env-only.md)).
- **FR-005**: When the `HEAD` reachability probe or a subsequent fetch runs,
  Sauron shall bound it by the configured timeout (default `30s`).
- **FR-006**: When the installed personas' definitions are refreshed from the
  backend, Sauron shall derive each persona's last-modified timestamp from the
  backend's `Last-Modified` response header or the equivalent index metadata for
  that persona's definition.

### Unwanted behavior

- **FR-007**: If the server cannot be reached (connection, TLS, or non-success
  `HEAD` response), then Sauron shall reject the request, leave the existing
  configuration unchanged, and report that the backend cannot be
  reached.
- **FR-008**: If a username or password is supplied as a raw value rather than
  `${env:VAR}`, then Sauron shall reject the request and report that only the
  `${env:VAR}` form is supported (see
  [Credentials via environment variables only](../../0001-add-registry/architecture/ADR-0001-credentials-via-env-only.md)).
- **FR-009**: If a `${env:VAR}` reference names a variable that is not set at
  set time, then Sauron shall reject the request and report that the variable is
  unset.
- **FR-010**: If `--timeout` is not a valid positive duration, then Sauron shall
  reject the request and report that a valid timeout is required.

## Decision Records

- [Credentials via environment variables only](../../0001-add-registry/architecture/ADR-0001-credentials-via-env-only.md)

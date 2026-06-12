# HTTP Registry Support

**Type:** capability

**Enables:** [add registry](../spec.md)

## Overview

Registries of kind `http` are HTTP(S) servers serving artifacts. This
capability defines the HTTP-specific behavior of
[add registry](../spec.md): URL validation, the `HEAD` reachability probe,
HTTP Basic credentials supplied as environment references only, TLS and mutual
TLS options, and the network timeout. Common registration behavior (name,
priority, persistence, transactionality) is owned by the
[feature spec](../spec.md).

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to register an HTTP(S) server
  as a registry source of artifacts.

### Event-driven

- **FR-002**: When a user submits a `uri`, Sauron shall verify it is a
  syntactically valid URL whose scheme is `http` or `https` before registering
  it.
- **FR-003**: When a user submits a `uri`, Sauron shall verify the server is
  reachable by issuing an HTTP `HEAD` request — honoring the supplied
  authentication and TLS options — before registering it.
- **FR-004**: When a username or password is supplied, Sauron shall accept only
  the `${env:VAR}` form, persist the reference literally, and resolve it from
  the environment at use time (see
  [Credentials via environment variables only](../architecture/ADR-0001-credentials-via-env-only.md)).
- **FR-005**: When the `HEAD` reachability probe or a subsequent fetch runs,
  Sauron shall bound it by the configured timeout (default `30s`).

### Unwanted behavior

- **FR-006**: If the `uri` is not a valid `http`/`https` URL, then Sauron shall
  reject the request and report that the `uri` is invalid.
- **FR-007**: If the server cannot be reached (connection, TLS, or non-success
  `HEAD` response), then Sauron shall reject the request, leave the
  configuration unchanged, and report that the registry cannot be reached.
- **FR-008**: If `--client-cert` is given without `--client-key` (or vice
  versa), then Sauron shall reject the request and report that both are
  required for mutual TLS.
- **FR-009**: If a username or password is supplied as a raw value rather than
  `${env:VAR}`, then Sauron shall reject the request and report that only the
  `${env:VAR}` form is supported (see
  [Credentials via environment variables only](../architecture/ADR-0001-credentials-via-env-only.md)).
- **FR-010**: If a `${env:VAR}` reference names a variable that is not set at
  add time, then Sauron shall reject the request and report that the variable
  is unset.
- **FR-011**: If a referenced CA, client-cert, or client-key file cannot be
  read, then Sauron shall reject the request and report that the file cannot
  be accessed.
- **FR-012**: If `--timeout` is not a valid positive duration, then Sauron
  shall reject the request and report that a valid timeout is required.

## Decision Records

- [Credentials via environment variables only](../architecture/ADR-0001-credentials-via-env-only.md)

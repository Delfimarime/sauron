# Git Backend Support

**Type:** capability

**Enables:** [set backend](../spec.md)

## Overview

A backend of kind `git` is a Git remote, reached over SSH, holding
persona definitions. This capability defines the git-specific behavior of
[set backend](../spec.md): SSH-only URI validation, authentication via
`--ssh-key` or the system's regular SSH credentials, the `git ls-remote`
reachability check, the network timeout, and how the backend exposes a
per-persona last-modified timestamp. Common configuration behavior (singleton
upsert, persistence, transactionality, teardown) is owned by the
[feature spec](../spec.md).

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to configure a Git remote,
  accessed over SSH, as the backend.

### Event-driven

- **FR-002**: When a user submits a git URI, Sauron shall verify it is a valid
  SSH-based git URI (scp-like `user@host:path` or an `ssh://` URL) before
  configuring it (see
  [SSH-only remotes](../../0001-add-registry/architecture/ADR-0002-ssh-only-remotes.md)).
- **FR-003**: When a user submits a git URI, Sauron shall verify the remote is
  reachable and authentication succeeds by running `git ls-remote` — honoring
  the configured timeout and SSH key — before configuring it.
- **FR-004**: When a network operation runs (e.g. `git ls-remote`), Sauron shall
  bound it by the configured timeout (default `30s`).
- **FR-005**: When the installed personas' definitions are refreshed from the
  backend, Sauron shall derive each persona's last-modified timestamp from the
  last commit that touched that persona's definition.

### Unwanted behavior

- **FR-006**: If the URI is not a valid SSH-based git URI (e.g. an
  `http(s)://`, `git://`, or `file://` scheme, or a malformed address), then
  Sauron shall reject the request and report that an SSH-based git URI is
  required (see
  [SSH-only remotes](../../0001-add-registry/architecture/ADR-0002-ssh-only-remotes.md)).
- **FR-007**: If the remote cannot be reached or authentication fails within the
  timeout, then Sauron shall reject the request, leave the existing
  configuration unchanged, and report that the backend cannot be
  reached.
- **FR-008**: If the `--ssh-key` file cannot be read, then Sauron shall reject
  the request and report that the key file cannot be accessed.
- **FR-009**: If `--timeout` is not a valid positive duration, then Sauron shall
  reject the request and report that a valid timeout is required.

### Optional

- **FR-010**: Where `--ssh-key` is provided, Sauron shall authenticate with that
  private key; where it is absent, Sauron shall use the system's regular SSH
  credentials (agent, `~/.ssh/config`, default keys).

## Decision Records

- [SSH-only remotes](../../0001-add-registry/architecture/ADR-0002-ssh-only-remotes.md)

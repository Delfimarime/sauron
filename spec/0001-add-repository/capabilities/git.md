# Git Repository Support

**Type:** capability
**Enables:** [add repository](../spec.md)

## Overview

Repositories of kind `git` are Git remotes reached over SSH. This capability
defines the git-specific behavior of [add repository](../spec.md): SSH-only
URI validation, authentication via `--ssh-key` or the system's regular SSH
credentials, the `git ls-remote` reachability check, and the network timeout.
Common registration behavior (name, priority, persistence, transactionality)
is owned by the [feature spec](../spec.md).

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to register a Git repository,
  accessed over SSH, as a repository source of artifacts.

### Event-driven

- **FR-002**: When a user submits a git URI, Sauron shall verify it is a valid
  SSH-based git URI (scp-like `user@host:path` or an `ssh://` URL) before
  registering it (see
  [SSH-only remotes](../architecture/ADR-0002-ssh-only-remotes.md)).
- **FR-003**: When a user submits a git URI, Sauron shall verify the remote is
  reachable and authentication succeeds by running `git ls-remote` — honoring
  the configured timeout and SSH key — before registering it.
- **FR-004**: When a network operation runs (e.g. `git ls-remote`), Sauron
  shall bound it by the configured timeout (default `30s`).

### Unwanted behavior

- **FR-005**: If the URI is not a valid SSH-based git URI (e.g. an
  `http(s)://`, `git://`, or `file://` scheme, or a malformed address), then
  Sauron shall reject the request and report that an SSH-based git URI is
  required (see
  [SSH-only remotes](../architecture/ADR-0002-ssh-only-remotes.md)).
- **FR-006**: If the remote cannot be reached or authentication fails within
  the timeout, then Sauron shall reject the request, leave the configuration
  unchanged, and report that the repository cannot be reached.
- **FR-007**: If the `--ssh-key` file cannot be read, then Sauron shall reject
  the request and report that the key file cannot be accessed.
- **FR-008**: If `--timeout` is not a valid positive duration, then Sauron
  shall reject the request and report that a valid timeout is required.

### Optional

- **FR-009**: Where `--ssh-key` is provided, Sauron shall authenticate with
  that private key; where it is absent, Sauron shall use the system's regular
  SSH credentials (agent, `~/.ssh/config`, default keys).

## Decision Records

- [SSH-only remotes](../architecture/ADR-0002-ssh-only-remotes.md)

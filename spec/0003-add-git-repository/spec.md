# Feature Specification: Add Git Repository

**Created**: 2026-06-09

**Status**: Draft

**Input**: "Allow a user to register a Git repository (over SSH) as a source of skills and agents that Sauron can watch and deliver."

## Overview

A person responsible for a team's agentic-AI setup needs to register a Git repository, reached over SSH, as a source of skills and agents, so that Sauron can watch it and keep the team's targets in sync with its latest contents.

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to register a Git repository, accessed over SSH, as a repository source of skills and agents.
- **FR-002**: Every registered repository shall have a name that is unique across all repositories, regardless of kind.
- **FR-003**: Every registered repository shall have a priority that is unique across all repositories, regardless of kind.

### Event-driven (*When*)

- **FR-004**: When a user submits a request to add a repository, Sauron shall require a name, a priority, and a git URI, and shall require `--kind git` to be specified (because `http` is the default kind).
- **FR-005**: When a user submits a git URI, Sauron shall verify it is a valid SSH-based git URI (scp-like `user@host:path` or an `ssh://` URL) before registering it. (See ADR-0001.)
- **FR-006**: When a user submits a git URI, Sauron shall verify the remote is reachable and authentication succeeds by running `git ls-remote` — honoring the configured timeout and SSH key — before registering it.
- **FR-007**: When `--ssh-key` is provided, Sauron shall authenticate with that private key; when it is absent, Sauron shall use the system's regular SSH credentials (agent, `~/.ssh/config`, default keys).
- **FR-008**: When a repository is registered, Sauron shall identify it by its name.
- **FR-009**: When a repository passes validation, Sauron shall persist it to its configuration (`~/.sauron/settings.json`) so it becomes a watched source.
- **FR-010**: When a repository is successfully registered, Sauron shall report success.
- **FR-011**: When a network operation runs (e.g. `git ls-remote`), Sauron shall bound it by the configured timeout (default 30s).

### State-driven (*While*)

- **FR-012**: While a repository is being validated, Sauron shall leave the existing configuration unchanged until validation succeeds.

### Unwanted-behavior (*If / then*)

- **FR-013**: If no git URI is provided, then Sauron shall reject the request and report that a git URI is required.
- **FR-014**: If the URI is not a valid SSH-based git URI (e.g. an `http(s)://`, `git://`, or `file://` scheme, or a malformed address), then Sauron shall reject the request and report that an SSH-based git URI is required. (See ADR-0001.)
- **FR-015**: If the remote cannot be reached or authentication fails within the timeout, then Sauron shall reject the request, leave the configuration unchanged, and report that the repository cannot be reached.
- **FR-016**: If no name is provided, then Sauron shall reject the request and report that a name is required.
- **FR-017**: If the name is not a valid slug (`^[a-z0-9]+(-[a-z0-9]+)*$`), then Sauron shall reject the request and report that the name format is invalid.
- **FR-018**: If no priority is provided, or it is not a positive integer, then Sauron shall reject the request and report that a valid priority is required.
- **FR-019**: If the name is already used by another repository, then Sauron shall reject the request, leave the configuration unchanged, and report that the name must be unique.
- **FR-020**: If the priority is already used by another repository, then Sauron shall reject the request, leave the configuration unchanged, and report that the priority must be unique.
- **FR-021**: If the `--ssh-key` file cannot be read, then Sauron shall reject the request and report that the key file cannot be accessed.
- **FR-022**: If `--timeout` is not a valid positive duration, then Sauron shall reject the request and report that a valid timeout is required.
- **FR-023**: If `--kind git` is not specified, then Sauron shall not treat the request as a git repository; because `http` is the default kind, `git` must be selected explicitly. (The `sauron add repository` interface is shared across kinds, each covered by its own feature.)
- **FR-024**: If the git-only flag `--ssh-key` is used with a non-`git` kind, then Sauron shall reject the request and report that the flag applies only to `git`.

## Key Entities

- **Git Repository**: A registered SSH git remote providing skills and/or agents, identified by its **name**. Carries:
  - **name** — unique slug, unique across all kinds.
  - **priority** — unique positive integer, unique across all kinds, lower = higher precedence.
  - **uri** — the SSH-based git URI. Not required to be unique.
  - **ssh** (optional) — the private-key file path used to authenticate; omitted means the system's regular SSH credentials are used.
  - **timeout** — a duration bounding network operations; defaults to 30s.

## Decision Records

- `architecture/ADR-0001-ssh-only-remotes.md` — git repositories are SSH remotes only.

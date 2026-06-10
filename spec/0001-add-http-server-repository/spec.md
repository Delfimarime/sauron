# Feature Specification: Add HTTP Server Repository

**Created**: 2026-06-09

**Status**: Draft

**Input**: "Allow a user to register an HTTP(S) server as a source of skills and agents that Sauron can watch and deliver."

## Overview

A person responsible for a team's agentic-AI setup needs to register an HTTP(S) server as a source of skills and agents, so that Sauron can watch it and keep the team's targets in sync with its latest contents.

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to register an HTTP(S) server as a repository source of skills and agents.
- **FR-002**: Every registered repository shall have a name that is unique across all repositories, regardless of kind.
- **FR-003**: Every registered repository shall have a priority that is unique across all repositories, regardless of kind.

### Event-driven (*When*)

- **FR-004**: When a user submits a request to add a repository, Sauron shall require a name, a priority, and a URL, and shall default the kind to `http` when none is given.
- **FR-005**: When a user submits a URL, Sauron shall verify it is a syntactically valid URL whose scheme is `http` or `https` before registering it.
- **FR-006**: When a user submits a URL, Sauron shall verify the server is reachable by issuing an HTTP `HEAD` request — honoring the supplied authentication and TLS options — before registering it.
- **FR-007**: When a username or password is supplied, Sauron shall accept only the `${env:VAR}` form, persist the reference literally, and resolve it from the environment at use time. (See ADR-0001.)
- **FR-008**: When a repository is registered, Sauron shall identify it by its name.
- **FR-009**: When a repository passes validation, Sauron shall persist it to its configuration (`~/.sauron/settings.yaml`) so it becomes a watched source.
- **FR-010**: When a repository is successfully registered, Sauron shall report success.
- **FR-025**: When the `HEAD` reachability probe (and future fetches) run, Sauron shall bound them by the configured timeout (default 30s).

### State-driven (*While*)

- **FR-011**: While a repository is being validated, Sauron shall leave the existing configuration unchanged until validation succeeds.

### Unwanted-behavior (*If / then*)

- **FR-012**: If no URL is provided, then Sauron shall reject the request and report that a URL is required.
- **FR-013**: If the URL is not a valid `http`/`https` URL, then Sauron shall reject the request and report that the URL is invalid.
- **FR-014**: If the server cannot be reached (connection, TLS, or non-success HEAD response), then Sauron shall reject the request, leave the configuration unchanged, and report that the repository cannot be reached.
- **FR-015**: If no name is provided, then Sauron shall reject the request and report that a name is required.
- **FR-016**: If the name is not a valid slug (`^[a-z0-9]+(-[a-z0-9]+)*$`), then Sauron shall reject the request and report that the name format is invalid.
- **FR-017**: If no priority is provided, or it is not a positive integer, then Sauron shall reject the request and report that a valid priority is required.
- **FR-018**: If the name is already used by another repository, then Sauron shall reject the request, leave the configuration unchanged, and report that the name must be unique.
- **FR-019**: If the priority is already used by another repository, then Sauron shall reject the request, leave the configuration unchanged, and report that the priority must be unique.
- **FR-020**: If `--client-cert` is given without `--client-key` (or vice versa), then Sauron shall reject the request and report that both are required for mutual TLS.
- **FR-021**: If a username or password is supplied as a raw value rather than `${env:VAR}`, then Sauron shall reject the request and report that only the `${env:VAR}` form is supported. (See ADR-0001.)
- **FR-022**: If a `${env:VAR}` reference names a variable that is not set at add time, then Sauron shall reject the request and report that the variable is unset.
- **FR-023**: If a referenced CA, client-cert, or client-key file cannot be read, then Sauron shall reject the request and report that the file cannot be accessed.
- **FR-024**: If any HTTP-only flag (`--username`, `--password`, `--skip-tls-verify`, `--ca-cert`, `--client-cert`, `--client-key`) is used with a non-`http` kind, then Sauron shall reject the request and report that the flag applies only to `http`.
- **FR-026**: If `--timeout` is not a valid positive duration, then Sauron shall reject the request and report that a valid timeout is required.

## Key Entities

- **HTTP Repository**: A registered HTTP(S) server providing skills and/or agents, identified by its **name**. Carries:
  - **name** — unique slug, unique across all kinds.
  - **priority** — unique positive integer, unique across all kinds, lower = higher precedence.
  - **url** — the `http`/`https` location. Not required to be unique.
  - **auth** (optional) — HTTP Basic credentials, where username and password are each an `${env:VAR}` reference resolved at use time (never a stored secret; see ADR-0001).
  - **tls** (optional) — skip-verify flag (default false), a CA bundle path (server trust), and a client cert + key path pair (mutual TLS).
  - **timeout** — a duration bounding the reachability probe and fetches; defaults to 30s.

## Decision Records

- `architecture/ADR-0001-credentials-via-env-only.md` — credentials are environment references only; no secret is persisted.

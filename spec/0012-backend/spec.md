# Backend

**Type:** feature
**Realized by:** [http](capabilities/http.md), [filesystem](capabilities/filesystem.md), [git](capabilities/git.md)

## Overview

A person responsible for a team's agentic-AI setup needs a single backend that
owns the team's persona definitions — the persona analog of a
[registry](../0001-add-registry/spec.md) — so that Sauron can pull a
[catalog](#key-entities) of personas and keep it current. The backend
is a **singleton** per instance: there is exactly one configured backend, and
re-configuring it overrides the previous one.

This feature owns two commands. `set backend [--kind http|filesystem|git]
<location>` configures the backend, mirroring
[add registry](../0001-add-registry/spec.md): `--kind` defaults to `http`;
`--username`/`--password` accept `${env:VAR}` references only (no raw secrets,
see [Credentials via environment variables only](../0001-add-registry/architecture/ADR-0001-credentials-via-env-only.md));
git is reached over SSH only (see
[SSH-only remotes](../0001-add-registry/architecture/ADR-0002-ssh-only-remotes.md)).
Sauron validates per-kind reachability before persisting and, on success, writes
the registry config to the settings. `unset backend [--keep-artifacts]`
tears the backend down: by default it removes the registry config, the local
catalog, all install records, and the delivered artifacts; with
`--keep-artifacts` it removes everything except the delivered artifacts.

Kind-specific validation, reachability, authentication, and the per-persona
last-modified timestamp are defined by the [http](capabilities/http.md),
[filesystem](capabilities/filesystem.md), and [git](capabilities/git.md)
capabilities. The catalog is a read-only mirror populated by
[sync personas](../0013-sync-personas/spec.md); the installed set is owned by
[select personas](../0014-select-personas/spec.md).

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to configure a backend —
  of kind `http`, `filesystem`, or `git` — as the singleton backend that owns
  persona definitions.
- **FR-002**: Sauron shall maintain at most one configured backend per
  instance.
- **FR-003**: Sauron shall provide the ability to tear down the configured
  backend.

### Event-driven

- **FR-004**: When a user submits a request to set the backend, Sauron
  shall require a location and default the kind to `http` when none is given (so
  `filesystem` and `git` must be selected explicitly via `--kind`).
- **FR-005**: When a backend passes validation, Sauron shall persist
  its configuration to the settings (`~/.sauron/settings.yaml`) so it becomes the
  active backend in subsequent runs.
- **FR-006**: When a backend is set while one is already configured,
  Sauron shall override the previous configuration with the new one (upsert).
- **FR-007**: When a backend is successfully set, Sauron shall report
  success with a single confirmation line on stdout.
- **FR-008**: When `unset backend` runs, Sauron shall remove the
  registry configuration, the local [catalog](#key-entities), and all
  [install records](../0014-select-personas/spec.md), and report what was torn
  down.
- **FR-009**: When `unset backend` runs without `--keep-artifacts`,
  Sauron shall additionally remove the delivered artifacts from the provider and
  from the [track file](../0006-sync-artifacts/data/configuration.md), reporting them in
  the shared plan/report format.
- **FR-010**: When `unset backend` runs with `--keep-artifacts`, Sauron
  shall remove the configuration, catalog, and install records but leave the
  delivered artifacts in place on the provider and in the track file.

### State-driven

- **FR-011**: While a backend is being validated, Sauron shall leave the
  existing configuration unchanged until validation succeeds.

### Unwanted behavior

- **FR-012**: When a user deletes a backend that does not exist, Sauron
  shall exit successfully and report that nothing was deleted.
- **FR-013**: If no location is provided to `set backend`, then Sauron
  shall reject the request and report that a location is required.
- **FR-014**: If `--kind` is given a value other than `http`, `filesystem`, or
  `git`, then Sauron shall reject the request and report that the kind is
  unknown.
- **FR-015**: If a username or password is supplied as a raw value rather than
  `${env:VAR}`, then Sauron shall reject the request and report that only the
  `${env:VAR}` form is supported (see
  [Credentials via environment variables only](../0001-add-registry/architecture/ADR-0001-credentials-via-env-only.md)).
- **FR-016**: If a `${env:VAR}` reference names a variable that is not set at set
  time, then Sauron shall reject the request, leave the existing configuration
  unchanged, and report that the variable is unset.
- **FR-017**: If the location is malformed for the selected kind, then Sauron
  shall reject the request and report that the location is invalid for the kind
  (see the kind capability for each form).
- **FR-018**: If the backend cannot be reached during validation, then Sauron
  shall reject the request, leave the existing configuration unchanged, and
  report that the backend cannot be reached.
- **FR-019**: If `--timeout` is not a valid positive duration, then Sauron shall
  reject the request and report that a valid timeout is required.
- **FR-020**: If required arguments or flags are missing or invalid, then Sauron
  shall exit with code 2 without executing the command.
- **FR-021**: If a command fails, then Sauron shall write exactly one
  human-readable message to stderr.

## Key Entities

- **Backend**: The singleton backend that owns persona definitions —
  the persona analog of a [registry](../0001-add-registry/spec.md). It
  carries:
  - **kind** — `http`, `filesystem`, or `git`; selects which capability
    validates, reaches, and reads the backend.
  - **location** — where the persona definitions live: a URL for
    [http](capabilities/http.md), a directory path for
    [filesystem](capabilities/filesystem.md), an SSH git URI for
    [git](capabilities/git.md).
  - **auth / transport** — kind-scoped credentials and transport settings
    (HTTP Basic env references and TLS, git SSH key); see the capabilities.
  - **timeout** — bound on network operations (default `30s`).
  - **lastSyncedAt** — when the catalog was last refreshed from this backend by
    [sync personas](../0013-sync-personas/spec.md).
- **Catalog**: The local read-only mirror of persona definitions pulled from the
  backend, populated by [sync personas](../0013-sync-personas/spec.md).
  Each entry records a per-persona `lastModifiedAt` (from the backend) and a
  `lastSyncedAt` (local). The catalog schema is defined in
  [data/configuration.md](data/configuration.md); the installed subset is owned
  by [select personas](../0014-select-personas/spec.md).

## Decision Records

- [Credentials via environment variables only](../0001-add-registry/architecture/ADR-0001-credentials-via-env-only.md)
- [SSH-only remotes](../0001-add-registry/architecture/ADR-0002-ssh-only-remotes.md)

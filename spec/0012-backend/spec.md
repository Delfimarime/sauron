# Set Backend

**Type:** feature

**Realized by:** [http](capabilities/http.md), [filesystem](capabilities/filesystem.md), [git](capabilities/git.md)

## Overview

A person responsible for a team's agentic-AI setup needs a single backend that
owns the team's persona definitions — the persona analog of a
[registry](../0001-add-registry/spec.md) — so that Sauron can browse and install
the personas the backend offers. The backend is a **singleton** per instance:
there is exactly one configured backend, and re-configuring it overrides the
previous one.

This feature owns two commands. `set backend [--kind http|filesystem|git]
<uri>` configures the backend, mirroring
[add registry](../0001-add-registry/spec.md): `--kind` defaults to `http`;
`--username`/`--password` accept `${env:VAR}` references only (no raw secrets,
see [Credentials via environment variables only](../0001-add-registry/architecture/ADR-0001-credentials-via-env-only.md));
git is reached over SSH only (see
[SSH-only remotes](../0001-add-registry/architecture/ADR-0002-ssh-only-remotes.md)).
Sauron validates per-kind reachability before persisting and, on success, writes
the backend connection to [backend.yaml](../contracts/configuration.md#backendyaml).
`unset backend [--keep-artifacts]` tears the backend down: by default it removes
the backend connection, the [installed personas](../0014-select-personas/spec.md),
all install records, and the delivered artifacts; with `--keep-artifacts` it
removes everything except the delivered artifacts.

Kind-specific validation, reachability, authentication, and the per-persona
last-modified timestamp are defined by the [http](capabilities/http.md),
[filesystem](capabilities/filesystem.md), and [git](capabilities/git.md)
capabilities. The backend offers persona definitions that are fetched live; it
persists no catalog — the set of available personas is a
[live view](../contracts/configuration.md#live-persona-view) assembled at command
time. The installed set is owned by
[set personas](../0014-select-personas/spec.md) and refreshed by
[sync personas](../0013-sync-personas/spec.md).

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
  shall require a `uri` and default the kind to `http` when none is given (so
  `filesystem` and `git` must be selected explicitly via `--kind`).
- **FR-005**: When a backend passes validation, Sauron shall persist
  its connection to [backend.yaml](../contracts/configuration.md#backendyaml) so
  it becomes the active backend in subsequent runs.
- **FR-006**: When a backend is set while one is already configured,
  Sauron shall override the previous configuration with the new one (upsert).
- **FR-007**: When a backend is successfully set, Sauron shall report
  success with a single confirmation line on stdout.
- **FR-008**: When `unset backend` runs, Sauron shall remove the
  backend connection from [backend.yaml](../contracts/configuration.md#backendyaml),
  the [installed personas](../0014-select-personas/spec.md) from
  [personas.yaml](../contracts/configuration.md#personasyaml), and all install
  records, and report what was torn down.
- **FR-009**: When `unset backend` runs without `--keep-artifacts`,
  Sauron shall additionally remove the delivered artifacts from the provider and
  from the [track file](../0006-sync-artifacts/data/configuration.md), reporting them in
  the shared plan/report format.
- **FR-010**: When `unset backend` runs with `--keep-artifacts`, Sauron
  shall remove the backend connection, the installed personas, and the install
  records but leave the delivered artifacts in place on the provider and in the
  track file.

### State-driven

- **FR-011**: While a backend is being validated, Sauron shall leave the
  existing configuration unchanged until validation succeeds.

### Unwanted behavior

- **FR-012**: When a user deletes a backend that does not exist, Sauron
  shall exit successfully and report that nothing was deleted.
- **FR-013**: If no `uri` is provided to `set backend`, then Sauron
  shall reject the request and report that a `uri` is required.
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
- **FR-017**: If the `uri` is malformed for the selected kind, then Sauron
  shall reject the request and report that the `uri` is invalid for the kind
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
  - **uri** — where the persona definitions live: a URL for
    [http](capabilities/http.md), a directory path for
    [filesystem](capabilities/filesystem.md), an SSH git URI for
    [git](capabilities/git.md).
  - **auth / transport** — kind-scoped credentials and transport settings
    (HTTP Basic env references and TLS, git SSH key); see the capabilities.
  - **timeout** — bound on network operations (default `30s`).
  - **last_synced_at** — when the installed personas' definitions were last
    refreshed from this backend by
    [sync personas](../0013-sync-personas/spec.md).
  The backend connection schema is owned by the
  [configuration data contract](../contracts/configuration.md#backendyaml); the
  feature's [data/configuration.md](data/configuration.md) links it.
- **Available personas**: The personas a user can browse or install are a
  [live view](../contracts/configuration.md#live-persona-view) — never persisted —
  assembled at command time from the installed personas plus a live fetch of the
  definitions the backend offers; when the backend is unreachable, only installed
  personas appear. The installed subset is owned by
  [set personas](../0014-select-personas/spec.md).

## Decision Records

- [Credentials via environment variables only](../0001-add-registry/architecture/ADR-0001-credentials-via-env-only.md)
- [SSH-only remotes](../0001-add-registry/architecture/ADR-0002-ssh-only-remotes.md)

## Notes

- **Persona-model redesign (intentional behavior change).** This feature was
  realigned to the canonical persona model in the
  [configuration data contract](../contracts/configuration.md):
  - **No persisted catalog.** The local read-only catalog mirror was removed.
    The backend no longer owns or populates a persisted catalog; it *offers*
    persona definitions that are fetched live, and the set of available personas
    is a [live view](../contracts/configuration.md#live-persona-view) computed at
    command time. FR-006 of the [http](capabilities/http.md),
    [filesystem](capabilities/filesystem.md) FR-004, and [git](capabilities/git.md)
    FR-005 derive a per-persona last-modified timestamp into the installed
    personas' stored definitions ([personas.yaml](../contracts/configuration.md#personasyaml))
    rather than into a catalog.
  - **`personaRegistry` → `backend`.** The persisted singleton block is renamed
    from `personaRegistry` to `backend`, persisted in
    [backend.yaml](../contracts/configuration.md#backendyaml) rather than in
    `settings.yaml`. FR-005 was redefined accordingly (persist the connection to
    `backend.yaml`).
  - **`location` → `uri`.** The source-location field is renamed from `location`
    to `uri` (snake_case, matching a registry's `uri`). FR-004, FR-013, and
    FR-017 were reworded to name `uri`.
  - **`unset backend` cascade.** FR-008 and FR-010 were redefined: teardown
    cascades to the [installed personas](../0014-select-personas/spec.md) in
    [personas.yaml](../contracts/configuration.md#personasyaml) (clearing
    `installed`) — and, unless `--keep-artifacts`, the install records and
    delivered artifacts — instead of to a persisted catalog, per the contract's
    [cross-file write semantics](../contracts/configuration.md#cross-file-write-semantics).
  - FR ids are preserved; content was redefined in place where the redesign
    removed the old behavior.

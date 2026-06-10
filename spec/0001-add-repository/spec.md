# Add Repository

**Type:** feature
**Realized by:** [http](capabilities/http.md), [filesystem](capabilities/filesystem.md), [git](capabilities/git.md)

## Overview

A person responsible for a team's agentic-AI setup needs to register sources of
artifacts — an HTTP(S) server, a filesystem directory, or a Git repository
reached over SSH — so that Sauron can watch them and keep the team's target in
sync with their latest contents. The `sauron add repository` command registers
any of these sources under a single interface: the user supplies a name, a
priority, a location, and a kind (defaulting to `http`), and Sauron validates
the source before persisting it to the settings. Kind-specific validation and
transport behavior are defined by the [http](capabilities/http.md),
[filesystem](capabilities/filesystem.md), and [git](capabilities/git.md)
capabilities.

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to register a repository — of
  kind `http`, `filesystem`, or `git` — as a source of artifacts.
- **FR-002**: Sauron shall require every registered repository to have a name
  that is unique across all repositories, regardless of kind.
- **FR-003**: Sauron shall require every registered repository to have a
  priority that is unique across all repositories, regardless of kind.

### Event-driven

- **FR-004**: When a user submits a request to add a repository, Sauron shall
  require a name, a priority, and a location, and shall default the kind to
  `http` when none is given (so `filesystem` and `git` must be selected
  explicitly via `--kind`).
- **FR-005**: When a repository is registered, Sauron shall identify it by its
  name.
- **FR-006**: When a repository passes validation, Sauron shall persist it to
  the settings (`~/.sauron/settings.yaml`) so it becomes a watched source in
  subsequent runs.
- **FR-007**: When a repository is registered, Sauron shall record its name and
  priority alongside its kind and location.
- **FR-008**: When a repository is successfully registered, Sauron shall report
  success with a single confirmation line on stdout.

### State-driven

- **FR-009**: While a repository is being validated, Sauron shall leave the
  existing configuration unchanged until validation succeeds.

### Unwanted behavior

- **FR-010**: If no name is provided, then Sauron shall reject the request and
  report that a name is required.
- **FR-011**: If the name is not a valid slug (`^[a-z0-9]+(-[a-z0-9]+)*$`),
  then Sauron shall reject the request and report that the name format is
  invalid.
- **FR-012**: If no location is provided, then Sauron shall reject the request
  and report that a location is required.
- **FR-013**: If no priority is provided, or the priority is not a positive
  integer, then Sauron shall reject the request and report that a valid
  priority is required.
- **FR-014**: If the name is already used by another repository, then Sauron
  shall reject the request, leave the configuration unchanged, and report that
  the name must be unique.
- **FR-015**: If the priority is already used by another repository, then
  Sauron shall reject the request, leave the configuration unchanged, and
  report that the priority must be unique.
- **FR-016**: If the submitted location is already used by another registered
  repository, then Sauron shall still register the new repository; duplicate
  locations are permitted and shall not, on their own, cause rejection.
- **FR-017**: If a kind-scoped flag is used with a kind it does not apply to,
  then Sauron shall reject the request and report that the flag applies only to
  its kind (see the [command-line contract](contracts/command-line.md) for each
  flag's scope).
- **FR-018**: If required arguments or flags are missing or invalid, then
  Sauron shall exit with code 2 without executing the command.
- **FR-019**: If a command fails, then Sauron shall write exactly one
  human-readable message to stderr.

## Key Entities

- **Repository**: A registered source of artifacts, identified by its
  **name**. Every repository, regardless of kind, carries:
  - **name** — a unique slug (`^[a-z0-9]+(-[a-z0-9]+)*$`), unique across all
    repositories regardless of kind.
  - **kind** — `http`, `filesystem`, or `git`; selects which capability
    validates and fetches from the source.
  - **location** — where the artifacts live: a URL for
    [http](capabilities/http.md), a directory path for
    [filesystem](capabilities/filesystem.md), an SSH git URI for
    [git](capabilities/git.md). Not required to be unique across repositories.
  - **priority** — a unique positive integer, unique across all repositories
    regardless of kind; lower value wins.

  Kind-specific attributes (HTTP auth/TLS/timeout, git SSH key/timeout, and
  filesystem path resolution) are defined in [http](capabilities/http.md),
  [filesystem](capabilities/filesystem.md), and [git](capabilities/git.md).

## Decision Records

- [Credentials via environment variables only](architecture/ADR-0001-credentials-via-env-only.md)
- [SSH-only remotes](architecture/ADR-0002-ssh-only-remotes.md)

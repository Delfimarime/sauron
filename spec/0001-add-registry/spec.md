# Add Registry

**Type:** feature

**Realized by:** [http](capabilities/http.md), [filesystem](capabilities/filesystem.md), [git](capabilities/git.md)

## Overview

A person responsible for a team's agentic-AI setup needs to register sources of
artifacts — an HTTP(S) server, a filesystem directory, or a Git repository
reached over SSH — so that Sauron can watch them and keep the team's provider in
sync with their latest contents. The `sauron add registry` command registers
any of these sources under a single interface: the user supplies a name, a
`uri`, and a kind (defaulting to `http`), optionally a priority, and Sauron
validates the source before persisting it to the configuration. Priority follows the
[unified priority model](../AUTHORING.md#priority-model):
it is optional, always defined, and unique across all registries — the first
registry takes `0` and an omitted value appends at the end (`max + 1`).
Kind-specific validation and transport behavior are defined by the
[http](capabilities/http.md), [filesystem](capabilities/filesystem.md), and
[git](capabilities/git.md) capabilities.

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to register a registry — of
  kind `http`, `filesystem`, or `git` — as a source of artifacts.
- **FR-002**: Sauron shall require every registered registry to have a name
  that is unique across all registries, regardless of kind.
- **FR-003**: Sauron shall require every registered registry to have a
  priority — a non-negative integer that is always defined and unique across all
  registries, regardless of kind — per the
  [unified priority model](../AUTHORING.md#priority-model).

### Event-driven

- **FR-004**: When a user submits a request to add a registry, Sauron shall
  require a name and a `uri`, treat `--priority` as optional, and default the
  kind to `http` when none is given (so `filesystem` and `git` must be selected
  explicitly via `--kind`).
- **FR-005**: When a registry is registered, Sauron shall identify it by its
  name.
- **FR-006**: When a registry passes validation, Sauron shall persist it to
  `~/.sauron/registries.yaml` so it becomes a watched source in
  subsequent runs.
- **FR-007**: When a registry is registered, Sauron shall record its name and
  priority alongside its kind and `uri`.
- **FR-008**: When a registry is successfully registered, Sauron shall report
  success with a single confirmation line on stdout.
- **FR-009**: When the first registry is registered (none exist yet), Sauron
  shall assign it priority `0`, whether `--priority` is omitted or given as `0`,
  per the
  [unified priority model](../AUTHORING.md#priority-model).
- **FR-010**: When a registry is registered while one or more registries
  already exist and `--priority` is omitted, Sauron shall assign it the value one
  greater than the highest existing priority (`max + 1`), so it appends at the
  end without colliding, per the
  [unified priority model](../AUTHORING.md#priority-model).

### State-driven

- **FR-011**: While a registry is being validated, Sauron shall leave the
  existing configuration unchanged until validation succeeds.

### Unwanted behavior

- **FR-012**: If no name is provided, then Sauron shall reject the request and
  report that a name is required.
- **FR-013**: If the name is not a valid slug (`^[a-z0-9]+(-[a-z0-9]+)*$`),
  then Sauron shall reject the request and report that the name format is
  invalid.
- **FR-014**: If no `uri` is provided, then Sauron shall reject the request
  and report that a `uri` is required.
- **FR-015**: If a `--priority` value is provided and it is not a non-negative
  integer, then Sauron shall reject the request and report that a valid priority
  is required; omitting `--priority` is valid and is never an error.
- **FR-016**: If the name is already used by another registry, then Sauron
  shall reject the request, leave the configuration unchanged, and report that
  the name must be unique.
- **FR-017**: If a provided `--priority` value is already used by another
  registry, then Sauron shall reject the request, leave the configuration
  unchanged, and report that the priority must be unique across all registries,
  per the
  [unified priority model](../AUTHORING.md#priority-model).
- **FR-018**: If the submitted `uri` is already used by another registered
  registry, then Sauron shall still register the new registry; duplicate
  `uri` values are permitted and shall not, on their own, cause rejection.
- **FR-019**: If a kind-scoped flag is used with a kind it does not apply to,
  then Sauron shall reject the request and report that the flag applies only to
  its kind (see the [command-line contract](contracts/command-line.md) for each
  flag's scope).
- **FR-020**: If required arguments or flags are missing or invalid, then
  Sauron shall exit with code 2 without executing the command.
- **FR-021**: If a command fails, then Sauron shall write exactly one
  human-readable message to stderr.
- **FR-022**: If `--priority` is given a value other than `0` while no
  registry exists yet, then Sauron shall reject the request and report that the
  first registry takes priority `0`, per the
  [unified priority model](../AUTHORING.md#priority-model).

## Key Entities

- **Registry**: A registered source of artifacts, identified by its
  **name**. Every registry, regardless of kind, carries:
  - **name** — a unique slug (`^[a-z0-9]+(-[a-z0-9]+)*$`), unique across all
    registries regardless of kind.
  - **kind** — `http`, `filesystem`, or `git`; selects which capability
    validates and fetches from the source.
  - **uri** — where the artifacts live: a URL for
    [http](capabilities/http.md), a directory path for
    [filesystem](capabilities/filesystem.md), an SSH git URI for
    [git](capabilities/git.md). Not required to be unique across registries.
  - **priority** — an optional non-negative integer, always defined and unique
    across all registries regardless of kind; the first registry is `0`, an
    omitted value appends at the end (`max + 1`); lower value wins. See the
    [unified priority model](../AUTHORING.md#priority-model).

  Kind-specific attributes (HTTP auth/TLS/timeout, git SSH key/timeout, and
  filesystem path resolution) are defined in [http](capabilities/http.md),
  [filesystem](capabilities/filesystem.md), and [git](capabilities/git.md).

## Decision Records

- [Credentials via environment variables only](architecture/ADR-0001-credentials-via-env-only.md)
- [SSH-only remotes](architecture/ADR-0002-ssh-only-remotes.md)

## Notes

- The registry source location is a single field named `uri`, unifying the
  previously per-kind names (`url` for http, `path` for filesystem, `uri` for
  git) and the former `<location>` positional argument. The value remains
  kind-shaped — an `http`/`https` URL for `http`, an absolute symlink-resolved
  path for `filesystem`, an SSH git URI for `git` — so only the name unifies,
  not what each kind accepts. The schema for this field is owned by the
  [configuration data contract](../contracts/configuration.md#registriesyaml).

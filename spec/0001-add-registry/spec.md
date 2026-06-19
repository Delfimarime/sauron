# Add Registry

**Type:** feature

**Realized by:** [git transport](capabilities/git.md)

**Realized by:** [http transport](capabilities/http.md)

**Realized by:** [filesystem transport](capabilities/filesystem.md)

## Overview

A developer needs to tell Sauron where artifacts come from before anything can be
installed. `add registry` registers a named source and its connection details,
validates that the source is reachable and well-formed, and persists it. The
registry's transport — `git`, `http`, or `filesystem` — selects how the source is
reached and validated; the mechanics of each transport live in the capabilities
this feature is realized by.

## Requirements

### Ubiquitous

- FR-001: Sauron shall register a source under a unique, path-safe name together
  with its URI and transport, persisting it as a `Registry` document in
  `registries.yaml`.
- FR-002: Sauron shall default the transport to `http` when `--kind` is not given.
- FR-003: Sauron shall store credential material as environment references
  (`${env:VAR}`) only, never resolving or persisting secret values.

### Event-driven

- FR-004: When a registry is added, Sauron shall validate that the source is
  reachable and hosts at least one skill or agent before persisting it.
- FR-005: When validation succeeds, Sauron shall report the registered name and
  transport on stdout.

### State-driven

- FR-006: While a registry is being validated, Sauron shall leave the existing
  state unchanged until validation succeeds.

### Unwanted behavior

- FR-007: If a registry with the given name already exists, then Sauron shall fail
  with a runtime error and leave the existing registry unchanged.
- FR-008: If the name is not path-safe, then Sauron shall exit with a usage error
  without contacting the source.
- FR-009: If required arguments or flags are missing or invalid, then Sauron shall
  exit with code 2 without executing the command.
- FR-010: If the source is unreachable or hosts no artifact, then Sauron shall
  fail with a runtime error and persist nothing.

### Optional

- FR-011: Where transport-specific authentication or TLS flags are provided,
  Sauron shall apply them when validating and persist them on the `Registry`
  document.
- FR-012: Where `--timeout` is provided, Sauron shall bound the validation network
  operation by it (default `30s`).
- FR-013: Where `--ref` is provided for a git registry, Sauron shall apply it when
  validating and persist it on the `Registry` document as `spec.ref`; when `--ref`
  is absent, the registry resolves from the repository's default branch.

## Key Entities

- **Registry** — the registered source, persisted as a `Registry` document; see
  the [state data contract](../contracts/state.md).
- **Transport** — `git`, `http`, or `filesystem`, selected by `--kind` and stored
  as `spec.transport`.

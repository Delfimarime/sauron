# Set Registry

**Type:** feature

**Status:** Built

**Realized by:** [git transport](capabilities/git.md)

**Realized by:** [http transport](capabilities/http.md)

**Realized by:** [filesystem transport](capabilities/filesystem.md)

## Overview

A developer needs to tell Sauron where artifacts come from before anything can be
installed. Sauron has a single registry. `set registry` configures that source
and its connection details, validates that the source is reachable and
well-formed, and persists it — replacing any registry already configured. The
registry's transport — `git`, `http`, or `filesystem` — selects how the source is
reached and validated; the mechanics of each transport live in the capabilities
this feature is realized by.

## Requirements

### Ubiquitous

- FR-001: Sauron shall persist the source's URI and transport as the single
  `Registry` document in `settings.yaml`.
- FR-002: Sauron shall default the transport to `http` when `--kind` is not given.
- FR-003: Sauron shall store a secret credential (password or token) as an
  environment reference (`${env:VAR}`) only, never resolving or persisting the
  secret value; a non-secret username may be a literal or a reference.

### Event-driven

- FR-004: When a registry is set, Sauron shall validate that the source is
  reachable and hosts at least one skill or agent before persisting it.
- FR-005: When validation succeeds, Sauron shall report the configured URI and
  transport on stdout.
- FR-014: When the `Registry` document is persisted, Sauron shall stamp
  `metadata.creationTimestamp` and `metadata.lastUpdatedTimestamp` as equal
  RFC3339 UTC instants taken from its clock.

### State-driven

- FR-006: While a registry is being validated, Sauron shall leave the existing
  state unchanged until validation succeeds.

### Unwanted behavior

- FR-007: When a registry is already configured, Sauron shall replace it only
  after the new source validates; if validation fails, Sauron shall leave the
  existing registry unchanged.
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

- **Registry** — the single configured source, persisted as a `Registry` document;
  see the [state data contract](../contracts/state.md).
- **Transport** — `git`, `http`, or `filesystem`, selected by `--kind` and stored
  as `spec.transport`.

## Notes

`set` is an upsert: with no registry configured it creates one, and with a
registry already configured it replaces it (FR-007). There is no separate "already
exists" failure, and the registry carries no user-given name — Sauron has exactly
one. FR-008 (path-safe name validation) is retired with the name; its identifier
is left unused rather than reassigned.

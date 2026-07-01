# Install Artifacts

**Type:** feature

**Status:** Specified

**Realized by:** [artifact versioning](capabilities/artifact-versioning.md)

**Depends on:** [set registry](../0001-set-registry/spec.md)

**Depends on:** [set provider](../0005-set-provider/spec.md)

## Overview

A developer installs named artifacts from the registry into the active provider.
`install` fetches each named skill or agent and places it under the provider at
`sauron-<name>`, recording it in the track file with its `version` and exact
path.

## Requirements

### Ubiquitous

- FR-001: Sauron shall install each named artifact of the given kind from the
  configured registry, placing it under the active provider at `sauron-<name>`
  and recording a tracked document for it.
- FR-002: Sauron shall record each installed artifact's `version` read from the
  source, exact `path`, and `installedAt`/`updatedAt`.

### Event-driven

- FR-003: When an artifact is already installed, Sauron shall reconcile it to the
  source (updating it if its `version` changed) rather than duplicate it.
- FR-004: When the install is applied, Sauron shall print the plan under the
  kind's heading (`skills:` or `agents:`), prefixed `+` for additions and `~` for
  updates, followed by a summary count.

### State-driven

- FR-005: While no provider is set, Sauron shall fail with a runtime error and
  install nothing.

### Unwanted behavior

- FR-006: If a named artifact is not offered by the registry, then Sauron shall
  report it and continue with the remaining names.
- FR-007: If the registry is unreachable, then Sauron shall fail with a runtime
  error.
- FR-008: If required arguments or flags are missing or invalid, then Sauron shall
  exit with code 2 without executing the command.

## Key Entities

- **Artifact** — the installed skill or agent; tracked per the
  [state data contract](../contracts/state.md).

# Install Artifacts

**Type:** feature

**Realized by:** [persona resolution](capabilities/persona-resolution.md)

**Realized by:** [artifact versioning](capabilities/artifact-versioning.md)

**Depends on:** [add registry](../0001-add-registry/spec.md)

**Depends on:** [set provider](../0012-set-provider/spec.md)

## Overview

A developer installs named artifacts from a registry into the active provider.
`install` fetches each named skill, agent, or persona and places it under the
provider at `sauron-<registry>-<name>`, recording it in the track file with its
content `digest`, optional `version`, exact path, and provenance. Installing a
**persona** resolves its membership and installs each member.

## Requirements

### Ubiquitous

- FR-001: Sauron shall install each named artifact of the given kind from the
  named registry, placing it under the active provider at
  `sauron-<registry>-<name>` and recording a tracked document for it.
- FR-002: Sauron shall record each installed artifact's source `registry`,
  content `digest`, exact `path`, and `installedAt`/`updatedAt`, and its optional
  `version` when available.
- FR-003: Sauron shall mark a directly-installed artifact's provenance `direct:
  true`.

### Event-driven

- FR-004: When a persona is installed, Sauron shall resolve its membership and
  install each member, adding the persona to each member's provenance `personas`
  list and recording the persona with a snapshot of the resolved membership.
- FR-005: When an artifact is already installed, Sauron shall reconcile it to the
  source (updating it if its `digest` changed) and update its provenance rather
  than duplicate it.
- FR-006: When the install is applied, Sauron shall print the plan grouped under
  `skills:`, `agents:`, and `personas:`, prefixed `+` for additions and `~` for
  updates, followed by a summary count.

### State-driven

- FR-007: While no provider is set, Sauron shall fail with a runtime error and
  install nothing.

### Unwanted behavior

- FR-008: If a named artifact is not offered by the registry, then Sauron shall
  report it and continue with the remaining names.
- FR-009: If the registry is unreachable, then Sauron shall fail with a runtime
  error.
- FR-010: If required arguments or flags are missing or invalid, then Sauron shall
  exit with code 2 without executing the command.

## Key Entities

- **Artifact** — the installed skill, agent, or persona; tracked per the
  [configuration data contract](../contracts/configuration.md).
- **Provenance** — `direct` plus the `personas` that bring an artifact in.
- **Membership** — the skills and agents a persona references.

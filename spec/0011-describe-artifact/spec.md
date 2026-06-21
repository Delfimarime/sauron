# Describe Artifact

**Type:** feature

**Status:** Specified

## Overview

A developer needs the full detail of one installed artifact. `describe skill`,
`describe agent`, and `describe persona` show a single installed artifact's fields,
with column selection.

## Requirements

### Ubiquitous

- FR-001: Sauron shall show the full detail of the named installed artifact of the
  chosen kind, reading from `track.yaml` — including its source registry, optional
  version, digest, path, and provenance.
- FR-002: Sauron shall, for a persona, include its resolved membership.

### Optional

- FR-003: Where `--fields <list>` is provided, Sauron shall show those fields in
  order, with the name first.

### Unwanted behavior

- FR-004: If no installed artifact of that kind and name exists, then Sauron shall
  fail with a runtime error.
- FR-005: If required arguments or flags are missing or invalid, then Sauron shall
  exit with code 2 without executing the command.

## Key Entities

- **Artifact** — the described skill, agent, or persona; tracked per the
  [state data contract](../contracts/state.md).

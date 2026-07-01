# Describe Artifact

**Type:** feature

**Status:** Specified

## Overview

A developer needs the full detail of one installed artifact. `describe skill` and
`describe agent` show a single installed artifact's fields, with column selection.

## Requirements

### Ubiquitous

- FR-001: Sauron shall show the full detail of the named installed artifact of the
  chosen kind, reading from `track.yaml` — including its version and path.

### Optional

- FR-002: Where `--fields <list>` is provided, Sauron shall show those fields in
  order, with the name first.

### Unwanted behavior

- FR-003: If no installed artifact of that kind and name exists, then Sauron shall
  fail with a runtime error.
- FR-004: If required arguments or flags are missing or invalid, then Sauron shall
  exit with code 2 without executing the command.

## Key Entities

- **Artifact** — the described skill or agent; tracked per the
  [state data contract](../contracts/state.md).

## Notes

- The single registry is a global setting (one registry, implicit), so describe
  output carries no registry field.

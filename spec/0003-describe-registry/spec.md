# Describe Registry

**Type:** feature

## Overview

A developer needs the full detail of one registered source. `describe registry`
shows a single registry's fields, with column selection.

## Requirements

### Ubiquitous

- FR-001: Sauron shall show the full detail of the named registry, reading from
  `registries.yaml`.
- FR-002: Sauron shall never display secret values — credential fields are shown
  as the stored environment reference.

### Optional

- FR-003: Where `--fields <list>` is provided, Sauron shall show those fields in
  order, with the name first.

### Unwanted behavior

- FR-004: If no registry of that name exists, then Sauron shall fail with a
  runtime error.
- FR-005: If required arguments or flags are missing or invalid, then Sauron shall
  exit with code 2 without executing the command.

## Key Entities

- **Registry** — the described source; see the
  [state data contract](../contracts/state.md).

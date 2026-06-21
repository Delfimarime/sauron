# List Catalogue

**Type:** feature

**Status:** Specified

**Depends on:** [add registry](../0001-add-registry/spec.md)

## Overview

Before installing, a developer needs to see what a registry offers. `list
catalogue` fetches a registry's offerings live and prints them as a paginated
table, scoped to one artifact kind. The catalogue is never persisted: it is always
fetched fresh, so the command requires the registry to be reachable and has no
offline behavior.

## Requirements

### Ubiquitous

- FR-001: Sauron shall fetch, live from the named registry, the skills, agents, or
  personas it offers — as selected by the kind noun — and print them as a table.
- FR-002: Sauron shall page results with `--offset` (leading results skipped,
  default `0`) and `--limit` (maximum returned), reporting the applied paging.

### Optional

- FR-003: Where `--search <term>` is provided, Sauron shall include only entries
  whose name contains the term (case-insensitive).
- FR-004: Where `--sort <field>` and `--order` are provided, Sauron shall order
  the entries before paging.

### Unwanted behavior

- FR-005: If the registry is unreachable, then Sauron shall fail with a runtime
  error (there is no offline catalogue).
- FR-006: If no registry of that name exists, then Sauron shall fail with a
  runtime error.
- FR-007: If required arguments or flags are missing or invalid, then Sauron shall
  exit with code 2 without executing the command.

## Key Entities

- **Catalogue** — the live view of a registry's offerings; not persisted.
- **Registry** — the source browsed; its connection is read from
  `registries.yaml` (see the
  [state data contract](../contracts/state.md)).

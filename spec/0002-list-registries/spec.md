# List Registries

**Type:** feature

## Overview

A developer needs to review which sources are registered. `list registries` prints
the registered registries as a table, with filtering, column selection, and
sorting.

## Requirements

### Ubiquitous

- FR-001: Sauron shall list every registered registry, one row each, reading from
  `registries.yaml`.
- FR-002: Sauron shall show the registry name and transport by default, and accept
  `--fields` to choose the displayed columns (the name column is always present
  and first).

### Optional

- FR-003: Where `--search <term>` is provided, Sauron shall include only
  registries whose name contains the term (case-insensitive).
- FR-004: Where `--sort <field>` and `--order` are provided, Sauron shall order
  the rows accordingly; `--sort` accepts `name` (default) and `transport`.

### State-driven

- FR-005: While no registry is registered, Sauron shall print an empty result and
  exit successfully.

### Unwanted behavior

- FR-006: If `registries.yaml` is unreadable, then Sauron shall fail with a
  runtime error.
- FR-007: If required arguments or flags are missing or invalid, then Sauron shall
  exit with code 2 without executing the command.

## Key Entities

- **Registry** — the listed source; see the
  [configuration data contract](../contracts/configuration.md).

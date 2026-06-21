# List Artifacts

**Type:** feature

**Status:** Specified

## Overview

A developer needs to see what is installed. `list skills`, `list agents`, and
`list personas` print the installed artifacts of a kind as a table, with
filtering, column selection, and sorting. This is the local counterpart to
[list catalogue](../0005-list-catalogue/spec.md), which browses a registry.

## Requirements

### Ubiquitous

- FR-001: Sauron shall list every installed artifact of the chosen kind, one row
  each, reading from `track.yaml`.
- FR-002: Sauron shall show the name and source registry by default, and accept
  `--fields` to choose the displayed columns (the name column is always present
  and first).

### Optional

- FR-003: Where `--search <term>` is provided, Sauron shall include only artifacts
  whose name contains the term (case-insensitive).
- FR-004: Where `--sort <field>` and `--order` are provided, Sauron shall order
  the rows; `--sort` accepts `name` (default), `registry`, and `updated`.

### State-driven

- FR-005: While nothing of the chosen kind is installed, Sauron shall print an
  empty result and exit successfully.

### Unwanted behavior

- FR-006: If `track.yaml` is unreadable, then Sauron shall fail with a runtime
  error.
- FR-007: If required arguments or flags are missing or invalid, then Sauron shall
  exit with code 2 without executing the command.

## Key Entities

- **Artifact** — the installed skill, agent, or persona; tracked per the
  [state data contract](../contracts/state.md).

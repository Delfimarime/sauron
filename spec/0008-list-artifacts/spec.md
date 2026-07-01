# List Artifacts

**Type:** feature

**Status:** Specified

## Overview

A developer needs to see what is installed. `list skills` and `list agents`
print the installed artifacts of a kind as a table, with filtering, column
selection, sorting, and pagination. This is the local counterpart to
[list catalogue](../0004-list-catalogue/spec.md), which browses a registry.

## Requirements

### Ubiquitous

- FR-001: Sauron shall list every installed artifact of the chosen kind, one row
  each, reading from `track.yaml`.
- FR-002: Sauron shall show the name by default, and accept `--fields` to choose
  the displayed columns (the name column is always present and first).
- FR-003: Sauron shall page the filtered, sorted rows with `--page` (1-based page
  number, default `1`) and `--limit` (page size, default `20`), and report the
  applied paging without a total count.

### Optional

- FR-004: Where `--search <term>` is provided, Sauron shall include only artifacts
  whose name contains the term (case-insensitive).
- FR-005: Where `--sort <field>` and `--order` are provided, Sauron shall order
  the rows; `--sort` accepts `name` (default) and `lastUpdatedAt`.

### State-driven

- FR-006: While nothing of the chosen kind is installed, Sauron shall print an
  empty result and exit successfully.

### Unwanted behavior

- FR-007: If `track.yaml` is unreadable, then Sauron shall fail with a runtime
  error.
- FR-008: If required arguments or flags are missing or invalid, then Sauron shall
  exit with code 2 without executing the command.

## Key Entities

- **Artifact** — the installed skill or agent; tracked per the
  [state data contract](../contracts/state.md).

## Notes

- The single registry is a global setting (one registry, implicit), so list
  output carries no registry column and `--sort` no longer offers a `registry`
  field.

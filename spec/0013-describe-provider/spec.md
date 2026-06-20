# Describe Provider

**Type:** feature

## Overview

A developer needs to see which provider is active. `describe provider` shows the
active provider's detail.

## Requirements

### Ubiquitous

- FR-001: Sauron shall show the active provider, reading the `Provider` document
  from `settings.yaml`.

### Optional

- FR-002: Where `--fields <list>` is provided, Sauron shall show those fields in
  order.

### State-driven

- FR-003: While no provider is set, Sauron shall report that none is set and exit
  successfully.

### Unwanted behavior

- FR-004: If `settings.yaml` is unreadable, then Sauron shall fail with a runtime
  error.
- FR-005: If required arguments or flags are missing or invalid, then Sauron shall
  exit with code 2 without executing the command.

## Key Entities

- **Provider** — the active destination; see the
  [state data contract](../contracts/state.md).

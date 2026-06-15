# Set Provider

**Type:** feature

## Overview

A developer chooses where artifacts are installed. `set provider` records the
single global provider. Setting a different provider migrates every installed
artifact from the current provider's directories to the new one's. A provider must
be set before any install or reconcile.

## Requirements

### Ubiquitous

- FR-001: Sauron shall record the chosen provider as the single `Provider`
  document in `settings.yaml`.

### Event-driven

- FR-002: When the provider changes, Sauron shall migrate every installed artifact
  to the new provider's directories, update each artifact's recorded `path`, and
  print the migration plan grouped under `skills:`, `agents:`, and `personas:`
  with a summary count.

### State-driven

- FR-003: While the chosen provider is already the active one, Sauron shall exit
  successfully and change nothing.

### Unwanted behavior

- FR-004: If the provider name is not one Sauron supports, then Sauron shall exit
  with a usage error.
- FR-005: If a migration step fails, then Sauron shall report it and continue,
  leaving the `Provider` setting updated and the track file consistent with what
  migrated.
- FR-006: If required arguments or flags are missing or invalid, then Sauron shall
  exit with code 2 without executing the command.

## Key Entities

- **Provider** — the single global destination, recorded as the `Provider`
  document; see the [configuration data contract](../contracts/configuration.md).

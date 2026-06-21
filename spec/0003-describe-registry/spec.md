# Describe Registry

**Type:** feature

**Status:** Built

## Overview

A developer needs the full detail of one registered source. `describe registry`
shows a single registry's fields, with column selection.

## Requirements

### Ubiquitous

- FR-001: Sauron shall show the full detail of the named registry, reading from
  `registries.yaml`.
- FR-002: Sauron shall never display secret values — credential fields are shown
  as the stored environment reference.
- FR-006: Sauron shall surface the registry's audit timestamps
  (`metadata.creationTimestamp`, `metadata.lastUpdatedTimestamp`) in the detail
  view when they are populated.

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

## Notes

- **Not-found is its own error class.** FR-004 (no registry of that name exists)
  maps to a `TypeNotFound` use-case error, which the single exit-code site
  resolves to exit 1 — distinct from a usage error (exit 2). `TypeNotFound` joins
  the existing non-usage classes (`conflict`/`unreachable`/`validation`/`io`) that
  already resolve to exit 1, so no new exit-code mapping arm is introduced; the
  type is reused by later `describe`/`get`-style features.
- **The `--fields` valid set agrees across spec, state, and contract:**
  `{name, transport, uri, ref, auth, tls, sshKey, timeout, creationTimestamp,
  lastUpdatedTimestamp}`, with `name` always present and first. The two audit
  timestamps display by default when populated (FR-006). No drift to reconcile.

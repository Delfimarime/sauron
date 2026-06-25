# Unset Registry

**Type:** feature

**Status:** Built

## Overview

A developer needs to disconnect the configured source. `unset registry` removes
the single `Registry` document, leaving every already-installed artifact in place
— removing the source does not remove what it delivered. Installed artifacts are
removed only with [uninstall](../0006-uninstall-artifacts/spec.md).

## Requirements

### Ubiquitous

- FR-001: Sauron shall remove the `Registry` document from `settings.yaml`.
- FR-002: Sauron shall leave every tracked artifact in `track.yaml` and in the
  provider untouched.

### Optional

- FR-004: Where `--dry-run` is provided, Sauron shall report what would be unset
  without changing state.

### Unwanted behavior

- FR-005: If no registry is configured, then Sauron shall exit successfully and
  report that nothing was unset.
- FR-006: If required arguments or flags are missing or invalid, then Sauron shall
  exit with code 2 without executing the command.

## Key Entities

- **Registry** — the removed source; see the
  [state data contract](../contracts/state.md).

## Notes

Unregistering preserves already-installed artifacts, per the Constitution
(Chapter II, Article 2): removing the source leaves what it delivered in place,
now unmanaged by any registry until a new one is set. This is the inverse of the
earlier design, in which deleting a source cascade-uninstalled everything it had
delivered; that cascade, and the shared cleaning step it depended on, are gone.

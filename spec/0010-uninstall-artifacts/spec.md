# Uninstall Artifacts

**Type:** feature

**Status:** Specified

**Depends on:** [install](../0007-install-artifacts/spec.md)

## Overview

A developer removes named artifacts they previously installed. `uninstall` removes
each named skill or agent from the provider and the track file. A dry run previews
the removal.

## Requirements

### Ubiquitous

- FR-001: Sauron shall remove each named installed artifact of the given kind,
  deleting it from the provider at its recorded `path` and from `track.yaml`.

### Event-driven

- FR-002: When the uninstall is applied, Sauron shall print the plan of removed
  artifacts grouped under `skills:` and `agents:`, prefixed `-`, followed by a
  summary count.

### Optional

- FR-003: Where `--dry-run` is provided, Sauron shall print the plan without
  changing the environment or the track file.

### Unwanted behavior

- FR-004: When a user uninstalls an artifact that is not installed, Sauron shall
  exit successfully and report that nothing was removed.
- FR-005: If required arguments or flags are missing or invalid, then Sauron shall
  exit with code 2 without executing the command.
- FR-006: If an individual artifact removal fails, then Sauron shall report it and
  continue.

## Key Entities

- **Artifact** — the removed skill or agent; tracked per the
  [state data contract](../contracts/state.md).

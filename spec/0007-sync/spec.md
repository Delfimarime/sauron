# Sync

**Type:** feature

**Status:** Specified

**Depends on:** [install](../0005-install-artifacts/spec.md)

## Overview

`sync` reconciles the installed set against the registry. It operates only on
artifacts already installed — it never installs something new on its own. For each
tracked artifact it refreshes changed content, repairs local drift, and removes
artifacts that have vanished upstream. An optional kind list scopes the run; a dry
run previews the plan.

## Requirements

### Ubiquitous

- FR-001: Sauron shall reconcile every tracked artifact against the registry,
  comparing the registry's content `digest` to the tracked one.
- FR-002: Sauron shall update an artifact whose source `digest` changed, and
  repair drift by reinstalling a tracked artifact whose provider files are missing
  while it still exists upstream.
- FR-003: Sauron shall remove a tracked artifact that no longer exists in the
  registry, deleting it from the provider and from `track.yaml`.

### Event-driven

- FR-005: When the reconcile is applied, Sauron shall print the plan grouped under
  `skills:` and `agents:`, prefixed `+` (added), `~` (updated), or `-` (removed),
  followed by a summary count.

### State-driven

- FR-008: While no provider is set, Sauron shall fail with a runtime error.

### Unwanted behavior

- FR-009: If an artifact cannot be fetched from the registry, then Sauron shall
  report the affected artifact and continue reconciling the rest.
- FR-010: If required arguments or flags are missing or invalid, then Sauron shall
  exit with code 2 without executing the command.

### Optional

- FR-006: Where a kind list (`skills`, `agents`) is given, Sauron
  shall reconcile only those kinds; with none, it reconciles all.
- FR-007: Where `--dry-run` is provided, Sauron shall print the plan without
  changing the environment or the track file.

## Key Entities

- **Installed set** — the tracked artifacts, per the
  [state data contract](../contracts/state.md).
- **Digest** — the content identity compared to detect change and drift.

## Notes

- FR-009 was a multi-registry requirement in the prior draft ("continue
  reconciling the reachable registries"). With a single registry it is narrowed
  to per-artifact resilience: an artifact that cannot be fetched is reported and
  the rest of the run continues.

# Sync

**Type:** feature

**Depends on:** [install](../0006-install-artifacts/spec.md)

## Overview

`sync` reconciles the installed set against its sources. It operates only on
artifacts already installed — it never installs something new on its own. For each
tracked artifact it refreshes changed content, repairs local drift, removes
artifacts that have vanished upstream, and re-resolves persona membership in full
(adding new members and removing dropped ones). An optional kind list scopes the
run; a dry run previews the plan.

## Requirements

### Ubiquitous

- FR-001: Sauron shall reconcile every tracked artifact against its source,
  comparing the source's content `digest` to the tracked one.
- FR-002: Sauron shall update an artifact whose source `digest` changed, and
  repair drift by reinstalling a tracked artifact whose provider files are missing
  while it still exists upstream.
- FR-003: Sauron shall remove a tracked artifact that no longer exists on its
  source, deleting it from the provider and from `track.yaml`.

### Event-driven

- FR-004: When a persona is reconciled, Sauron shall re-resolve its membership and
  install members added upstream, remove members dropped upstream (subject to
  provenance), refresh retained members, and update the persona's `members`
  snapshot.
- FR-005: When the reconcile is applied, Sauron shall print the plan grouped under
  `skills:`, `agents:`, and `personas:`, prefixed `+` (added), `~` (updated), or
  `-` (removed), followed by a summary count.

### Optional

- FR-006: Where a kind list (`skills`, `agents`, `personas`) is given, Sauron
  shall reconcile only those kinds; with none, it reconciles all.
- FR-007: Where `--dry-run` is provided, Sauron shall print the plan without
  changing the environment or the track file.

### State-driven

- FR-008: While no provider is set, Sauron shall fail with a runtime error.

### Unwanted behavior

- FR-009: If a source registry is unreachable, then Sauron shall report the
  affected artifacts and continue reconciling the reachable ones.
- FR-010: If required arguments or flags are missing or invalid, then Sauron shall
  exit with code 2 without executing the command.

## Key Entities

- **Installed set** — the tracked artifacts, per the
  [configuration data contract](../contracts/configuration.md).
- **Digest** — the content identity compared to detect change and drift.
- **Membership** — re-resolved per persona, additions and removals.

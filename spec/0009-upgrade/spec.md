# Upgrade

**Type:** feature

**Status:** Specified

**Depends on:** [install](../0006-install-artifacts/spec.md)

## Overview

`upgrade` is the non-destructive refresh of the installed set. It operates only on
artifacts already installed and never removes anything: it refreshes those whose
source content changed and adds newly-added persona members, but leaves vanished
artifacts and dropped persona members in place. An optional kind list scopes the
run; a dry run previews the plan.

## Requirements

### Ubiquitous

- FR-001: Sauron shall refresh every tracked artifact whose source content
  `digest` changed, updating it under the provider and in `track.yaml`.
- FR-002: Sauron shall never remove an artifact during upgrade, including
  artifacts that have vanished from their source.

### Event-driven

- FR-003: When a persona is upgraded, Sauron shall re-resolve its membership and
  install members added upstream and refresh existing members, but shall not
  remove members dropped upstream; it updates the persona's `members` snapshot to
  include the additions.
- FR-004: When the upgrade is applied, Sauron shall print the plan grouped under
  `skills:`, `agents:`, and `personas:`, prefixed `+` (added) or `~` (updated),
  followed by a summary count.

### State-driven

- FR-007: While no provider is set, Sauron shall fail with a runtime error.

### Unwanted behavior

- FR-008: If a source registry is unreachable, then Sauron shall report the
  affected artifacts and continue upgrading the reachable ones.
- FR-009: If required arguments or flags are missing or invalid, then Sauron shall
  exit with code 2 without executing the command.

### Optional

- FR-005: Where a kind list (`skills`, `agents`, `personas`) is given, Sauron
  shall upgrade only those kinds; with none, it upgrades all.
- FR-006: Where `--dry-run` is provided, Sauron shall print the plan without
  changing the environment or the track file.

## Key Entities

- **Installed set** — the tracked artifacts, per the
  [state data contract](../contracts/state.md).
- **Digest** — the content identity compared to detect change.
- **Membership** — re-resolved per persona, additions only.

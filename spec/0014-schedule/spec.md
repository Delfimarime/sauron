# Schedule

**Type:** feature

**Status:** Specified

**Realized by:** [crontab](capabilities/crontab.md)

**Depends on:** [sync](../0008-sync/spec.md)

**Depends on:** [upgrade](../0009-upgrade/spec.md)

## Overview

A developer wants the reconcile operations to run automatically. `schedule sync`
and `schedule upgrade` register an OS-crontab entry that runs the corresponding
command on a cron expression; `unschedule sync` and `unschedule upgrade` remove
them. Each operation has at most one schedule, recorded as a `Schedule` document.

## Requirements

### Ubiquitous

- FR-001: Sauron shall, for `schedule (sync|upgrade) <expression>`, register an
  OS-crontab entry that runs the corresponding command and record a `Schedule`
  document in `settings.yaml`.
- FR-002: Sauron shall, for `unschedule (sync|upgrade)`, remove the operation's
  crontab entry and its `Schedule` document.

### Event-driven

- FR-003: When a schedule for the operation already exists, Sauron shall replace
  it (its expression and crontab entry) rather than add a second.

### Unwanted behavior

- FR-004: If the cron expression is invalid, then Sauron shall exit with a usage
  error and register nothing.
- FR-005: When a user unschedules an operation that is not scheduled, Sauron shall
  exit successfully and report that nothing was removed.
- FR-006: If the OS crontab cannot be read or written, then Sauron shall fail with
  a runtime error.
- FR-007: If required arguments or flags are missing or invalid, then Sauron shall
  exit with code 2 without executing the command.

## Key Entities

- **Schedule** — the per-operation cron schedule, recorded as a `Schedule`
  document; see the [state data contract](../contracts/state.md).

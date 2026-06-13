# Schedule Persona Sync

**Type:** feature

**Depends on:** [sync personas](../0013-sync-personas/spec.md)

## Overview

A person responsible for a team's agentic-AI setup needs Sauron to run
[`sauron sync personas`](../0013-sync-personas/spec.md) on a schedule without
running it by hand, so that the installed personas' definitions stay current
automatically. Scheduling installs a managed entry in the operating system's
crontab (see [ADR-0001](architecture/ADR-0001-cron-via-os-crontab.md)).

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to schedule and unschedule
  [`sauron sync personas`](../0013-sync-personas/spec.md) via the operating
  system's crontab.

### Event-driven

- **FR-002**: When a user runs `schedule sync personas` with a cron expression,
  Sauron shall validate it and install or replace the managed crontab entry that
  runs `sauron sync personas`.
- **FR-003**: When the schedule is installed, Sauron shall record it under
  `schedules.sync_personas` in its configuration (`~/.sauron/settings.yaml`).
- **FR-004**: When the scheduled entry runs, it shall invoke `sauron sync
  personas` with no flags, refreshing the installed personas' definitions from
  the configured backend.
- **FR-005**: When the schedule is installed or replaced, Sauron shall report the
  active schedule.
- **FR-006**: When a user runs `unschedule sync personas`, Sauron shall remove
  the managed `sync personas` crontab entry and the `schedules.sync_personas`
  record, and report that the persona-sync schedule is removed.

### State-driven

- **FR-007**: While installing or removing the schedule, Sauron shall change only
  its own managed `sync personas` crontab entry, leaving any other crontab entries
  untouched.

### Unwanted behavior

- **FR-008**: If `schedule sync personas` is run without a cron expression, then
  Sauron shall reject the request and report that an expression is required.
- **FR-009**: If the cron expression is invalid, then Sauron shall reject the
  request and report that the expression is invalid.
- **FR-010**: If `unschedule sync personas` is run when no such schedule is
  installed, then Sauron shall make no change and report that it is already
  unscheduled (treated as success).
- **FR-011**: If the crontab cannot be read or written, then Sauron shall reject
  the request and report that the schedule cannot be updated.

## Key Entities

- **Schedule**: the cron expression under which
  [`sauron sync personas`](../0013-sync-personas/spec.md) runs, recorded under
  `schedules.sync_personas` in `~/.sauron/settings.yaml` and realized as a
  managed, marker-bracketed entry in the user's crontab.

## Decision Records

- [ADR-0001](architecture/ADR-0001-cron-via-os-crontab.md) — scheduling uses the
  operating system's crontab, one managed entry per operation.

## Notes

- Removing *all* sync schedules at once (`unschedule sync`, both artifacts and
  personas) is owned by
  [schedule artifact sync](../0011-schedule-sync/spec.md); this feature owns only
  the persona-specific `schedule sync personas` / `unschedule sync personas` pair.

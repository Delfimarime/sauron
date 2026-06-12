# Schedule Artifact Sync

**Type:** feature

**Depends on:** [sync artifacts](../0006-sync-artifacts/spec.md)

## Overview

A person responsible for a team's agentic-AI setup needs Sauron to run
[`sauron sync artifacts`](../0006-sync-artifacts/spec.md) on a schedule without
running it by hand, so that providers stay current automatically. Scheduling
installs a managed entry in the operating system's crontab (see
[ADR-0001](architecture/ADR-0001-cron-via-os-crontab.md)).

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to schedule and unschedule
  [`sauron sync artifacts`](../0006-sync-artifacts/spec.md) via the operating
  system's crontab.

### Event-driven

- **FR-002**: When a user runs `schedule sync artifacts` with a cron expression,
  Sauron shall validate it and install or replace the managed crontab entry that
  runs `sauron sync artifacts`.
- **FR-003**: When the schedule is installed, Sauron shall record it under
  `schedules.sync_artifacts` in its configuration (`~/.sauron/settings.yaml`).
- **FR-004**: When the scheduled entry runs, it shall invoke `sauron sync
  artifacts` with no flags, so the sync follows the configured global provider
  and the union of personas (or everything when no personas are installed).
- **FR-005**: When the schedule is installed or replaced, Sauron shall report the
  active schedule.
- **FR-006**: When a user runs `unschedule sync artifacts`, Sauron shall remove
  the managed `sync artifacts` crontab entry and the `schedules.sync_artifacts`
  record, and report that the artifact-sync schedule is removed.
- **FR-007**: When a user runs `unschedule sync` without an operation, Sauron
  shall remove every Sauron-managed sync schedule — both `sync artifacts` and
  [`sync personas`](../0019-schedule-sync-personas/spec.md) — and report what was
  removed.

### State-driven

- **FR-008**: While installing or removing a schedule, Sauron shall change only
  its own managed crontab entries, leaving any other crontab entries untouched.

### Unwanted behavior

- **FR-009**: If `schedule sync artifacts` is run without a cron expression, then
  Sauron shall reject the request and report that an expression is required.
- **FR-010**: If the cron expression is invalid, then Sauron shall reject the
  request and report that the expression is invalid.
- **FR-011**: If `unschedule sync artifacts` or `unschedule sync` is run when no
  matching schedule is installed, then Sauron shall make no change and report that
  it is already unscheduled (treated as success).
- **FR-012**: If the crontab cannot be read or written, then Sauron shall reject
  the request and report that the schedule cannot be updated.

## Key Entities

- **Schedule**: the cron expression under which
  [`sauron sync artifacts`](../0006-sync-artifacts/spec.md) runs, recorded under
  `schedules.sync_artifacts` in `~/.sauron/settings.yaml` and realized as a
  managed, marker-bracketed entry in the user's crontab.

## Decision Records

- [ADR-0001](architecture/ADR-0001-cron-via-os-crontab.md) — scheduling uses the
  OS crontab, one managed entry per operation, each with its own marker.

## Notes

- `unschedule sync` (no operation noun) removes both schedules. Unlike `prune` and
  `delete`, which require an artifact-type noun, the operation noun here is
  optional (omitted = all): unscheduling only stops future runs and never touches
  installed artifacts, so a clear-all convenience is safe.
- Scheduling the persona sync is a separate feature,
  [schedule sync personas](../0019-schedule-sync-personas/spec.md); this feature
  owns the cross-cutting `unschedule sync` (both) form.

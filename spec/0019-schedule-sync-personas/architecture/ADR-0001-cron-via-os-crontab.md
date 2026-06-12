# ADR-0001: Scheduling uses the operating system's crontab

**Status**: Accepted

**Date**: 2026-06-12

**Feature**: Schedule Persona Sync

## Context

Sauron must run [`sauron sync personas`](../../0013-sync-personas/spec.md) on a
schedule with no daemon of its own, on macOS and Linux. The options include the
user's crontab, macOS `launchd`, and systemd timers.

## Decision

Scheduling installs a **managed entry in the user's crontab** that runs
`sauron sync personas`, bracketed by an operation-qualified marker so Sauron only
touches its own line and never disturbs other crontab content (including the
`sync artifacts` schedule):

```
# managed by sauron: sync personas
0 6 * * * sauron sync personas
```

The schedule is also recorded under `schedules.sync_personas` in
`~/.sauron/settings.yaml`. `unschedule sync personas` removes both. This mirrors
the per-operation marker scheme of
[schedule artifact sync ADR-0002](../../0011-schedule-sync/architecture/ADR-0002-per-operation-crontab-markers.md).

## Consequences

**Positive**

- Works uniformly on macOS and Linux with no Sauron daemon.
- The operation-qualified marker isolates this schedule from the artifact-sync
  one, so each can be installed or removed independently.

**Negative**

- Relies on a running cron service; environments without one are unsupported.
- `launchd`/systemd-timer users get a crontab entry rather than a native unit.

## Revisit when

A native scheduler integration (`launchd` agent, systemd timer) is wanted, or the
persona sync needs arguments beyond the plain command.

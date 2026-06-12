# ADR-0001: Scheduling uses the OS crontab, one managed entry per operation

**Status**: Accepted

**Date**: 2026-06-12

**Feature**: Schedule Artifact Sync

## Context

`spec/README.md` states Sauron can run a sync automatically on a schedule. That
requires a scheduling mechanism. The options include the user's crontab, macOS
`launchd`, and systemd timers. Sauron is a command-line tool that should run on
macOS and Linux with no daemon of its own.

Sauron schedules more than one operation — artifact sync and
[persona sync](../../0019-schedule-sync-personas/spec.md) — each independently
installable and removable, so a single shared marker would not let one be removed
without disturbing the other.

## Decision

Scheduling installs a **managed entry in the user's crontab**, one per scheduled
operation, each bracketed by its own operation-qualified marker so Sauron only
ever reads, replaces, or removes its own line and never disturbs other crontab
content:

```
# managed by sauron: sync artifacts
0 * * * * sauron sync artifacts
# managed by sauron: sync personas
0 6 * * * sauron sync personas
```

`schedule sync <operation>` installs or replaces only the matching marker line;
`unschedule sync <operation>` removes only it; `unschedule sync` removes every
`# managed by sauron:` line. Each schedule is also recorded under
`schedules.<operation>` in `~/.sauron/settings.yaml` for visibility. The scheduled
command is plain (`sauron sync artifacts` / `sauron sync personas`), so it always
follows the current global provider and persona configuration.

## Consequences

**Positive**

- Works uniformly on macOS and Linux with no Sauron daemon.
- Schedules are managed independently; removing one never disturbs the other, and
  the marker namespace (`# managed by sauron: <operation>`) extends to future
  scheduled operations.

**Negative**

- Relies on a running cron service; environments without one are unsupported.
- `launchd`/systemd-timer users get a crontab entry rather than a native unit.
- Multiple managed lines to parse and keep consistent with `settings.yaml`.

## Revisit when

A native scheduler integration (`launchd` agent, systemd timer) is wanted, or a
scheduled operation needs arguments beyond the plain command.

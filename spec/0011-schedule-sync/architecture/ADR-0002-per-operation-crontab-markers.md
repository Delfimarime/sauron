# ADR-0002: One managed crontab entry per scheduled operation

**Status**: Accepted

**Date**: 2026-06-12

**Feature**: Schedule Artifact Sync

## Context

[ADR-0001](ADR-0001-cron-via-os-crontab.md) installed a single managed crontab
entry bracketed by one `# managed by sauron` marker, for the lone
`sauron sync artifacts` job. Sauron now schedules more than one operation —
artifact sync and
[persona sync](../../0019-schedule-sync-personas/spec.md) — each independently
installable and removable. A single shared marker cannot distinguish the two, so
unscheduling one would disturb the other.

## Decision

Each scheduled operation gets its **own managed crontab entry with its own
operation-qualified marker**:

```
# managed by sauron: sync artifacts
0 * * * * sauron sync artifacts
# managed by sauron: sync personas
0 6 * * * sauron sync personas
```

`schedule sync <operation>` installs or replaces only the matching marker line;
`unschedule sync <operation>` removes only it; `unschedule sync` removes every
`# managed by sauron:` line. Each schedule is also recorded under
`schedules.<operation>` in `~/.sauron/settings.yaml`. Supersedes ADR-0001.

## Consequences

**Positive**

- Schedules are managed independently; removing one never disturbs the other.
- Still no Sauron daemon; works on macOS and Linux via the user's crontab.
- The marker namespace (`# managed by sauron: <operation>`) extends to future
  scheduled operations.

**Negative**

- Multiple managed lines to parse and keep consistent with `settings.yaml`.
- Relies on a running cron service; environments without one are unsupported.

## Revisit when

A native scheduler integration (`launchd` agent, systemd timer) is wanted, or a
scheduled operation needs arguments beyond the plain command.

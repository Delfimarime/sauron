# ADR-0001: Scheduling uses the operating system's crontab

**Status**: Superseded by [ADR-0002](ADR-0002-per-operation-crontab-markers.md)

**Date**: 2026-06-10

**Feature**: Schedule Artifact Sync (formerly Cron Sync)

## Context

`spec/README.md` states Sauron can register a cron job so the whole process runs automatically on a schedule. That requires a scheduling mechanism. The options include the user's crontab, macOS `launchd`, and systemd timers. Sauron is a command-line tool that should run on macOS and Linux with no daemon of its own.

## Decision

Scheduling installs a **managed entry in the user's crontab** that runs `sauron sync artifacts`. The entry is bracketed by a `# managed by sauron` marker so Sauron only ever reads, replaces, or removes its own line and never disturbs other crontab content:

```
# managed by sauron
0 * * * * sauron sync artifacts
```

The schedule (the cron expression) is also recorded in `~/.sauron/settings.yaml` for visibility. `--disable` removes both the managed crontab entry and the recorded schedule entirely; re-enabling requires supplying a cron expression again.

## Consequences

**Positive**

- Works uniformly on macOS and Linux with no Sauron daemon.
- The managed marker keeps Sauron's footprint isolated and reversible.
- The scheduled command is plain `sauron sync artifacts`, so it always follows the current global provider and persona configuration.

**Negative**

- Relies on a running cron service; environments without one are unsupported.
- `launchd`/systemd-timer users get a crontab entry rather than a native unit.

## Revisit when

A native scheduler integration (`launchd` agent, systemd timer) is wanted, or scheduling needs to run more than the single `sauron sync artifacts` job.

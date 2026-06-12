# Contract: Command Line — Schedule Artifact Sync

Conventions: [CLI conventions](../../AUTHORING.md).

**Spec**: [Schedule Artifact Sync](../spec.md)

Defines the command-line interface for scheduling and unscheduling
[`sauron sync artifacts`](../../0006-sync-artifacts/spec.md) via the operating
system's cron. This is the user-facing contract only.

## Synopsis

```
sauron schedule sync artifacts <expression>
sauron unschedule sync artifacts
sauron unschedule sync
```

Command hierarchy: `sauron` (root) → `schedule`/`unschedule` (verb) → `sync`
(noun) → `artifacts` (noun). `unschedule sync` omits the operation noun to remove
every managed sync schedule.

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `<expression>` | Yes, for `schedule sync artifacts` | A cron expression (e.g. `0 * * * *`). Realizes [spec](../spec.md) FR-002, FR-009, FR-010. |

## Flags

None.

## Output

- **Success**: a single confirmation line — the active schedule when installed, or what was removed when unscheduled.
- **Failure**: a single human-readable message on stderr.

## Exit codes

Exit-status meanings are owned by the [CLI conventions](../../AUTHORING.md);
this table refines which conditions map to each code.

| Code | Meaning | Realizes |
|------|---------|----------|
| `0` | Schedule installed/replaced, removed, or an already-unscheduled no-op | [spec](../spec.md) FR-002, FR-006, FR-007, FR-011 |
| `2` | Usage error — `schedule sync artifacts` without an expression, or an invalid expression | [spec](../spec.md) FR-009, FR-010 |
| `1` | The crontab cannot be read or written | [spec](../spec.md) FR-012 |

## Examples

```
# Schedule hourly artifact sync
$ sauron schedule sync artifacts "0 * * * *"
Scheduled 'sauron sync artifacts' at: 0 * * * *

# Replace with a daily schedule
$ sauron schedule sync artifacts "0 3 * * *"
Scheduled 'sauron sync artifacts' at: 0 3 * * *

# Unschedule the artifact sync
$ sauron unschedule sync artifacts
Unscheduled 'sauron sync artifacts'.

# Unschedule every managed sync (artifacts + personas)
$ sauron unschedule sync
Unscheduled 'sauron sync artifacts' and 'sauron sync personas'.

# Unschedule when nothing is scheduled (no-op, exit 0)
$ sauron unschedule sync artifacts
'sauron sync artifacts' is not scheduled.

# Missing expression (usage error, exit 2)
$ sauron schedule sync artifacts
Error: schedule sync artifacts requires a cron expression

# Invalid expression (usage error, exit 2)
$ sauron schedule sync artifacts "every hour"
Error: invalid cron expression: 'every hour'
```

Scheduling the persona sync is documented separately in
[schedule sync personas](../../0019-schedule-sync-personas/contracts/command-line.md).

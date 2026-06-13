# Contract: Command Line — Schedule Persona Sync

Conventions: [CLI conventions](../../AUTHORING.md).

**Spec**: [Schedule Persona Sync](../spec.md)

Defines the command-line interface for scheduling and unscheduling
[`sauron sync personas`](../../0013-sync-personas/spec.md) via the operating
system's cron. This is the user-facing contract only.

## Synopsis

```
sauron schedule sync personas <expression>
sauron unschedule sync personas
```

Command hierarchy: `sauron` (root) → `schedule`/`unschedule` (verb) → `sync`
(noun) → `personas` (noun).

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `<expression>` | Yes, for `schedule sync personas` | A cron expression (e.g. `0 6 * * *`). Realizes [spec](../spec.md) FR-002, FR-008, FR-009. |

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
| `0` | Schedule installed/replaced, removed, or an already-unscheduled no-op | [spec](../spec.md) FR-002, FR-006, FR-010 |
| `2` | Usage error — `schedule sync personas` without an expression, or an invalid expression | [spec](../spec.md) FR-008, FR-009 |
| `1` | The crontab cannot be read or written | [spec](../spec.md) FR-011 |

## Examples

```
# Schedule a daily persona sync at 06:00
$ sauron schedule sync personas "0 6 * * *"
Scheduled 'sauron sync personas' at: 0 6 * * *

# Unschedule the persona sync
$ sauron unschedule sync personas
Unscheduled 'sauron sync personas'.

# Unschedule when nothing is scheduled (no-op, exit 0)
$ sauron unschedule sync personas
'sauron sync personas' is not scheduled.

# Missing expression (usage error, exit 2)
$ sauron schedule sync personas
Error: schedule sync personas requires a cron expression
```

Removing every sync schedule at once is `sauron unschedule sync`, owned by
[schedule artifact sync](../../0011-schedule-sync/contracts/command-line.md).

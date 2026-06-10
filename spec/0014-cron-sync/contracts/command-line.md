# Contract: Command Line — Cron Sync

Conventions: [CLI conventions](../../AUTHORING.md).

**Spec**: [Cron Sync](../spec.md)

Defines the command-line interface for scheduling `sauron sync` via the operating system's cron. This is the user-facing contract only.

## Synopsis

```
sauron cron sync <expression>
sauron cron sync --disable
```

Command hierarchy: `sauron` (root) → `cron` (group) → `sync` (subcommand).

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `<expression>` | Conditional | A cron expression (e.g. `0 * * * *`). Required unless `--disable` is given. Realizes FR-002, FR-008, FR-010. |

## Flags

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--disable` | No | false | Remove the managed schedule. Mutually exclusive with `<expression>`. Realizes FR-006, FR-009. |

## Output

- **Success**: a single confirmation line — the active schedule when installed, or that scheduling is disabled.
- **Failure**: a single human-readable message on stderr.

## Exit codes

Exit-status meanings are owned by the [CLI conventions](../../AUTHORING.md);
this table refines which conditions map to each code.

| Code | Meaning | Realizes |
|------|---------|----------|
| `0` | Schedule installed/replaced, or disabled (including the already-disabled no-op) | FR-002, FR-006, FR-011 |
| `2` | Usage error — neither expression nor `--disable`, both provided, or an invalid expression | FR-008, FR-009, FR-010 |
| `1` | The crontab cannot be read or written | FR-012 |

## Examples

```
# Schedule hourly sync
$ sauron cron sync "0 * * * *"
Scheduled 'sauron sync' at: 0 * * * *

# Replace with a daily schedule
$ sauron cron sync "0 3 * * *"
Scheduled 'sauron sync' at: 0 3 * * *

# Disable
$ sauron cron sync --disable
Scheduled sync disabled.

# Disable when nothing is scheduled (no-op, exit 0)
$ sauron cron sync --disable
Scheduled sync is already disabled.

# Neither expression nor --disable (usage error, exit 2)
$ sauron cron sync
Error: provide a cron expression or --disable

# Invalid expression (usage error, exit 2)
$ sauron cron sync "every hour"
Error: invalid cron expression: 'every hour'
```

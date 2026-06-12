# Data Model: Configuration — Sauron Settings (Cron Sync)

**Spec**: `../spec.md` (Cron Sync)
**Status**: Draft

Describes how the Cron Sync feature records the schedule and the operating-system crontab entry it manages.

## Recorded schedule — `~/.sauron/settings.yaml`

- **Path**: `~/.sauron/settings.yaml` (home directory resolved per platform).
- **Format**: a single YAML document.

Top-level field:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `cron` | object | No | Scheduling configuration; absent when nothing is scheduled. |

`cron` object:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `sync` | string | No | The cron expression for the scheduled `sauron sync artifacts`. Realizes FR-003. |

Example:

```yaml
provider: claude
cron:
  sync: "0 * * * *"
```

## Managed crontab entry

The user's crontab holds the executable schedule, bracketed by a managed marker so Sauron only touches its own line (see ADR-0001):

```
# managed by sauron
0 * * * * sauron sync artifacts
```

- Installing or replacing the schedule rewrites this single entry. Realizes FR-002, FR-007.
- `--disable` removes both the crontab entry and the `cron` section of `settings.yaml`. Realizes FR-006.
- The scheduled command is plain `sauron sync artifacts` — it carries no flags and follows the configured global provider and personas. Realizes FR-004.

## Write semantics

- The `settings.yaml` `cron` section is written atomically: serialize to a temporary file in `~/.sauron/`, then rename over `settings.yaml`.
- The crontab is updated through the system's crontab facility; only the managed entry is changed.
- On any failure, neither the crontab nor `settings.yaml` is left partially changed.

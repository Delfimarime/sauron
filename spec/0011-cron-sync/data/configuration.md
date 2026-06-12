# Data Model: Configuration — Cron Sync (`settings.yaml`)

**Spec**: [Cron Sync](../spec.md)

Cron Sync owns the `cron` block of `settings.yaml` (`cron.expression`), the
recorded sync schedule, and the managed entry it installs into the operating
system's crontab. The schema is owned by the
[configuration data contract](../../contracts/configuration.md#settingsyaml);
this document does not restate it.

## Reads

- The `cron` block of `settings.yaml`, to determine whether a schedule is
  currently installed.

## Owns / Writes

- The `cron` block of `settings.yaml` (`cron.expression`) — the recorded sync
  schedule. Installing or replacing a schedule records the expression; `--disable`
  removes the block. Realizes FR-002, FR-003, FR-006, FR-007.

## Managed crontab entry

The operating system's crontab is an OS resource, not a configuration file. The
user's crontab holds the executable schedule, bracketed by a managed marker so
Sauron only touches its own line (see
[ADR-0001](../architecture/ADR-0001-cron-via-os-crontab.md)):

```
# managed by sauron
0 * * * * sauron sync artifacts
```

- Installing or replacing the schedule rewrites this single managed entry; the
  recorded `cron.expression` in `settings.yaml` is the schedule's source of
  truth. Realizes FR-002, FR-007.
- `--disable` removes both the managed crontab entry and the `cron` block of
  `settings.yaml`. Realizes FR-006.
- The scheduled command is plain `sauron sync artifacts` — it carries no flags
  and follows the configured global provider and personas. Realizes FR-004.

## Notes

Configuration is now split across files per the
[configuration data contract](../../contracts/configuration.md); file references
updated accordingly.

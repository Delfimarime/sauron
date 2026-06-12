# Data Model: Configuration — Schedule Artifact Sync (`settings.yaml`)

**Spec**: [Schedule Artifact Sync](../spec.md)

Schedule Artifact Sync owns the `schedules.sync_artifacts` block of
`settings.yaml` (the recorded artifact-sync schedule) and the managed entry it
installs into the operating system's crontab. The schema is owned by the
[configuration data contract](../../contracts/configuration.md#settingsyaml);
this document does not restate it.

## Reads / Owns / Writes

- `settings.yaml` `schedules.sync_artifacts.expression` — the recorded
  artifact-sync schedule. `schedule sync artifacts` records the expression;
  `unschedule sync artifacts` removes it; `unschedule sync` removes it together
  with `schedules.sync_personas`.

## Managed crontab entry

The operating system's crontab is an OS resource, not a configuration file. The
user's crontab holds the executable schedule, bracketed by a per-operation marker
so Sauron only touches its own line (see
[ADR-0002](../architecture/ADR-0002-per-operation-crontab-markers.md)):

```
# managed by sauron: sync artifacts
0 * * * * sauron sync artifacts
```

- Installing or replacing the schedule rewrites this single managed entry; the
  recorded `schedules.sync_artifacts.expression` is the schedule's source of
  truth. Realizes [spec](../spec.md) FR-002, FR-008.
- `unschedule sync artifacts` removes this marker line and the
  `schedules.sync_artifacts` record; `unschedule sync` additionally removes the
  `sync personas` marker and `schedules.sync_personas`. Realizes
  [spec](../spec.md) FR-006, FR-007.
- The scheduled command is plain `sauron sync artifacts` — it carries no flags.
  Realizes [spec](../spec.md) FR-004.

## Notes

Configuration is split across files per the
[configuration data contract](../../contracts/configuration.md); the recorded
schedule moved from the former single `cron` block to `schedules.sync_artifacts`.

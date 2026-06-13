# Data Model: Configuration — Schedule Persona Sync (`settings.yaml`)

**Spec**: [Schedule Persona Sync](../spec.md)

Schedule Persona Sync owns the `schedules.sync_personas` block of `settings.yaml`
(the recorded persona-sync schedule) and the managed entry it installs into the
operating system's crontab. The schema is owned by the
[configuration data contract](../../contracts/configuration.md#settingsyaml);
this document does not restate it.

## Reads / Owns / Writes

- `settings.yaml` `schedules.sync_personas.expression` — the recorded
  persona-sync schedule. `schedule sync personas` records the expression;
  `unschedule sync personas` removes it. (`unschedule sync`, owned by
  [schedule artifact sync](../../0011-schedule-sync/spec.md), also removes it.)

## Managed crontab entry

The operating system's crontab holds the executable schedule, bracketed by a
per-operation marker so Sauron only touches its own line (see
[ADR-0001](../architecture/ADR-0001-cron-via-os-crontab.md)):

```
# managed by sauron: sync personas
0 6 * * * sauron sync personas
```

- Installing or replacing rewrites this single managed entry; the recorded
  `schedules.sync_personas.expression` is the source of truth. Realizes
  [spec](../spec.md) FR-002, FR-007.
- `unschedule sync personas` removes this marker line and the record. Realizes
  [spec](../spec.md) FR-006.
- The scheduled command is plain `sauron sync personas` — no flags. Realizes
  [spec](../spec.md) FR-004.

## Notes

Configuration is split across files per the
[configuration data contract](../../contracts/configuration.md); the schedule
lives under `schedules.sync_personas`.

# Schedule — configuration

This feature **writes** `settings.yaml`. The document schema is owned by the
[configuration data contract](../../contracts/configuration.md).

## Writes

- `settings.yaml`: upsert (on `schedule`) or remove (on `unschedule`) the
  `Schedule` document whose `metadata.name` is the operation (`sync` or
  `upgrade`). The write is atomic and lock-guarded, and mirrors the OS-crontab
  entry the [crontab capability](../capabilities/crontab.md) manages.

## Field realization

| Field | Requirement |
|---|---|
| `Schedule.metadata.name` | FR-001, FR-002, FR-003 |
| `Schedule.spec.cron` | FR-001, FR-003, FR-004 |

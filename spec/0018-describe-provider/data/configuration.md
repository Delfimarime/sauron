# Data Model: Configuration — Describe Provider (`settings.yaml`)

**Spec**: [Describe Provider](../spec.md)

Describe Provider reads the active `provider` from `settings.yaml`; it never
writes. The schema is owned by the
[configuration data contract](../../contracts/configuration.md#settingsyaml);
this document does not restate it.

## Reads

- The `provider` field of `settings.yaml` — the active provider; an absent value
  resolves to the default `claude`. Realizes FR-003, FR-005.

## Owns / Writes

- Nothing. Describing is read-only and offline.

## Notes

The displayed `skills-location` and `agents-location` are not persisted fields:
they are the resolved provider's target locations, derived from the active
`provider` per [set provider](../../0009-set-provider/spec.md). Realizes FR-004.

Configuration is now split across files per the
[configuration data contract](../../contracts/configuration.md); file references
updated accordingly.

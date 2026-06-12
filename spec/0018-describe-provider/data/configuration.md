# Data Model: Configuration — Describe Provider

**Spec**: [Describe Provider](../spec.md)

Describes how the Describe Provider feature reads the persisted configuration. It
is read-only: it never writes the settings or the track file.

## Schema

This feature defines no schema of its own. It reads the active provider and that
provider's skills/agents target locations from the provider schema owned by
[set provider](../../0009-set-provider/data/configuration.md), which is the
authoritative source for the `provider` field in `~/.sauron/settings.yaml`, its
default (`claude`), and where each provider persists skills and agents.

## Operation

- The described provider is the active `provider` value; an absent value
  resolves to the default `claude`. Realizes [spec](../spec.md) FR-003, FR-005.
- The `skills-location` and `agents-location` fields are the resolved
  provider's target locations as defined by
  [set provider](../../0009-set-provider/data/configuration.md). Realizes
  [spec](../spec.md) FR-004.
- If the settings exist but cannot be read or parsed, the request is rejected.
  Realizes [spec](../spec.md) FR-009.

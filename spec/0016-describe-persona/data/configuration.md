# Data Model: Configuration — Describe Persona (personas.yaml + live backend)

**Spec**: [Describe Persona](../spec.md)

Describe Persona reads an installed persona from `personas.yaml` or fetches it
live from the backend; it never writes. The schema is owned by the
[configuration data contract](../../contracts/configuration.md#personasyaml),
and the resolution is the [live view](../../contracts/configuration.md#live-persona-view).

## Reads
- `personas.yaml` `items` — for an installed persona (offline-capable).
- the [backend](../../contracts/configuration.md#backendyaml) live — for a persona that is not installed.

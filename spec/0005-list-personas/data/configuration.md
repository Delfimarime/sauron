# Data Model: Configuration — List Personas (personas.yaml + live backend)

**Spec**: [List Personas](../spec.md)

List Personas reads the installed personas from `personas.yaml` and merges
them with a live fetch from the backend; it never writes. The schema is owned
by the [configuration data contract](../../contracts/configuration.md#personasyaml),
and the available-personas assembly is the [live view](../../contracts/configuration.md#live-persona-view).

## Reads
- `personas.yaml` `items` — always shown.
- the [backend](../../contracts/configuration.md#backendyaml) live — when reachable; skipped offline.

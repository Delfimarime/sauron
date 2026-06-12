# Data Model: Configuration — Describe Artifact (track.yaml)

**Spec**: [Describe Artifact](../spec.md)

Describe Artifact reads one managed artifact's entry from the
[track file](../../contracts/configuration.md#trackyaml); it never writes. The
schema is owned by the
[configuration data contract](../../contracts/configuration.md#trackyaml); this
document does not restate it.

## Reads

- `track.yaml` `items` — the entry whose `type` and `name` match the request,
  contributing `name`, `type`, source `registry`, `provider`, `path`, `pinned`,
  and `persona`.

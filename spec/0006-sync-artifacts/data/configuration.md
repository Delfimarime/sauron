# Data Model: Configuration — Sync Artifacts (track.yaml, registries.yaml, personas.yaml, settings.yaml)

**Spec**: [Sync](../spec.md)

This feature owns `track.yaml` — it creates and maintains the record of
installed artifacts and their provenance — and reads `registries.yaml`,
`personas.yaml`, and `settings.yaml` to compute the desired set and the active
provider. The schema is owned by the
[configuration data contract](../../contracts/configuration.md#trackyaml); this
document does not restate it.

## Reads

- `registries.yaml` `items`: the sources of artifacts; `priority`
  resolves same-named artifacts
  ([ADR-0001](../architecture/ADR-0001-conflict-resolution-by-registry-priority.md),
  [#registriesyaml](../../contracts/configuration.md#registriesyaml)).
- `personas.yaml` `items`: the installed personas and their stored
  definitions — the desired set when no `--persona` narrows it; persona ordering
  follows the [priority model](../../AUTHORING.md#priority-model)
  ([#personasyaml](../../contracts/configuration.md#personasyaml)).
- `settings.yaml` `provider`: the active provider to deliver to (`claude` by
  default; managed by [set provider](../../0009-set-provider/spec.md)). Realizes
  [spec](../spec.md) FR-007
  ([#settingsyaml](../../contracts/configuration.md#settingsyaml)).

## Owns

- `track.yaml` `items`: the delivered artifacts and their provenance —
  `type`, `name`, `provider`, `path`, `registry`, and (when personas are in
  play) `persona`, the highest-precedence installed persona that brought the
  artifact into the desired set. Realizes [spec](../spec.md) FR-006. Created on
  the first successful run if absent.

## Writes

- `track.yaml` `items`: records installed/updated artifacts, removes
  entries for tracked artifacts no longer desired; only tracked artifacts are
  ever removed. With `--dry-run`, neither the environment nor `track.yaml` is
  touched. Realizes [spec](../spec.md) FR-005, FR-006, FR-009, FR-017. Atomic
  single-file write per the
  [configuration data contract](../../contracts/configuration.md#cross-file-write-semantics);
  no other configuration file is written.

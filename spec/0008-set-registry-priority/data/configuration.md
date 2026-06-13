# Data Model: Configuration — Set Registry Priority (registries.yaml)

**Spec**: [Set Registry Priority](../spec.md)

This feature rewrites a single registry's `priority` in the `items` array
of `registries.yaml`. The schema is owned by the
[configuration data contract](../../contracts/configuration.md#registriesyaml);
this document does not restate it.

## Reads

- `registries.yaml` `items` — to locate the entry by `name`, to check the
  new value against the other entries' `priority`, and to count how many
  registries exist.

## Owns

- Nothing. The `items` array is owned by
  [add registry](../../0001-add-registry/spec.md); this feature only adjusts an
  existing entry's `priority`.

## Writes

- `registries.yaml` `items[].priority` — the matched entry's `priority` is
  set to the given value; all other fields and all other registries are
  preserved. The value must be a unique, non-negative integer across all kinds
  (`0` only when no registry holds it). No write occurs when the value already
  equals the current priority, and the request is rejected while a single
  registry exists.

## Realizes

- `registries.yaml` `items[].priority` write → [spec](../spec.md) FR-003 (assign
  the new value), FR-004 (no-op when equal), FR-005 (transactional), FR-009
  (uniqueness), FR-010 (read failure), FR-011 (rejected while a single registry
  exists).

## Notes

Configuration is now split across files per the
[configuration data contract](../../contracts/configuration.md); file references
updated accordingly.

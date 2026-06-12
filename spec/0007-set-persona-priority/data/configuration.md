# Data Model: Configuration — Set Persona Priority (personas.yaml)

**Spec**: [Set Persona Priority](../spec.md)

This feature rewrites a single installed persona's `priority` in the `items`
array of `personas.yaml`. The schema is owned by the
[configuration data contract](../../contracts/configuration.md#personasyaml);
this document does not restate it.

## Reads

- `personas.yaml` `items` — to locate the entry by `name`, to check the
  new value against the other entries' `priority`, and to count how many
  personas are installed.

## Owns

- Nothing. The `items` array is owned by
  [select personas](../../0014-select-personas/spec.md); this feature only
  adjusts an existing entry's `priority`.

## Realizes

- `personas.yaml` `items[].priority` write → [spec](../spec.md) FR-003 (set the
  new value), FR-006 (transactional — unchanged until the write succeeds),
  FR-012 (read failure).

## Writes

- `personas.yaml` `items[].priority` — the matched entry's `priority` is
  set to the given value; all other fields and all other installed personas are
  preserved. The value must be a unique, non-negative integer (`0` only when no
  installed persona holds it). No write occurs when the value already equals the
  current priority, and the request is rejected while a single persona is
  installed. The override persists until the next
  [select personas](../../0014-select-personas/spec.md) `set persona`
  redeclaration resets positional priorities.

## Notes

Configuration is now split across files per the
[configuration data contract](../../contracts/configuration.md); file references
updated accordingly.

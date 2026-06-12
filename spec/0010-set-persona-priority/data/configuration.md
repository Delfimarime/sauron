# Data Model: Configuration — Sauron Settings (Persona Priority)

**Spec**: [Set Persona Priority](../spec.md)

Describes how the Set Persona Priority feature modifies the persisted configuration.

## Location & format

- **Path**: `~/.sauron/settings.yaml` (home directory resolved per platform).
- **Format**: a single YAML document with a `personas` array.

## Operation

- The `priority` field of the persona whose `name` matches the argument is set to the given value; all other fields and all other personas are preserved unchanged. Realizes [spec](../spec.md) FR-003.
- The persona's existing `priority` value is replaced with the new value. Realizes [spec](../spec.md) FR-005.
- Uniqueness is enforced over all personas — the new value must not be held by another persona. `0` is assignable only when free (see [import persona ADR-0002](../../0005-import-persona/architecture/ADR-0002-unified-priority-model.md)). Realizes [spec](../spec.md) FR-011.
- The request is rejected while only one persona exists; that persona keeps `0`. Realizes [spec](../spec.md) FR-009.

## Write semantics

- The whole document is loaded, the single field changed, and the document written back only after all validation passes. The file is left untouched on any failure. Realizes [spec](../spec.md) FR-006.
- Writes are atomic: serialize to a temporary file in `~/.sauron/`, then rename over `settings.yaml`.
- When the value equals the current priority, no write is performed (no-op). Realizes [spec](../spec.md) FR-004.

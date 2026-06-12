# Data Model: Configuration — Sauron Settings (Persona Priority)

**Spec**: [Set Persona Priority](../spec.md)

Describes how the Set Persona Priority feature modifies the persisted configuration.

## Location & format

- **Path**: `~/.sauron/settings.yaml` (home directory resolved per platform).
- **Format**: a single YAML document. Installed personas and their priorities live in the `installed` block, owned by [select personas](../../0014-select-personas/spec.md) (which defines its full schema). This feature reads that block and rewrites a single entry's `priority`.

## Operation

- The `priority` field of the installed persona whose `name` matches the argument is set to the given value; all other fields and all other installed personas in the `installed` block are preserved unchanged. Realizes [spec](../spec.md) FR-003.
- The installed persona's existing `priority` value is replaced with the new value. Realizes [spec](../spec.md) FR-005.
- Uniqueness is enforced over all installed personas — the new value must not be held by another installed persona. `0` is assignable only when free (see [priority model](../../AUTHORING.md#priority-model)). Realizes [spec](../spec.md) FR-011.
- The request is rejected while only one persona is installed; that persona keeps `0`. Realizes [spec](../spec.md) FR-009.

## Write semantics

- The whole document is loaded, the single `priority` field in the `installed` block changed, and the document written back only after all validation passes. The file is left untouched on any failure. Realizes [spec](../spec.md) FR-006.
- Writes are atomic: serialize to a temporary file in `~/.sauron/`, then rename over `settings.yaml`.
- When the value equals the current priority, no write is performed (no-op). Realizes [spec](../spec.md) FR-004.
- The override written here persists until the next [select personas](../../0014-select-personas/spec.md) `set persona` redeclaration rewrites the `installed` block and resets positional priorities.

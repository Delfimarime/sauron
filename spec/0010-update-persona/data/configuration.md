# Data Model: Configuration — Sauron Settings (Persona Update)

**Spec**: `../spec.md` (Update Persona)
**Status**: Draft

Describes how the Update Persona feature modifies the persisted configuration. The input file format is owned by `0007-import-persona` (`data/configuration.md` there); this feature consumes it unchanged.

## Location & format

- **Path**: `~/.sauron/settings.yaml` (home directory resolved per platform).
- **Format**: a single YAML document with a `personas` array.

## Operation

The persona entry whose `name` matches the definition file's `name` is updated in place:

| Field | Treatment |
|-------|-----------|
| `name` | Lookup key — never changed by update. Realizes FR-004. |
| `description` | Replaced with the file's value. Realizes FR-005. |
| `tags` | Replaced with the file's value (removed when the file omits it). Realizes FR-005. |
| `agents` | Replaced with the file's value. Realizes FR-005. |
| `skills` | Replaced with the file's value. Realizes FR-005. |
| `priority` | Preserved exactly — defined or undefined. Realizes FR-006. |

All other personas are preserved unchanged.

## Write semantics

- The whole document is loaded, the matching entry replaced, and the document written back only after all validation passes. The file is left untouched on any failure. Realizes FR-008.
- Writes are atomic: serialize to a temporary file in `~/.sauron/`, then rename over `settings.yaml`.

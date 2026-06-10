# Data Model: Configuration — Sauron Settings (Repository Priority)

**Spec**: [Set Repository Priority](../spec.md)

Describes how the Set Repository Priority feature modifies the persisted configuration.

## Location & format

- **Path**: `~/.sauron/settings.yaml` (home directory resolved per platform).
- **Format**: a single YAML document with a `repositories` array.

## Operation

- The `priority` field of the repository whose `name` matches the argument is set to the given value; all other fields and all other repositories are preserved unchanged. Realizes [spec](../spec.md) FR-003.
- The new value must be a positive integer not held by another repository (unique across all kinds). Realizes [spec](../spec.md) FR-007, FR-009.
- There is no undefined priority and no zero-anchor for repositories — every repository always has a defined, unique, positive priority.

## Write semantics

- The whole document is loaded, the single field changed, and the document written back only after all validation passes. The file is left untouched on any failure. Realizes [spec](../spec.md) FR-005.
- Writes are atomic: serialize to a temporary file in `~/.sauron/`, then rename over `settings.yaml`.
- When the value equals the current priority, no write is performed (no-op). Realizes [spec](../spec.md) FR-004.

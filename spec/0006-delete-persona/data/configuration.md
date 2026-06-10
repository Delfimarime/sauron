# Data Model: Configuration — Sauron Settings (Persona Deletion)

**Spec**: [Delete Persona](../spec.md)

Describes how the Delete Persona feature modifies the persisted configuration.

## Location & format

- **Path**: `~/.sauron/settings.yaml` (home directory resolved per platform).
- **Format**: a single YAML document with a `personas` array.

## Operation

- The entry whose `name` matches the argument is removed from `personas[]`; all other entries are preserved unchanged. Realizes [spec](../spec.md) FR-002.
- Identity is `name` (unique across personas), so at most one entry matches.
- Remaining personas keep their priorities exactly as they are — no renumbering. Realizes [spec](../spec.md) FR-009.
- Installed skills and agents are not referenced or modified — deletion touches only the persona definition. Realizes [spec](../spec.md) FR-003 (see [ADR-0001](../architecture/ADR-0001-unregister-keeps-installed-artifacts.md)).

## Write semantics

- The whole document is loaded, the matching entry removed, and the document written back only on success. The file is left untouched on any failure. Realizes [spec](../spec.md) FR-006.
- Writes are atomic: serialize to a temporary file in `~/.sauron/`, then rename over `settings.yaml`.
- When no entry matches, no write is performed (idempotent no-op). Realizes [spec](../spec.md) FR-005.

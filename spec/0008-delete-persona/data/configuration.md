# Data Model: Configuration — Sauron Settings (Persona Deletion)

**Spec**: `../spec.md` (Delete Persona)
**Status**: Draft

Describes how the Delete Persona feature modifies the persisted configuration.

## Location & format

- **Path**: `~/.sauron/settings.yaml` (home directory resolved per platform).
- **Format**: a single YAML document with a `personas` array.

## Operation

- The entry whose `name` matches the argument is removed from `personas[]`; all other entries are preserved unchanged. Realizes FR-002.
- Identity is `name` (unique across personas), so at most one entry matches.
- Remaining personas keep their priorities exactly as they are — no renumbering. Realizes FR-009.
- Installed skills and agents are not referenced or modified — deletion touches only the persona definition. Realizes FR-003 (see ADR-0001).

## Write semantics

- The whole document is loaded, the matching entry removed, and the document written back only on success. The file is left untouched on any failure. Realizes FR-005.
- Writes are atomic: serialize to a temporary file in `~/.sauron/`, then rename over `settings.yaml`.
- When no entry matches, no write is performed (idempotent no-op). Realizes FR-007.

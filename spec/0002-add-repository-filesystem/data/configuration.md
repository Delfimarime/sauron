# Data Model: Configuration — Sauron Settings

**Spec**: `../spec.md` (Add Filesystem Repository)
**Status**: Draft

Describes the persisted configuration that the Add Filesystem Repository feature reads and writes.

## Location & format

- **Path**: `~/.sauron/settings.json` (home directory resolved per platform).
- **Format**: a single JSON document.
- **Lifecycle**: created on first successful write if absent. Realizes FR-006.

## Schema

Top-level document:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `repositories` | array of Repository | Yes | Registered sources. Empty array when none are registered. |

Repository entry:

| Field | Type | Required | Constraints | Description |
|-------|------|----------|-------------|-------------|
| `kind` | string | Yes | enum: `filesystem` | Repository kind. Realizes FR-002, FR-010. |
| `name` | string | Yes | slug `^[a-z0-9]+(-[a-z0-9]+)*$`; unique across all kinds | Repository identity. Realizes FR-002, FR-005, FR-014, FR-021. |
| `priority` | integer | Yes | positive; unique across all kinds; lower = higher precedence | Repository ordering. Realizes FR-002, FR-015, FR-022. |
| `path` | string | Yes | absolute, symlink-resolved | Directory location. May repeat across entries. Realizes FR-002, FR-009. |

## Identity

A repository is identified by its `name`. `name` and `priority` are each unique across all repositories regardless of kind. `path` is not an identity key — two entries may share the same resolved path.

## Write semantics

- The whole document is loaded, the new entry appended, and the document written back only after all validation passes. The file is left untouched on any failure. Realizes FR-008.
- Writes are atomic: serialize to a temporary file in `~/.sauron/`, then rename over `settings.json`.

## Example

```json
{
  "repositories": [
    { "kind": "filesystem", "name": "team-skills",   "priority": 1, "path": "/home/user/team-skills" },
    { "kind": "filesystem", "name": "shared-agents", "priority": 2, "path": "/opt/shared/agents" }
  ]
}
```

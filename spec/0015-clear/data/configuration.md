# Data Model: Configuration — Sauron Settings (Clear)

**Spec**: `../spec.md` (Clear)
**Status**: Draft

Describes the data the Clear feature reads and updates. Clear operates on the tracking record only.

## Tracking record — `~/.sauron/track.yaml`

The installed-artifact record owned by `0011-sync` (which defines its full schema). Clear reads it and removes the entries it clears.

- **Path**: `~/.sauron/track.yaml` (home directory resolved per platform).
- **Format**: a single YAML document. A missing file is treated as nothing to clear.

## Operation

- Without `--persona`, every entry in `installed[]` is in scope (all targets). Realizes FR-002.
- With `--persona <name>`, only entries whose `persona` equals `<name>` are in scope, whether or not that persona is still defined in `settings.yaml`. Realizes FR-003.
- Each in-scope artifact is deleted from its `path` and its entry removed from `installed[]`. Realizes FR-004. With `--dry-run`, nothing is deleted or written. Realizes FR-006.
- `settings.yaml` is never read or written by clear.

## Write semantics

- Clear writes only `track.yaml`.
- Updates are atomic: serialize to a temporary file in `~/.sauron/`, then rename over `track.yaml`. The file is left untouched on `--dry-run` or when nothing is in scope.

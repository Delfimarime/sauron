# Contract: Command Line — Set Target

**Spec**: `../spec.md` (Set Target)
**Status**: Draft

Defines the command-line interface for setting the active target provider. This is the user-facing contract only.

## Synopsis

```
sauron set target <claude|zencoder> [--copy-only]
```

Command hierarchy: `sauron` (root) → `set` (group) → `target` (subcommand).

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `<target>` | Yes | The provider to make active: `claude` or `zencoder`. Realizes FR-003, FR-010, FR-011. |

## Flags

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--copy-only` | No | false | Copy installed artifacts to the new target without deleting them from the previous one. Realizes FR-006. |

## Output

- **Success**: a confirmation line naming the new target, followed by a sync-style report of what moved or was copied (grouped by `skills:`/`agents:`). A no-op states the target is already set.
- **Failure**: a single human-readable message on stderr; per-artifact failures are reported without stopping the run.

## Exit codes

| Code | Meaning | Realizes |
|------|---------|----------|
| `0` | Target set, or already set (no-op) | FR-007, FR-008, FR-012 |
| `2` | Usage error — missing or unsupported target value | FR-010, FR-011 |
| `1` | `settings.yaml`/`track.yaml` unreadable, or an artifact failed to move/copy | FR-013, FR-014 |

## Examples

```
# Switch from claude (default) to zencoder, moving artifacts
$ sauron set target zencoder
Target set to 'zencoder'; moved from 'claude':
skills:
+ design-oas3
agents:
+ software-engineer

# Copy instead of move (artifacts remain on the old target)
$ sauron set target zencoder --copy-only
Target set to 'zencoder'; copied from 'claude':
skills:
+ design-oas3
agents:
+ software-engineer

# Already set (no-op, exit 0)
$ sauron set target claude
Target is already 'claude'.

# Unsupported value (usage error, exit 2)
$ sauron set target codex
Error: target must be one of: claude, zencoder
```

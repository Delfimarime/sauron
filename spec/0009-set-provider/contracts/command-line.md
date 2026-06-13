# Contract: Command Line — Set Provider

Conventions: [CLI conventions](../../AUTHORING.md).

**Spec**: [Set Provider](../spec.md)

Defines the command-line interface for setting the active provider. This is the user-facing contract only.

## Synopsis

```
sauron set provider <claude|zencoder> [--copy-only]
```

Command hierarchy: `sauron` (root) → `set` (group) → `provider` (subcommand).

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `<provider>` | Yes | The provider to make active: `claude` or `zencoder`. Realizes [spec](../spec.md) FR-003, FR-009, FR-010. |

## Flags

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--copy-only` | No | false | Copy installed artifacts to the new provider without deleting them from the previous one. Realizes [spec](../spec.md) FR-014. |

## Output

- **Success**: a confirmation line naming the new provider, followed by a sync-style report of what moved or was copied (grouped by `skills:`/`agents:`). A no-op states the provider is already set.
- **Failure**: a single human-readable message on stderr; per-artifact failures are reported without stopping the run.

## Exit codes

Exit-status meanings are owned by the [CLI conventions](../../AUTHORING.md);
this table refines which conditions map to each code.

| Code | Meaning | Realizes |
|------|---------|----------|
| `0` | Provider set, or already set (no-op) | [spec](../spec.md) FR-006, FR-007, FR-011 |
| `2` | Usage error — missing or unsupported provider value | [spec](../spec.md) FR-009, FR-010 |
| `1` | Runtime error — the settings or the track file unreadable, or an artifact failed to move/copy | [spec](../spec.md) FR-012, FR-013 |

## Examples

```
# Switch from claude (default) to zencoder, moving artifacts
$ sauron set provider zencoder
Provider set to 'zencoder'; moved from 'claude':
skills:
+ design-oas3
agents:
+ software-engineer

# Copy instead of move (artifacts remain on the old provider)
$ sauron set provider zencoder --copy-only
Provider set to 'zencoder'; copied from 'claude':
skills:
+ design-oas3
agents:
+ software-engineer

# Already set (no-op, exit 0)
$ sauron set provider claude
Provider is already 'claude'.

# Unsupported value (usage error, exit 2)
$ sauron set provider codex
Error: provider must be one of: claude, zencoder
```

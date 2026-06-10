# Contract: Command Line — Sync

**Spec**: `../spec.md` (Sync)
**Status**: Draft

Defines the command-line interface for synchronizing skills and agents from the registered repositories to a target environment. This is the user-facing contract only.

## Synopsis

```
sauron sync [--persona <name>] [--target <claude|zencoder>] [--dry-run]
```

Command hierarchy: `sauron` (root) → `sync` (command).

## Arguments

None.

## Flags

| Flag | Required | Default | Values | Description |
|------|----------|---------|--------|-------------|
| `--persona` | No | — | persona name | Scope the sync to one persona's artifacts. Realizes FR-002, FR-012. |
| `--target` | No | zencoder | claude, zencoder | Provider environment to deliver to. Realizes FR-009, FR-013. |
| `--dry-run` | No | false | — | Print the plan without changing the environment or `track.yaml`. Realizes FR-008. |

## Output

- **Plan** (always printed; with `--dry-run` it is the only output): grouped by `skills:` and `agents:`, one artifact per line, `+` for additions/updates and `-` for removals.
- **Success**: after applying, a summary line with added/updated/removed counts; when nothing changes, an up-to-date message.
- **Failure**: human-readable messages on stderr; per-artifact and per-repository failures are reported without stopping the run.

## Exit codes

| Code | Meaning | Realizes |
|------|---------|----------|
| `0` | Synced, or already up to date (including `--dry-run`) | FR-007, FR-008, FR-010, FR-018 |
| `2` | Usage error — unsupported `--target` | FR-013 |
| `1` | Persona not found; configuration or tracking record unreadable; a repository unreachable; a desired artifact missing from every repository; or an artifact failed to install/remove | FR-012, FR-014, FR-015, FR-016, FR-017 |

## Examples

```
# Plan only
$ sauron sync --dry-run
skills:
+ design-oas3
- delete-mock
agents:
+ software-engineer

# Apply (default target zencoder)
$ sauron sync
skills:
+ design-oas3
- delete-mock
agents:
+ software-engineer
Synced target 'zencoder': 2 added, 1 removed.

# Sync one persona to claude
$ sauron sync --persona backend-developer --target claude
skills:
+ design-oas3
+ code-review
agents:
+ software-engineer
Synced target 'claude': 3 added, 0 removed.

# Up to date (exit 0)
$ sauron sync
Target 'zencoder' is up to date.

# Unknown persona (exit 1)
$ sauron sync --persona ghost
Error: no persona named 'ghost'

# Unsupported target (usage error, exit 2)
$ sauron sync --target codex
Error: --target must be one of: claude, zencoder
```

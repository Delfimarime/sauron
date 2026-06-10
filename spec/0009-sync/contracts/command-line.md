# Contract: Command Line — Sync

Conventions: [CLI contract](../../contracts/cli.md).

**Spec**: [Sync](../spec.md)

Defines the command-line interface for synchronizing artifacts from the registered repositories to the active target environment. This is the user-facing contract only. The target is the global setting managed by [set target](../../0012-set-target/spec.md); sync takes no target flag.

## Synopsis

```
sauron sync [--persona <name>] [--dry-run]
```

Command hierarchy: `sauron` (root) → `sync` (command).

## Arguments

None.

## Flags

| Flag | Required | Default | Values | Description |
|------|----------|---------|--------|-------------|
| `--persona` | No | — | persona name | Scope the sync to one persona's artifacts. Realizes [spec](../spec.md) FR-016, FR-010. |
| `--dry-run` | No | false | — | Print the plan without changing the environment or the track file. Realizes [spec](../spec.md) FR-017. |

## Output

- **Plan** (always printed; with `--dry-run` it is the only output): grouped by `skills:` and `agents:`, one artifact per line, `+` for additions/updates and `-` for removals.
- **Success**: after applying, a summary line with added/updated/removed counts; when nothing changes, an up-to-date message.
- **Failure**: human-readable messages on stderr; per-artifact and per-repository failures are reported without stopping the run.

## Exit codes

Exit-status meanings are owned by the [CLI contract](../../contracts/cli.md);
this table refines which conditions map to each code.

| Code | Meaning | Realizes |
|------|---------|----------|
| `0` | Synced, or already up to date (including `--dry-run`) | [spec](../spec.md) FR-006, FR-008, FR-015, FR-017 |
| `2` | Usage error — malformed flags (e.g. `--persona` with no value) | — |
| `1` | Persona not found; the settings or the track file unreadable; a repository unreachable; a desired artifact missing from every repository; or an artifact failed to install/remove | [spec](../spec.md) FR-010, FR-011, FR-012, FR-013, FR-014 |

## Examples

```
# Plan only
$ sauron sync --dry-run
skills:
+ design-oas3
- delete-mock
agents:
+ software-engineer

# Apply (active target, claude by default)
$ sauron sync
skills:
+ design-oas3
- delete-mock
agents:
+ software-engineer
Synced target 'claude': 2 added, 1 removed.

# Sync one persona
$ sauron sync --persona backend-developer
skills:
+ design-oas3
+ code-review
agents:
+ software-engineer
Synced target 'claude': 3 added, 0 removed.

# Up to date (exit 0)
$ sauron sync
Target 'claude' is up to date.

# Unknown persona (exit 1)
$ sauron sync --persona ghost
Error: no persona named 'ghost'
```

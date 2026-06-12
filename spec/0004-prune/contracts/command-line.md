# Contract: Command Line — Prune

Conventions: [CLI conventions](../../AUTHORING.md).

**Spec**: [Prune](../spec.md)

Defines the command-line interface for pruning orphaned artifacts. This is the user-facing contract only.

## Synopsis

```
sauron prune [skills|agents] [--dry-run]
```

Command hierarchy: `sauron` (root) → `prune` (command).

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `[type]` | No | `skills` or `agents`. Omitted = both. Realizes [spec](../spec.md) FR-002, FR-003, FR-008. |

## Flags

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--dry-run` | No | false | Print the plan without changing the environment or the track file. Realizes [spec](../spec.md) FR-012. |

## Output

- **Plan/summary**: grouped by `skills:` and `agents:`, one artifact per line with `-` (prune only removes). With `--dry-run` it is the plan only; otherwise it is what was removed, followed by a count line. When nothing is orphaned, a single message.
- **Failure**: a single human-readable message on stderr; per-artifact deletion failures are reported but do not stop the run.

## Exit codes

Exit-status meanings are owned by the [CLI conventions](../../AUTHORING.md);
this table refines which conditions map to each code.

| Code | Meaning | Realizes |
|------|---------|----------|
| `0` | Pruned, or nothing to prune (including `--dry-run`) | [spec](../spec.md) FR-005, FR-006, FR-012, FR-009 |
| `2` | Usage error — type other than `skills` or `agents` | [spec](../spec.md) FR-008 |
| `1` | Runtime error — the settings or the track file is unreadable, or one or more artifacts could not be deleted | [spec](../spec.md) FR-010, FR-011 |

## Examples

```
# Prune both skills and agents
$ sauron prune
skills:
- code-review
- release-notes
agents:
- triager
Pruned 3 artifacts (2 skills, 1 agent) from 1 unregistered registry.

# Prune only agents
$ sauron prune agents
agents:
- triager
Pruned 1 agent.

# Preview without deleting
$ sauron prune --dry-run
skills:
- code-review
- release-notes
agents:
- triager

# Nothing orphaned (exit 0)
$ sauron prune
Nothing to prune.

# Invalid type (usage error, exit 2)
$ sauron prune plugins
Error: prune accepts only 'skills' or 'agents'
```

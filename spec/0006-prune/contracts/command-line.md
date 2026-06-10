# Contract: Command Line — Prune Orphaned Skills and Agents

**Spec**: `../spec.md` (Prune Orphaned Skills and Agents)
**Status**: Draft

Defines the command-line interface for pruning orphaned skills and agents. This is the user-facing contract only.

## Synopsis

```
sauron prune [skills|agents] [--dry-run]
```

Command hierarchy: `sauron` (root) → `prune` (command).

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `[type]` | No | `skills` or `agents`. Omitted = both. Realizes FR-002, FR-003, FR-009. |

## Flags

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--dry-run` | No | false | Report what would be pruned without deleting anything or modifying `track.yaml`. Realizes FR-007. |

## Output

- **Success**: a summary on stdout naming the removed (or, with `--dry-run`, the would-be-removed) skills and agents, with a count. When nothing is orphaned, a single message.
- **Failure**: a single human-readable message on stderr; per-artifact deletion failures are reported but do not stop the run.

## Exit codes

| Code | Meaning | Realizes |
|------|---------|----------|
| `0` | Pruned, or nothing to prune (including `--dry-run`) | FR-005, FR-006, FR-007, FR-010 |
| `2` | Usage error — type other than `skills` or `agents` | FR-009 |
| `1` | `settings.yaml`/`track.yaml` unreadable, or one or more artifacts could not be deleted | FR-011, FR-012 |

## Examples

```
# Prune both skills and agents
$ sauron prune
Pruned 3 artifacts (2 skills, 1 agent) from 1 unregistered repository.

# Prune only agents
$ sauron prune agents
Pruned 1 agent.

# Preview without deleting
$ sauron prune --dry-run
Would prune 3 artifacts (2 skills, 1 agent):
  skill  code-review     (from removed repo 'team-deploy')
  skill  release-notes   (from removed repo 'team-deploy')
  agent  triager         (from removed repo 'old-http')

# Nothing orphaned (exit 0)
$ sauron prune
Nothing to prune.

# Invalid type (usage error, exit 2)
$ sauron prune plugins
Error: prune accepts only 'skills' or 'agents'
```

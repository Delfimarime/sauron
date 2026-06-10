# Contract: Command Line — Clear

**Spec**: `../spec.md` (Clear)
**Status**: Draft

Defines the command-line interface for erasing the skills and agents Sauron manages. This is the user-facing contract only.

## Synopsis

```
sauron clear [--persona <name>] [--dry-run]
```

Command hierarchy: `sauron` (root) → `clear` (command).

## Arguments

None.

## Flags

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--persona` | No | — | Limit clearing to artifacts recorded with this persona. Realizes FR-003, FR-008. |
| `--dry-run` | No | false | Report what would be removed without deleting anything or modifying `track.yaml`. Realizes FR-006. |

## Output

- **Plan/summary**: grouped by `skills:` and `agents:`, one artifact per line with `-` (everything in scope is removed). With `--dry-run` it is the plan only; otherwise it is what was removed.
- **Failure**: human-readable messages on stderr; per-artifact failures are reported without stopping the run.

## Exit codes

| Code | Meaning | Realizes |
|------|---------|----------|
| `0` | Cleared, or nothing to clear (including `--dry-run`) | FR-004, FR-005, FR-006, FR-009 |
| `2` | Usage error — `--persona` without a value | FR-008 |
| `1` | `track.yaml` unreadable, or one or more artifacts could not be deleted | FR-010, FR-011 |

## Examples

```
# Clear everything
$ sauron clear
skills:
- design-oas3
- code-review
agents:
- software-engineer
Cleared 3 artifacts (2 skills, 1 agent).

# Preview only
$ sauron clear --dry-run
skills:
- design-oas3
agents:
- software-engineer

# Clear one persona's artifacts
$ sauron clear --persona qa-engineer
skills:
- test-plan
Cleared 1 skill.

# Nothing to clear (exit 0)
$ sauron clear
Nothing to clear.
```

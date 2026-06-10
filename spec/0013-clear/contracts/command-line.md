# Contract: Command Line — Clear

Conventions: [CLI contract](../../contracts/cli.md).

**Spec**: [Clear](../spec.md)

Defines the command-line interface for erasing the artifacts Sauron manages. This is the user-facing contract only.

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
| `--persona` | No | — | Limit clearing to artifacts recorded with this persona. Realizes [spec](../spec.md) FR-010, FR-006. |
| `--dry-run` | No | false | Print the plan without changing the environment or the track file. Realizes [spec](../spec.md) FR-011. |

## Output

- **Plan/summary**: grouped by `skills:` and `agents:`, one artifact per line with `-` (everything in scope is removed). With `--dry-run` it is the plan only; otherwise it is what was removed.
- **Failure**: human-readable messages on stderr; per-artifact failures are reported without stopping the run.

## Exit codes

Exit-status meanings are owned by the [CLI contract](../../contracts/cli.md);
this table refines which conditions map to each code.

| Code | Meaning | Realizes |
|------|---------|----------|
| `0` | Cleared, or nothing to clear (including `--dry-run`) | [spec](../spec.md) FR-003, FR-004, FR-011, FR-007 |
| `2` | Usage error — `--persona` without a value | [spec](../spec.md) FR-006 |
| `1` | Runtime error — the track file unreadable, or one or more artifacts could not be deleted | [spec](../spec.md) FR-008, FR-009 |

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

# Contract: Command Line — Delete Artifacts

Conventions: [CLI conventions](../../AUTHORING.md).

**Spec**: [Delete Artifacts](../spec.md)

Defines the command-line interface for erasing the artifacts Sauron manages. This is the user-facing contract only.

## Synopsis

```
sauron delete (artifacts|skills|agents) [--persona <name>] [--dry-run]
```

Command hierarchy: `sauron` (root) → `delete` (verb) → `(artifacts|skills|agents)` (noun).

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `(artifacts\|skills\|agents)` | Yes | The artifact scope: `artifacts` (both), `skills`, or `agents`. Realizes [spec](../spec.md) FR-002, FR-012, FR-013. |

## Flags

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--persona` | No | — | Limit deletion to artifacts recorded with this persona. Realizes [spec](../spec.md) FR-010, FR-006. |
| `--dry-run` | No | false | Print the plan without changing the environment or the track file. Realizes [spec](../spec.md) FR-011. |

## Output

- **Plan/summary**: grouped by `skills:` and `agents:`, one artifact per line with `-` (everything in scope is removed). With `--dry-run` it is the plan only; otherwise it is what was removed.
- **Failure**: human-readable messages on stderr; per-artifact failures are reported without stopping the run.

## Exit codes

Exit-status meanings are owned by the [CLI conventions](../../AUTHORING.md);
this table refines which conditions map to each code.

| Code | Meaning | Realizes |
|------|---------|----------|
| `0` | Deleted, or nothing to delete (including `--dry-run`) | [spec](../spec.md) FR-003, FR-004, FR-011, FR-007 |
| `2` | Usage error — missing noun, a noun other than `artifacts`/`skills`/`agents`, or `--persona` without a value | [spec](../spec.md) FR-013, FR-006 |
| `1` | Runtime error — the track file unreadable, or one or more artifacts could not be deleted | [spec](../spec.md) FR-008, FR-009 |

## Examples

```
# Delete all tracked artifacts (skills + agents)
$ sauron delete artifacts
skills:
- design-oas3
- code-review
agents:
- software-engineer
Deleted 3 artifacts (2 skills, 1 agent).

# Delete only agents
$ sauron delete agents
agents:
- software-engineer
Deleted 1 agent.

# Preview only
$ sauron delete artifacts --dry-run
skills:
- design-oas3
agents:
- software-engineer

# Delete one persona's artifacts
$ sauron delete artifacts --persona qa-engineer
skills:
- test-plan
Deleted 1 skill.

# Nothing to delete (exit 0)
$ sauron delete artifacts
Nothing to delete.

# Missing or invalid noun (usage error, exit 2)
$ sauron delete plugins
Error: delete accepts only 'artifacts', 'skills', or 'agents'
```

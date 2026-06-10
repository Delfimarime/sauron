# Contract: Command Line — Set Persona Priority

Conventions: [CLI contract](../../contracts/cli.md).

**Spec**: [Set Persona Priority](../spec.md)

Defines the command-line interface for changing a persona's priority. This is the user-facing contract only. The `set priority` group also covers `repository` ([set repository priority](../../0011-set-repository-priority/spec.md)).

## Synopsis

```
sauron set priority persona <name> <value>
```

Command hierarchy: `sauron` (root) → `set` (group) → `priority` (group) → `persona` (subcommand). The sibling `sauron set priority repository` is covered by [set repository priority](../../0011-set-repository-priority/spec.md).

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `<name>` | Yes | Name of the persona. Realizes [spec](../spec.md) FR-002, FR-007, FR-010. |
| `<value>` | Yes | New priority — a non-negative integer not used by another persona. Realizes [spec](../spec.md) FR-002, FR-008, FR-011. |

## Flags

None.

## Output

- **Success**: a single confirmation line on stdout; a no-op states the priority was already set.
- **Failure**: a single human-readable message on stderr.

## Exit codes

Exit-status meanings are owned by the [CLI contract](../../contracts/cli.md);
this table refines which conditions map to each code.

| Code | Meaning | Realizes |
|------|---------|----------|
| `0` | Priority set (including the no-op case) | [spec](../spec.md) FR-003, FR-004, FR-005 |
| `2` | Usage error — missing `<name>`/`<value>`, or `<value>` not a non-negative integer | [spec](../spec.md) FR-007, FR-008 |
| `1` | Validation error — only one persona exists, persona not found, priority taken, or the settings unreadable | [spec](../spec.md) FR-009, FR-010, FR-011, FR-012 |

## Examples

```
# Success
$ sauron set priority persona qa-engineer 2
Set priority of persona 'qa-engineer' to 2

# Same value (no-op, exit 0)
$ sauron set priority persona qa-engineer 2
Priority of persona 'qa-engineer' is already 2

# Only one persona exists (validation error, exit 1)
$ sauron set priority persona backend-developer 5
Error: cannot change priority while only one persona exists; it keeps priority 0

# Priority taken (validation error, exit 1)
$ sauron set priority persona designer 0
Error: priority 0 is already in use

# Persona not found (validation error, exit 1)
$ sauron set priority persona ghost 3
Error: no persona named 'ghost'

# Invalid value (usage error, exit 2)
$ sauron set priority persona designer high
Error: priority must be a non-negative integer
```

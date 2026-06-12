# Contract: Command Line — Set Registry Priority

Conventions: [CLI conventions](../../AUTHORING.md).

**Spec**: [Set Registry Priority](../spec.md)

Defines the command-line interface for changing a registry's priority. This is the user-facing contract only.

## Synopsis

```
sauron set priority registry <name> <value>
```

Command hierarchy: `sauron` (root) → `set` (group) → `priority` (group) → `registry` (subcommand). The sibling `sauron set priority persona` is covered by [set persona priority](../../0007-set-persona-priority/spec.md).

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `<name>` | Yes | Name of the registry. Realizes [spec](../spec.md) FR-002, FR-006, FR-008. |
| `<value>` | Yes | New priority — a non-negative integer not used by another registry (`0` is assignable only when no registry holds it). See the [unified priority model](../../AUTHORING.md#priority-model). Realizes [spec](../spec.md) FR-002, FR-007, FR-009. |

## Flags

None.

## Output

- **Success**: a single confirmation line on stdout; a no-op states the priority was already set.
- **Failure**: a single human-readable message on stderr.

## Exit codes

Exit-status meanings are owned by the [CLI conventions](../../AUTHORING.md);
this table refines which conditions map to each code.

| Code | Meaning | Realizes |
|------|---------|----------|
| `0` | Priority set (including the no-op case) | [spec](../spec.md) FR-003, FR-004 |
| `2` | Usage error — missing `<name>`/`<value>`, or `<value>` not a non-negative integer | [spec](../spec.md) FR-006, FR-007 |
| `1` | Runtime error — registry not found, priority taken, only a single registry exists, or the settings unreadable | [spec](../spec.md) FR-008, FR-009, FR-010, FR-011 |

## Examples

```
# Success
$ sauron set priority registry team-http 5
Set priority of registry 'team-http' to 5

# Same value (no-op, exit 0)
$ sauron set priority registry team-http 5
Priority of registry 'team-http' is already 5

# Priority taken (validation error, exit 1)
$ sauron set priority registry team-http 1
Error: priority 1 is already in use

# Registry not found (runtime error, exit 1)
$ sauron set priority registry ghost 3
Error: no registry named 'ghost'

# Single registry (runtime error, exit 1)
$ sauron set priority registry only-repo 2
Error: priority cannot be changed while a single registry exists

# Invalid value (usage error, exit 2)
$ sauron set priority registry team-http -1
Error: priority must be a non-negative integer
```

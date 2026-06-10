# Contract: Command Line — Set Repository Priority

**Spec**: `../spec.md` (Set Repository Priority)
**Status**: Draft

Defines the command-line interface for changing a repository's priority. This is the user-facing contract only.

## Synopsis

```
sauron set priority repository <name> <value>
```

Command hierarchy: `sauron` (root) → `set` (group) → `priority` (group) → `repository` (subcommand). The sibling `sauron set priority persona` is covered by `0012-set-persona-priority`.

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `<name>` | Yes | Name of the repository. Realizes FR-002, FR-006, FR-008. |
| `<value>` | Yes | New priority — a positive integer (`1` or greater) not used by another repository. Realizes FR-002, FR-007, FR-009. |

## Flags

None.

## Output

- **Success**: a single confirmation line on stdout; a no-op states the priority was already set.
- **Failure**: a single human-readable message on stderr.

## Exit codes

| Code | Meaning | Realizes |
|------|---------|----------|
| `0` | Priority set (including the no-op case) | FR-003, FR-004 |
| `2` | Usage error — missing `<name>`/`<value>`, or `<value>` not a positive integer | FR-006, FR-007 |
| `1` | Validation error — repository not found, priority taken, or configuration unreadable | FR-008, FR-009, FR-010 |

## Examples

```
# Success
$ sauron set priority repository team-http 5
Set priority of repository 'team-http' to 5

# Same value (no-op, exit 0)
$ sauron set priority repository team-http 5
Priority of repository 'team-http' is already 5

# Priority taken (validation error, exit 1)
$ sauron set priority repository team-http 1
Error: priority 1 is already in use

# Repository not found (validation error, exit 1)
$ sauron set priority repository ghost 3
Error: no repository named 'ghost'

# Invalid value (usage error, exit 2)
$ sauron set priority repository team-http 0
Error: priority must be a positive integer (1 or greater)
```

# Contract: Command Line — Delete Repository

**Spec**: `../spec.md` (Delete Repository)
**Status**: Draft

Defines the command-line interface for deleting a registered repository. This is the user-facing contract only. Deletion is idempotent and spans all repository kinds.

## Synopsis

```
sauron delete repository <name>
```

Command hierarchy: `sauron` (root) → `delete` (group) → `repository` (subcommand).

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `<name>` | Yes | Name of the repository to delete. Realizes FR-002, FR-006. |

## Flags

None.

## Output

- **Success**: a single confirmation line on stdout; when no repository matched, a line noting nothing was deleted.
- **Failure**: a single human-readable message on stderr.

## Exit codes

| Code | Meaning | Realizes |
|------|---------|----------|
| `0` | Repository deleted, or no repository with that name existed (idempotent) | FR-004, FR-007 |
| `2` | Usage error — missing `<name>` | FR-006 |
| `1` | Configuration error — configuration cannot be read or parsed | FR-008 |

## Examples

```
# Success
$ sauron delete repository team-http
Deleted repository 'team-http'

# Not registered (idempotent, exit 0)
$ sauron delete repository ghost
No repository named 'ghost'; nothing to delete

# Missing name (usage error, exit 2)
$ sauron delete repository
Error: a repository name is required
```

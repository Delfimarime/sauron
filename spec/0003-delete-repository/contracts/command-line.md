# Contract: Command Line — Delete Repository

Conventions: [CLI contract](../../contracts/cli.md).

**Spec**: [Delete Repository](../spec.md)

Defines the command-line interface for deleting a registered repository. This is the user-facing contract only. Deletion is idempotent and spans all repository kinds.

## Synopsis

```
sauron delete repository <name>
```

Command hierarchy: `sauron` (root) → `delete` (group) → `repository` (subcommand).

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `<name>` | Yes | Name of the repository to delete. Realizes [spec](../spec.md) FR-002, FR-007. |

## Flags

None.

## Output

- **Success**: a single confirmation line on stdout; when no repository matched, a line noting nothing was deleted.
- **Failure**: a single human-readable message on stderr.

## Exit codes

Exit-status meanings are owned by the [CLI contract](../../contracts/cli.md);
this table refines which conditions map to each code.

| Code | Meaning | Realizes |
|------|---------|----------|
| `0` | Repository deleted, or no repository with that name existed (idempotent) | [spec](../spec.md) FR-004, FR-005 |
| `2` | Usage error — missing `<name>` | [spec](../spec.md) FR-007 |
| `1` | Runtime error — the settings cannot be read or parsed | [spec](../spec.md) FR-008 |

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

# Contract: Command Line — Delete Persona

Conventions: [CLI contract](../../contracts/cli.md).

**Spec**: [Delete Persona](../spec.md)

Defines the command-line interface for deleting a registered persona. This is the user-facing contract only. Deletion is idempotent.

## Synopsis

```
sauron delete persona <name>
```

Command hierarchy: `sauron` (root) → `delete` (group) → `persona` (subcommand).

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `<name>` | Yes | Name of the persona to delete. Realizes [spec](../spec.md) FR-002, FR-007. |

## Flags

None.

## Output

- **Success**: a single confirmation line on stdout; when no persona matched, a line noting nothing was deleted.
- **Failure**: a single human-readable message on stderr.

## Exit codes

Exit-status meanings are owned by the [CLI contract](../../contracts/cli.md);
this table refines which conditions map to each code.

| Code | Meaning | Realizes |
|------|---------|----------|
| `0` | Persona deleted, or no persona with that name existed (idempotent) | [spec](../spec.md) FR-004, FR-005 |
| `2` | Usage error — missing `<name>` | [spec](../spec.md) FR-007 |
| `1` | Runtime error — the settings cannot be read or parsed | [spec](../spec.md) FR-008 |

## Examples

```
# Success
$ sauron delete persona qa-engineer
Deleted persona 'qa-engineer'

# Not registered (idempotent, exit 0)
$ sauron delete persona ghost
No persona named 'ghost'; nothing to delete

# Missing name (usage error, exit 2)
$ sauron delete persona
Error: a persona name is required
```

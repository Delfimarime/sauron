# Contract: Command Line — Delete Persona

**Spec**: `../spec.md` (Delete Persona)
**Status**: Draft

Defines the command-line interface for deleting a registered persona. This is the user-facing contract only. Deletion is idempotent.

## Synopsis

```
sauron delete persona <name>
```

Command hierarchy: `sauron` (root) → `delete` (group) → `persona` (subcommand).

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `<name>` | Yes | Name of the persona to delete. Realizes FR-002, FR-006. |

## Flags

None.

## Output

- **Success**: a single confirmation line on stdout; when no persona matched, a line noting nothing was deleted.
- **Failure**: a single human-readable message on stderr.

## Exit codes

| Code | Meaning | Realizes |
|------|---------|----------|
| `0` | Persona deleted, or no persona with that name existed (idempotent) | FR-004, FR-007 |
| `2` | Usage error — missing `<name>` | FR-006 |
| `1` | Configuration error — configuration cannot be read or parsed | FR-008 |

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

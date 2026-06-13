# Contract: Command Line — Delete Registry

Conventions: [CLI conventions](../../AUTHORING.md).

**Spec**: [Delete Registry](../spec.md)

Defines the command-line interface for deleting a registered registry. This is the user-facing contract only. Deletion is idempotent and spans all registry kinds.

## Synopsis

```
sauron delete registry <name>
```

Command hierarchy: `sauron` (root) → `delete` (group) → `registry` (subcommand).

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `<name>` | Yes | Name of the registry to delete. Realizes [spec](../spec.md) FR-002, FR-007. |

## Flags

None.

## Output

- **Success**: a single confirmation line on stdout; when no registry matched, a line noting nothing was deleted.
- **Failure**: a single human-readable message on stderr.

## Exit codes

Exit-status meanings are owned by the [CLI conventions](../../AUTHORING.md);
this table refines which conditions map to each code.

| Code | Meaning | Realizes |
|------|---------|----------|
| `0` | Registry deleted, or no registry with that name existed (idempotent) | [spec](../spec.md) FR-004, FR-005 |
| `2` | Usage error — missing `<name>` | [spec](../spec.md) FR-007 |
| `1` | Runtime error — `registries.yaml` cannot be read or parsed | [spec](../spec.md) FR-008 |

## Examples

```
# Success
$ sauron delete registry team-http
Deleted registry 'team-http'

# Not registered (idempotent, exit 0)
$ sauron delete registry ghost
No registry named 'ghost'; nothing to delete

# Missing name (usage error, exit 2)
$ sauron delete registry
Error: a registry name is required
```

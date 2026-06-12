# Contract: Command Line — Select Personas

Conventions: [CLI conventions](../../AUTHORING.md).

**Spec**: [Select Personas](../spec.md)

Defines the command-line interface for declaring and clearing the set of
installed personas. This is the user-facing contract only. This feature owns two
commands: `set persona` and `unset persona`.

## Synopsis

```
sauron set persona <name>...
sauron unset persona [<name>...]
```

Command hierarchy: `sauron` (root) → `set` (group) → `persona` (subcommand), and
`sauron` (root) → `unset` (group) → `persona` (subcommand). `set persona`
declares the exact installed set; `unset persona` uninstalls. Adjusting a
persona's priority after installation is the sibling command
[set priority persona](../../0007-set-persona-priority/spec.md), not this one.

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `set persona <name>...` | Yes (one or more) | The exact set of personas to install. Order is significant: it assigns priority positionally (first listed is `0`). Every name must be offered by a live fetch from the [backend](../../0012-backend/spec.md); the install stores each persona's full definition from that fetch. Realizes [spec](../spec.md) FR-003, FR-004, FR-005, FR-012, FR-013, FR-017, FR-018. |
| `unset persona [<name>...]` | No | Personas to uninstall. With no name, every installed persona is uninstalled. Contacts no backend. Realizes [spec](../spec.md) FR-008, FR-009. |

## Flags

None.

## Output

- **`set persona` success**: on stdout, the full resulting installed set with
  each persona's priority, and separately the personas that were dropped.
  Realizes [spec](../spec.md) FR-006.
- **`unset persona` success**: on stdout, the personas that were uninstalled; an
  idempotent no-op states that nothing was deleted. Realizes
  [spec](../spec.md) FR-008, FR-009, FR-014.
- **Failure**: a single human-readable message on stderr. Realizes
  [spec](../spec.md) FR-016.

## Exit codes

Exit-status meanings are owned by the [CLI conventions](../../AUTHORING.md);
this table refines which conditions map to each code.

| Code | Meaning | Realizes |
|------|---------|----------|
| `0` | Installed set declared, personas uninstalled, or an idempotent `unset` no-op | [spec](../spec.md) FR-004, FR-008, FR-009, FR-014 |
| `2` | Usage error — `set persona` given no name | [spec](../spec.md) FR-012 |
| `1` | Runtime error — a name given to `set persona` not offered by the live backend fetch, the backend unreachable while installing, or the installed personas unreadable | [spec](../spec.md) FR-013, FR-018, FR-015 |

## Examples

```
# Install an exact set; priority follows argument order (success, exit 0)
$ sauron set persona platform security data
Installed personas:
  0  platform
  1  security
  2  data
Dropped: frontend

# Re-declare a smaller set; priorities reset, others dropped (success, exit 0)
$ sauron set persona security platform
Installed personas:
  0  security
  1  platform
Dropped: data

# Name not offered by the backend (runtime error, exit 1) — config unchanged
$ sauron set persona platform ghost
Error: persona 'ghost' is not offered by the backend

# Backend unreachable while installing (runtime error, exit 1) — config unchanged
$ sauron set persona platform security
Error: backend is unreachable; installed personas unchanged

# No name given (usage error, exit 2)
$ sauron set persona
Error: at least one persona name is required; use 'sauron unset persona' to clear the installed set

# Uninstall named personas (success, exit 0); no backend contacted
$ sauron unset persona data
Uninstalled persona 'data'

# Uninstall all (success, exit 0)
$ sauron unset persona
Uninstalled personas: security, platform

# Uninstall a persona that is not installed (idempotent no-op, exit 0)
$ sauron unset persona data
Persona 'data' is not installed; nothing was deleted
```

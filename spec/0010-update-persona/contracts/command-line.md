# Contract: Command Line — Update Persona

**Spec**: `../spec.md` (Update Persona)
**Status**: Draft

Defines the command-line interface for updating an existing persona from a YAML definition file. This is the user-facing contract only. The file format is the same as import's; the `name` in the file is the key.

## Synopsis

```
sauron update persona <path>
```

Command hierarchy: `sauron` (root) → `update` (group) → `persona` (subcommand).

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `<path>` | Yes | Path to the persona definition file (YAML, same format as import). Realizes FR-002, FR-009, FR-010. |

## Flags

None.

## Output

- **Success**: a single confirmation line on stdout naming the updated persona.
- **Failure**: a single human-readable message on stderr. No partial output, no stack traces.

## Exit codes

| Code | Meaning | Realizes |
|------|---------|----------|
| `0` | Persona updated | FR-005, FR-007 |
| `2` | Usage error — missing `<path>` | FR-009 |
| `1` | Validation error — definition unreadable/malformed, invalid or missing name, missing description, no artifacts, persona not found, or configuration unreadable | FR-010, FR-011, FR-012, FR-013, FR-014, FR-015 |

## Examples

```
# Success
$ sauron update persona ./backend-developer.yaml
Updated persona 'backend-developer'

# Persona not found (validation error, exit 1)
$ sauron update persona ./newcomer.yaml
Error: no persona named 'newcomer'; use 'sauron import persona' to create it

# No artifacts (validation error, exit 1)
$ sauron update persona ./empty.yaml
Error: a persona needs at least one skill or agent

# Missing path (usage error, exit 2)
$ sauron update persona
Error: a path to a persona definition file is required
```

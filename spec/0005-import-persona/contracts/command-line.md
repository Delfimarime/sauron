# Contract: Command Line — Import Persona

Conventions: [CLI contract](../../contracts/cli.md).

**Spec**: [Import Persona](../spec.md)

Defines the command-line interface for importing a persona from a YAML definition file. This is the user-facing contract only.

## Synopsis

```
sauron import persona [--priority <n>] <path>
```

Command hierarchy: `sauron` (root) → `import` (group) → `persona` (subcommand).

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `<path>` | Yes | Path to the persona definition file (YAML). Realizes [spec](../spec.md) FR-004, FR-011, FR-012. |

## Flags

| Flag | Required | Default | Values | Description |
|------|----------|---------|--------|-------------|
| `--priority` | No | — | integer ≥ 1 | Persona priority; allowed only when at least one persona already exists, and must be unused. Omitted = undefined priority. The first persona always takes `0`. Realizes [spec](../spec.md) FR-006, FR-020, FR-007, FR-017, FR-018, FR-019. |

## Output

- **Success**: a single confirmation line on stdout naming the imported persona (and its priority when defined).
- **Failure**: a single human-readable message on stderr. No partial output, no stack traces.

## Exit codes

Exit-status meanings are owned by the [CLI contract](../../contracts/cli.md);
this table refines which conditions map to each code.

| Code | Meaning | Realizes |
|------|---------|----------|
| `0` | Persona imported | [spec](../spec.md) FR-008, FR-009 |
| `2` | Usage error — missing `<path>`, `--priority` on the first import, or `--priority` not an integer ≥ 1 | [spec](../spec.md) FR-011, FR-017, FR-018 |
| `1` | Validation error — definition unreadable/malformed, invalid or missing name, missing description, no artifacts, duplicate name, or priority taken | [spec](../spec.md) FR-012, FR-013, FR-014, FR-015, FR-016, FR-019 |

## Examples

```
# First persona (forced priority 0)
$ sauron import persona ./backend-developer.yaml
Imported persona 'backend-developer' (priority 0)

# Subsequent persona with explicit priority
$ sauron import persona --priority 1 ./qa-engineer.yaml
Imported persona 'qa-engineer' (priority 1)

# Subsequent persona without priority (undefined)
$ sauron import persona ./designer.yaml
Imported persona 'designer'

# Priority on first import (usage error, exit 2)
$ sauron import persona --priority 1 ./backend-developer.yaml
Error: the first persona always takes priority 0; omit --priority

# Duplicate name (validation error, exit 1)
$ sauron import persona ./backend-developer.yaml
Error: a persona named 'backend-developer' already exists

# Priority taken (validation error, exit 1)
$ sauron import persona --priority 1 ./platform.yaml
Error: priority 1 is already in use

# No artifacts (validation error, exit 1)
$ sauron import persona ./empty.yaml
Error: a persona needs at least one skill or agent
```

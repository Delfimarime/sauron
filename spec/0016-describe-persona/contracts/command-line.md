# Contract: Command Line — Describe Persona

Conventions: [CLI conventions](../../AUTHORING.md).

**Spec**: [Describe Persona](../spec.md)

Defines the command-line interface for describing a single persona from the
catalog. This is the user-facing contract only. Describing is read-only and
works offline against the local catalog; it never writes the settings or the
track file.

## Synopsis

```
sauron describe persona <name> [--fields <list>]
```

Command hierarchy: `sauron` (root) → `describe` (group) → `persona`
(subcommand).

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `<name>` | Yes | Name of the catalog persona to describe; resolved whether it is installed or available. Realizes [spec](../spec.md) FR-001, FR-002, FR-006. |

## Flags

| Flag | Required | Default | Values | Description |
|------|----------|---------|--------|-------------|
| `--fields` | No | all fields | `description`, `tags`, `installed`, `priority`, `skills`, `agents`, `registry`, `last-updated`, `last-synced` | Comma-separated fields, in order; `name` is always present and first. Realizes [spec](../spec.md) FR-011, FR-007. |

## Output

- **Success**: one field per line on stdout, formatted `field: value`, with the
  identity field `name` first. `--fields` selects and orders the printed fields
  (`name` always first); when omitted the full field set is printed. `skills`
  and `agents` are shown as integer counts. An absent value is empty after the
  colon. `priority`, `last-updated`, and `last-synced` are empty for a
  not-installed persona, since those values exist only for installed personas.
- **Failure**: a single human-readable message on stderr.

## Exit codes

Exit-status meanings are owned by the [CLI conventions](../../AUTHORING.md);
this table refines which conditions map to each code.

| Code | Meaning | Realizes |
|------|---------|----------|
| `0` | Described — the persona was resolved and printed | [spec](../spec.md) FR-002, FR-003, FR-004 |
| `2` | Usage error — `<name>` missing, or `--fields` without a value or naming an unknown field | [spec](../spec.md) FR-006, FR-007 |
| `1` | Runtime error — the persona was not found, or the settings or track file cannot be read or parsed | [spec](../spec.md) FR-005, FR-008, FR-009 |

## Examples

```
# Describe an installed persona (full field set)
$ sauron describe persona backend-developer
name: backend-developer
description: Backend service engineer for Go projects
tags: backend, golang
installed: yes
priority: 0
skills: 2
agents: 1
registry: team-registry
last-updated: 2026-06-11T18:00:00Z
last-synced: 2026-06-12T09:30:00Z

# Describe an available (not-installed) persona — priority and timestamps empty
$ sauron describe persona designer
name: designer
description: Product and UI design assistant
tags: design
installed: no
priority:
skills: 1
agents: 1
registry: team-registry
last-updated:
last-synced:

# Select and order a subset of fields
$ sauron describe persona backend-developer --fields priority,tags
name: backend-developer
priority: 0
tags: backend, golang

# Persona not found (runtime error, exit 1)
$ sauron describe persona ghost
Error: persona "ghost" not found in the catalog

# Missing name argument (usage error, exit 2)
$ sauron describe persona
Error: persona name is required

# Unknown field (usage error, exit 2)
$ sauron describe persona backend-developer --fields foo
Error: --fields must be a subset of: description, tags, installed, priority, skills, agents, registry, last-updated, last-synced
```

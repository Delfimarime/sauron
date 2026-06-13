# Contract: Command Line — Describe Provider

Conventions: [CLI conventions](../../AUTHORING.md).

**Spec**: [Describe Provider](../spec.md)

Defines the command-line interface for describing the active provider. This is
the user-facing contract only. Describing is read-only and offline: it reads the
active provider from the settings and contacts no external resource. The
provider is a singleton, so the command takes no name.

## Synopsis

```
sauron describe provider [--fields <list>]
```

Command hierarchy: `sauron` (root) → `describe` (group) → `provider`
(subcommand).

## Arguments

This command takes no positional arguments; supplying one is a usage error.
Realizes [spec](../spec.md) FR-001, FR-007.

## Flags

| Flag | Required | Default | Values | Description |
|------|----------|---------|--------|-------------|
| `--fields` | No | all fields | comma-separated `name`, `skills-location`, `agents-location` | Fields to display, in order; `name` is always present and first. Realizes [spec](../spec.md) FR-011, FR-008. |

## Output

- **Success**: the provider's fields on stdout, one per line as `field: value`,
  with `name` first. A field with no value is shown with an empty value.
  Realizes [spec](../spec.md) FR-003, FR-004, FR-006.
- **Failure**: a single human-readable message on stderr. Realizes
  [spec](../spec.md) FR-009, FR-010.

## Exit codes

Exit-status meanings are owned by the [CLI conventions](../../AUTHORING.md);
this table refines which conditions map to each code.

| Code | Meaning | Realizes |
|------|---------|----------|
| `0` | The active provider was described | [spec](../spec.md) FR-003, FR-005 |
| `2` | Usage error — a name argument was supplied, or an unknown `--fields` value | [spec](../spec.md) FR-007, FR-008 |
| `1` | Runtime error — the settings cannot be read or parsed | [spec](../spec.md) FR-009 |

## Examples

```
# Describe the active provider (full field set, name first)
$ sauron describe provider
name: claude
skills-location: ~/.claude/skills
agents-location: ~/.claude/agents

# A subset of fields, in the given order (name always first)
$ sauron describe provider --fields agents-location
name: claude
agents-location: ~/.claude/agents

# Supplying a name (usage error, exit 2)
$ sauron describe provider claude
Error: describe provider takes no name

# Unknown field (usage error, exit 2)
$ sauron describe provider --fields color
Error: --fields must name only: name, skills-location, agents-location
```

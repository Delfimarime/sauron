# Contract: Command Line — Describe Artifact

Conventions: [CLI conventions](../../AUTHORING.md).

**Spec**: [Describe Artifact](../spec.md)

Defines the command-line interface for showing one managed skill's or agent's
detail. This is the user-facing contract only.

## Synopsis

```
sauron describe skill <name> [--fields <list>]
sauron describe agent <name> [--fields <list>]
```

Command hierarchy: `sauron` (root) → `describe` (verb) → `skill`/`agent` (noun) →
`<name>`.

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `<name>` | Yes | The managed artifact's name. Realizes [spec](../spec.md) FR-002, FR-004, FR-005. |

## Flags

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--fields <list>` | No | all | Comma-separated fields to show, in order, name first. Valid: `name`, `type`, `source`, `provider`, `path`, `pinned`, `persona`. Realizes [spec](../spec.md) FR-003, FR-004. |

## Output

- **Detail**: the artifact's fields (the selected `--fields`, or all), one per line.
- **Failure**: a single human-readable message on stderr.

## Exit codes

Exit-status meanings are owned by the [CLI conventions](../../AUTHORING.md);
this table refines which conditions map to each code.

| Code | Meaning | Realizes |
|------|---------|----------|
| `0` | The artifact's detail was shown | [spec](../spec.md) FR-002, FR-003 |
| `2` | Usage error — missing name, or an unknown `--fields` value | [spec](../spec.md) FR-004 |
| `1` | Runtime error — no such managed artifact, or the track file is unreadable | [spec](../spec.md) FR-005, FR-006 |

## Examples

```
# Describe a managed skill
$ sauron describe skill code-review
name:     code-review
type:     skill
source:   team-internal
provider: claude
path:     /home/user/.claude/skills/code-review
pinned:   true
persona:  backend-developer

# A subset of fields (name always first)
$ sauron describe agent triager --fields name,source,pinned
name:   triager
source: team-deploy
pinned: false

# Not installed (runtime error, exit 1)
$ sauron describe skill missing-skill
Error: skill 'missing-skill' is not installed
```

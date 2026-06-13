# Contract: Command Line — List Artifacts

Conventions: [CLI conventions](../../AUTHORING.md).

**Spec**: [List Artifacts](../spec.md)

Defines the command-line interface for listing managed (and, with `--available`,
offered) skills and agents. This is the user-facing contract only.

## Synopsis

```
sauron list skills [--available] [--registry <name>] [--search <term>] [--fields <list>] [--sort <name|registry|type>] [--order <asc|desc>]
sauron list agents [--available] [--registry <name>] [--search <term>] [--fields <list>] [--sort <name|registry|type>] [--order <asc|desc>]
```

Command hierarchy: `sauron` (root) → `list` (verb) → `skills`/`agents` (noun).

## Arguments

None.

## Flags

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--available` | No | false | List what registries offer (with `--registry`, that registry's offerings; without, the resolved catalog) instead of the managed set. Realizes [spec](../spec.md) FR-004, FR-005. |
| `--registry <name>` | No | — | Filter managed artifacts by source registry, or scope `--available` to one registry. Realizes [spec](../spec.md) FR-003, FR-004. |
| `--search <term>` | No | — | Case-insensitive substring filter on name and source registry. Realizes [spec](../spec.md) FR-010. |
| `--fields <list>` | No | `name,source` | Comma-separated columns, name first. Valid: `name`, `type`, `source`, `pinned`, `provider`, `persona`. Realizes [spec](../spec.md) FR-006. |
| `--sort <field>` | No | `name` | `name`, `registry`, or `type`. Realizes [spec](../spec.md) FR-011. |
| `--order <asc\|desc>` | No | `asc` | Sort direction. Realizes [spec](../spec.md) FR-011. |

## Output

- **Table**: one artifact per row, the columns selected by `--fields` (name first); when nothing matches, a single message.
- **Failure**: a single human-readable message on stderr.

## Exit codes

Exit-status meanings are owned by the [CLI conventions](../../AUTHORING.md);
this table refines which conditions map to each code.

| Code | Meaning | Realizes |
|------|---------|----------|
| `0` | Listed, including an empty set or no matches | [spec](../spec.md) FR-002, FR-012 |
| `2` | Usage error — invalid `--sort`, `--order`, or `--fields` value | [spec](../spec.md) FR-007 |
| `1` | Runtime error — the track file is unreadable, or an `--available` registry is unreachable | [spec](../spec.md) FR-008, FR-009 |

## Examples

```
# Managed skills with their source registry
$ sauron list skills
NAME         SOURCE
code-review  team-internal
design-oas3  team-deploy

# Managed skills from one registry, showing the pinned column
$ sauron list skills --registry team-internal --fields name,source,pinned
NAME         SOURCE         PINNED
code-review  team-internal  true

# What a registry offers (live)
$ sauron list agents --available --registry team-deploy
NAME              SOURCE
software-engineer team-deploy
triager           team-deploy

# The resolved catalog (winning registry per name, pin > priority)
$ sauron list skills --available

# Nothing installed (exit 0)
$ sauron list agents
No managed agents.

# Invalid sort (usage error, exit 2)
$ sauron list skills --sort size
Error: --sort accepts only 'name', 'registry', or 'type'
```

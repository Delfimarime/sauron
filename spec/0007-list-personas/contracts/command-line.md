# Contract: Command Line — List Personas

Conventions: [CLI conventions](../../AUTHORING.md).

**Spec**: [List Personas](../spec.md)

Defines the command-line interface for listing registered personas. This is the user-facing contract only. Listing is read-only.

## Synopsis

```
sauron list personas [--search <term>] [--tag <tag>]... [--sort <name|priority>] [--order <asc|desc>]
```

Command hierarchy: `sauron` (root) → `list` (group) → `personas` (subcommand).

## Arguments

None.

## Flags

| Flag | Required | Default | Values | Description |
|------|----------|---------|--------|-------------|
| `--search` | No | — | text | Case-insensitive substring; keeps personas whose name or description contains it. Realizes [spec](../spec.md) FR-010. |
| `--tag` | No | — | text (repeatable) | Keeps personas carrying every given tag (all-match). Realizes [spec](../spec.md) FR-011. |
| `--sort` | No | priority | name, priority | Attribute to order by. Realizes [spec](../spec.md) FR-013, FR-008. |
| `--order` | No | asc | asc, desc | Sort direction. Realizes [spec](../spec.md) FR-014, FR-009. |

## Output

- **Success**: a table on stdout with columns NAME, PRIORITY, TAGS, SKILLS, AGENTS (the last two as counts; PRIORITY shows `-` when undefined), ordered by the chosen attribute and direction. When nothing is registered or nothing matches, a single message instead of a table.
- **Failure**: a single human-readable message on stderr.

## Exit codes

Exit-status meanings are owned by the [CLI conventions](../../AUTHORING.md);
this table refines which conditions map to each code.

| Code | Meaning | Realizes |
|------|---------|----------|
| `0` | Listed (including an empty list or no matches) | [spec](../spec.md) FR-002, FR-004, FR-005 |
| `2` | Usage error — `--search`/`--tag` without a value, invalid `--sort`, or invalid `--order` | [spec](../spec.md) FR-010, FR-011, FR-008, FR-009 |
| `1` | Runtime error — the settings cannot be read or parsed | [spec](../spec.md) FR-007 |

## Examples

```
# List all (default: priority ascending; undefined priorities last)
$ sauron list personas
NAME               PRIORITY  TAGS             SKILLS  AGENTS
backend-developer  0         backend, golang  2       1
qa-engineer        1         qa               1       0
designer           -         design           1       1

# Filter by tags (all must match)
$ sauron list personas --tag backend --tag golang
NAME               PRIORITY  TAGS             SKILLS  AGENTS
backend-developer  0         backend, golang  2       1

# Filter by search (name or description)
$ sauron list personas --search qa
NAME         PRIORITY  TAGS  SKILLS  AGENTS
qa-engineer  1         qa    1       0

# Sort by name, descending
$ sauron list personas --sort name --order desc
NAME               PRIORITY  TAGS             SKILLS  AGENTS
qa-engineer        1         qa               1       0
designer           -         design           1       1
backend-developer  0         backend, golang  2       1

# Nothing registered (exit 0)
$ sauron list personas
No personas registered.

# Invalid sort attribute (usage error, exit 2)
$ sauron list personas --sort tags
Error: --sort must be one of: name, priority
```

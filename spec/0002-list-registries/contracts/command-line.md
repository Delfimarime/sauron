# Contract: Command Line — List Registries

Conventions: [CLI conventions](../../AUTHORING.md).

**Spec**: [List Registries](../spec.md)

Defines the command-line interface for listing registered registries. This is the user-facing contract only. Listing is read-only and spans all registry kinds.

## Synopsis

```
sauron list registries [--search <term>] [--fields <list>] [--sort <name|priority|kind>] [--order <asc|desc>]
```

Command hierarchy: `sauron` (root) → `list` (group) → `registries` (subcommand).

## Arguments

None.

## Flags

| Flag | Required | Default | Values | Description |
|------|----------|---------|--------|-------------|
| `--search` | No | — | text | Case-insensitive substring; keeps registries whose name or location contains it. Realizes [spec](../spec.md) FR-010. |
| `--fields` | No | all columns | comma-separated `name`, `kind`, `priority`, `location` | Columns to display, in order; `name` is always present and first. Realizes [spec](../spec.md) FR-013, FR-014. |
| `--sort` | No | priority | name, priority, kind | Attribute to order by. Realizes [spec](../spec.md) FR-011, FR-008. |
| `--order` | No | asc | asc, desc | Sort direction. Realizes [spec](../spec.md) FR-012, FR-009. |

## Output

- **Success**: a table on stdout with columns NAME, KIND, PRIORITY, LOCATION, ordered by the chosen attribute and direction (default: priority ascending). When nothing is registered or nothing matches, a single message instead of a table.
- **Failure**: a single human-readable message on stderr.

## Exit codes

Exit-status meanings are owned by the [CLI conventions](../../AUTHORING.md);
this table refines which conditions map to each code.

| Code | Meaning | Realizes |
|------|---------|----------|
| `0` | Listed (including an empty list or no matches) | [spec](../spec.md) FR-002, FR-004, FR-005 |
| `2` | Usage error — `--search` without a value, invalid `--sort`, invalid `--order`, or an unknown `--fields` column | [spec](../spec.md) FR-010, FR-008, FR-009, FR-014 |
| `1` | Runtime error — `registries.yaml` cannot be read or parsed | [spec](../spec.md) FR-007 |

## Examples

```
# List all (default: priority ascending)
$ sauron list registries
NAME         KIND        PRIORITY  LOCATION
team-http    http        1         https://skills.example.com
team-deploy  git         2         ssh://git@github.com/acme/agents.git
local-dir    filesystem  3         /home/user/team-skills

# Sort by name, descending
$ sauron list registries --sort name --order desc
NAME         KIND        PRIORITY  LOCATION
team-http    http        1         https://skills.example.com
team-deploy  git         2         ssh://git@github.com/acme/agents.git
local-dir    filesystem  3         /home/user/team-skills

# Sort by kind
$ sauron list registries --sort kind
NAME         KIND        PRIORITY  LOCATION
local-dir    filesystem  3         /home/user/team-skills
team-deploy  git         2         ssh://git@github.com/acme/agents.git
team-http    http        1         https://skills.example.com

# Filter by search (name or location)
$ sauron list registries --search github
NAME         KIND  PRIORITY  LOCATION
team-deploy  git   2         ssh://git@github.com/acme/agents.git

# Nothing registered (exit 0)
$ sauron list registries
No registries registered.

# Invalid sort attribute (usage error, exit 2)
$ sauron list registries --sort location
Error: --sort must be one of: name, priority, kind
```

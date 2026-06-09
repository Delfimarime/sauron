# Contract: Command Line — List Repositories

**Spec**: `../spec.md` (List Repositories)
**Status**: Draft

Defines the command-line interface for listing registered repositories. This is the user-facing contract only. Listing is read-only and spans all repository kinds.

## Synopsis

```
sauron list repositories [--search <term>] [--sort <name|priority|kind>] [--order <asc|desc>]
```

Command hierarchy: `sauron` (root) → `list` (group) → `repositories` (subcommand).

## Arguments

None.

## Flags

| Flag | Required | Default | Values | Description |
|------|----------|---------|--------|-------------|
| `--search` | No | — | text | Case-insensitive substring; keeps repositories whose name or location contains it. Realizes FR-003. |
| `--sort` | No | priority | name, priority, kind | Attribute to order by. Realizes FR-004, FR-011. |
| `--order` | No | asc | asc, desc | Sort direction. Realizes FR-005, FR-012. |

## Output

- **Success**: a table on stdout with columns NAME, KIND, PRIORITY, LOCATION, ordered by the chosen attribute and direction (default: priority ascending). When nothing is registered or nothing matches, a single message instead of a table.
- **Failure**: a single human-readable message on stderr.

## Exit codes

| Code | Meaning | Realizes |
|------|---------|----------|
| `0` | Listed (including an empty list or no matches) | FR-002, FR-007, FR-008 |
| `2` | Usage error — `--search` without a value, invalid `--sort`, or invalid `--order` | FR-003, FR-011, FR-012 |
| `1` | Configuration error — configuration cannot be read or parsed | FR-010 |

## Examples

```
# List all (default: priority ascending)
$ sauron list repositories
NAME         KIND        PRIORITY  LOCATION
team-http    http        1         https://skills.example.com
team-deploy  git         2         ssh://git@github.com/acme/agents.git
local-dir    filesystem  3         /home/user/team-skills

# Sort by name, descending
$ sauron list repositories --sort name --order desc
NAME         KIND        PRIORITY  LOCATION
team-http    http        1         https://skills.example.com
team-deploy  git         2         ssh://git@github.com/acme/agents.git
local-dir    filesystem  3         /home/user/team-skills

# Sort by kind
$ sauron list repositories --sort kind
NAME         KIND        PRIORITY  LOCATION
local-dir    filesystem  3         /home/user/team-skills
team-deploy  git         2         ssh://git@github.com/acme/agents.git
team-http    http        1         https://skills.example.com

# Filter by search (name or location)
$ sauron list repositories --search github
NAME         KIND  PRIORITY  LOCATION
team-deploy  git   2         ssh://git@github.com/acme/agents.git

# Nothing registered (exit 0)
$ sauron list repositories
No repositories registered.

# Invalid sort attribute (usage error, exit 2)
$ sauron list repositories --sort location
Error: --sort must be one of: name, priority, kind
```

# Contract: Command Line — List Personas

Conventions: [CLI conventions](../../AUTHORING.md).

**Spec**: [List Personas](../spec.md)

Defines the command-line interface for listing the available personas and
marking which entries are installed. This is the user-facing contract only. The
available personas are the [installed](../../0014-select-personas/spec.md) ones
merged with the personas the [backend](../../0012-backend/spec.md) offers,
fetched live when the command runs. Listing is read-only and never persists a
catalog. When the backend is unreachable the live fetch is skipped, so only the
installed personas are listed and the command still succeeds.

## Synopsis

```
sauron list personas [--search <term>] [--tag <tag>]... [--installed <true|false>] [--fields <list>] [--sort <name|installed|priority|last-updated|last-synced>] [--order <asc|desc>]
```

Command hierarchy: `sauron` (root) → `list` (group) → `personas` (subcommand).

## Arguments

None.

## Flags

| Flag | Required | Default | Values | Description |
|------|----------|---------|--------|-------------|
| `--search` | No | — | text | Case-insensitive substring; keeps personas whose name or description contains it. Realizes [spec](../spec.md) FR-013. |
| `--tag` | No | — | text (repeatable) | Keeps personas carrying every given tag (all-match). Realizes [spec](../spec.md) FR-014. |
| `--installed` | No | — | `true`, `false` | Keeps only installed personas (`true`) or only not-installed personas (`false`); omitted keeps both. Realizes [spec](../spec.md) FR-015, FR-011. |
| `--fields` | No | all columns | `installed`, `priority`, `tags`, `skills`, `agents`, `last-updated`, `last-synced` | Comma-separated columns, in order; `name` is always present and first. Realizes [spec](../spec.md) FR-017, FR-012. |
| `--sort` | No | priority | name, installed, priority, last-updated, last-synced | Attribute to order by; `priority` orders installed personas ascending and places not-installed ones last. Realizes [spec](../spec.md) FR-018, FR-009, FR-004. |
| `--order` | No | asc | asc, desc | Sort direction. Realizes [spec](../spec.md) FR-019, FR-010. |

## Output

- **Success**: a table on stdout. Columns: NAME, INSTALLED (`yes`/`no`),
  PRIORITY, TAGS, SKILLS, AGENTS, LAST UPDATED, LAST SYNCED. SKILLS and AGENTS
  are counts; PRIORITY and LAST SYNCED show `-` for a not-installed persona,
  since those values exist only for installed personas; LAST UPDATED comes from
  the backend definition and is shown for the personas the backend offers live.
  `--fields` selects and orders the displayed columns (NAME always first). Rows
  are ordered by the chosen attribute and direction. When no personas are
  available or nothing matches, a single message is printed instead of a table.
  When the backend is unreachable the live fetch is skipped and only installed
  personas are listed; the command still succeeds.
- **Failure**: a single human-readable message on stderr.

## Exit codes

Exit-status meanings are owned by the [CLI conventions](../../AUTHORING.md);
this table refines which conditions map to each code.

| Code | Meaning | Realizes |
|------|---------|----------|
| `0` | Listed (including no personas, no matches, or the backend unreachable so only installed personas are listed) | [spec](../spec.md) FR-002, FR-005, FR-006, FR-007 |
| `2` | Usage error — `--search`/`--tag`/`--installed`/`--fields`/`--sort`/`--order` without a value or with an invalid value, or `--fields` naming an unknown column | [spec](../spec.md) FR-009, FR-010, FR-011, FR-012, FR-013, FR-014, FR-015 |
| `1` | Runtime error — `personas.yaml` cannot be read or parsed | [spec](../spec.md) FR-008 |

## Examples

```
# List the available personas (default: priority ascending, installed first, not-installed last)
$ sauron list personas
NAME               INSTALLED  PRIORITY  TAGS             SKILLS  AGENTS  LAST UPDATED          LAST SYNCED
backend-developer  yes        0         backend, golang  2       1       2026-06-11T18:00:00Z  2026-06-12T09:30:00Z
qa-engineer        yes        1         qa               1       0       2026-06-10T12:00:00Z  2026-06-12T09:30:00Z
designer           no         -         design           1       1       2026-06-09T08:00:00Z  -

# Filter by tags (all must match)
$ sauron list personas --tag backend --tag golang
NAME               INSTALLED  PRIORITY  TAGS             SKILLS  AGENTS  LAST UPDATED          LAST SYNCED
backend-developer  yes        0         backend, golang  2       1       2026-06-11T18:00:00Z  2026-06-12T09:30:00Z

# Filter by search (name or description)
$ sauron list personas --search qa
NAME         INSTALLED  PRIORITY  TAGS  SKILLS  AGENTS  LAST UPDATED          LAST SYNCED
qa-engineer  yes        1         qa    1       0       2026-06-10T12:00:00Z  2026-06-12T09:30:00Z

# Only not-installed personas (offered live by the backend)
$ sauron list personas --installed false
NAME      INSTALLED  PRIORITY  TAGS    SKILLS  AGENTS  LAST UPDATED          LAST SYNCED
designer  no         -         design  1       1       2026-06-09T08:00:00Z  -

# Choose and order columns
$ sauron list personas --fields installed,tags,skills,agents
NAME               INSTALLED  TAGS             SKILLS  AGENTS
backend-developer  yes        backend, golang  2       1
qa-engineer        yes        qa               1       0
designer           no         design           1       1

# Sort by name, descending
$ sauron list personas --sort name --order desc
NAME               INSTALLED  PRIORITY  TAGS             SKILLS  AGENTS  LAST UPDATED          LAST SYNCED
qa-engineer        yes        1         qa               1       0       2026-06-10T12:00:00Z  2026-06-12T09:30:00Z
designer           no         -         design           1       1       2026-06-09T08:00:00Z  -
backend-developer  yes        0         backend, golang  2       1       2026-06-11T18:00:00Z  2026-06-12T09:30:00Z

# Backend unreachable: only installed personas are listed; still exits 0
$ sauron list personas
NAME               INSTALLED  PRIORITY  TAGS             SKILLS  AGENTS  LAST UPDATED          LAST SYNCED
backend-developer  yes        0         backend, golang  2       1       2026-06-11T18:00:00Z  2026-06-12T09:30:00Z
qa-engineer        yes        1         qa               1       0       2026-06-10T12:00:00Z  2026-06-12T09:30:00Z

# No personas available (nothing installed, backend offers none or is unreachable) — exit 0
$ sauron list personas
There are no personas.

# Invalid sort attribute (usage error, exit 2)
$ sauron list personas --sort tags
Error: --sort must be one of: name, installed, priority, last-updated, last-synced

# Unknown field (usage error, exit 2)
$ sauron list personas --fields foo
Error: --fields must be a subset of: installed, priority, tags, skills, agents, last-updated, last-synced
```

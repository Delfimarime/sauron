# `list skills` — command line

```
sauron list skills [--search <term>] [--fields <list>] [--sort name|lastUpdatedAt] [--order asc|desc] [--page <n>] [--limit <n>]
```

List the installed skills, paginated.

## Flags

| Flag | Meaning |
|---|---|
| `--search <term>` | Case-insensitive substring filter on the skill name |
| `--fields <list>` | Columns to display, in order; `name` is always present and first. Valid: `name`, `version`, `lastUpdatedAt` |
| `--sort <field>` | Sort field: `name` (default) or `lastUpdatedAt` |
| `--order <asc\|desc>` | Sort direction, default `asc` |
| `--page <n>` | Page number, 1-based (default `1`) |
| `--limit <n>` | Page size (default `20`) |

## Fields

| Field | Description |
|---|---|
| `name` | The skill's name |
| `version` | Optional human-meaningful version; `—` when none |
| `lastUpdatedAt` | When the artifact was last updated |

## Output

A table on stdout, one installed skill per row after filter, sort, and paging,
followed by a line reporting the applied page/limit (no total count). An empty
page prints no table row, only the paging line (`showing 0 results (page <n>,
limit <n>)`), and exits `0`.

## Example

```
$ sauron list skills --limit 20
name        version  lastUpdatedAt
go-style    v1.4.0   2026-06-15
sql-review  —        2026-06-12
showing 1–2 (page 1, limit 20)
```

## Exit codes

| Code | Condition |
|---|---|
| `0` | The list was produced (including empty) |
| `2` | Invalid flags |
| `1` | `track.yaml` is unreadable |

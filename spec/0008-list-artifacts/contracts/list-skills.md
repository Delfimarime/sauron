# `list skills` — command line

```
sauron list skills [--search <term>] [--fields <list>] [--sort name|lastUpdatedAt] [--order asc|desc]
```

List the installed skills.

## Flags

| Flag | Meaning |
|---|---|
| `--search <term>` | Case-insensitive substring filter on the skill name |
| `--fields <list>` | Columns to display, in order; `name` is always present and first. Valid: `name`, `version`, `lastUpdatedAt` |
| `--sort <field>` | Sort field: `name` (default) or `lastUpdatedAt` |
| `--order <asc\|desc>` | Sort direction, default `asc` |

## Fields

| Field | Description |
|---|---|
| `name` | The skill's name |
| `version` | Optional human-meaningful version; `—` when none |
| `lastUpdatedAt` | When the artifact was last updated |

## Output

A table on stdout, one installed skill per row. An empty set prints an empty
result and exits `0`.

## Example

```
$ sauron list skills
name        version  lastUpdatedAt
go-style    v1.4.0   2026-06-15
sql-review  —        2026-06-12
```

## Exit codes

| Code | Condition |
|---|---|
| `0` | The list was produced (including empty) |
| `2` | Invalid flags |
| `1` | `track.yaml` is unreadable |

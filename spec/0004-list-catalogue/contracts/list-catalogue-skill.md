# `list catalogue skill` — command line

```
sauron list catalogue skill [--search <term>] [--page <n>] [--limit <n>] [--sort name] [--order asc|desc]
```

Browse the skills the registry offers, live and paginated.

## Flags

| Flag | Meaning |
|---|---|
| `--search <term>` | Case-insensitive substring filter on the skill name |
| `--page <n>` | Page number, 1-based (default `1`) |
| `--limit <n>` | Page size (default `20`) |
| `--sort <field>` | Sort field: `name` (default) |
| `--order <asc\|desc>` | Sort direction, default `asc` |

## Output

A table on stdout of the registry's offered skills, after filter, sort, and
paging, with a line reporting the applied page/limit (no total count).

## Example

```
$ sauron list catalogue skill --limit 20
NAME        KIND
go-style    skill
sql-review  skill
showing 1–2 (page 1, limit 20)
```

## Exit codes

| Code | Condition |
|---|---|
| `0` | The catalogue page was produced |
| `2` | Missing/invalid arguments or flags |
| `1` | No registry is set, or the registry is unreachable |

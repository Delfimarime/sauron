# `list catalogue skill` — command line

```
sauron list catalogue skill <registry> [--search <term>] [--offset <n>] [--limit <n>] [--sort name] [--order asc|desc]
```

Browse the skills a registry offers, live and paginated.

## Arguments

| Argument | Required | Meaning |
|---|---|---|
| `<registry>` | yes | The registry to browse |

## Flags

| Flag | Meaning |
|---|---|
| `--search <term>` | Case-insensitive substring filter on the skill name |
| `--offset <n>` | Leading results to skip (default `0`) |
| `--limit <n>` | Maximum number of results to return |
| `--sort <field>` | Sort field: `name` (default) |
| `--order <asc\|desc>` | Sort direction, default `asc` |

## Output

A table on stdout of the registry's offered skills, after filter, sort, and
paging, with a line reporting the applied offset/limit.

## Example

```
$ sauron list catalogue skill acme --limit 20
NAME        KIND
go-style    skill
sql-review  skill
showing 1–2 of 2 (offset 0, limit 20)
```

## Exit codes

| Code | Condition |
|---|---|
| `0` | The catalogue page was produced |
| `2` | Missing/invalid arguments or flags |
| `1` | The registry does not exist or is unreachable |

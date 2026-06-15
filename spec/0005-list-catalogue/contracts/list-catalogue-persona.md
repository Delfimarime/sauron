# `list catalogue persona` — command line

```
sauron list catalogue persona <registry> [--search <term>] [--offset <n>] [--limit <n>] [--sort name] [--order asc|desc]
```

Browse the personas a registry offers, live and paginated. A persona entry can
surface the membership it would resolve to.

## Arguments

| Argument | Required | Meaning |
|---|---|---|
| `<registry>` | yes | The registry to browse |

## Flags

| Flag | Meaning |
|---|---|
| `--search <term>` | Case-insensitive substring filter on the persona name |
| `--offset <n>` | Leading results to skip (default `0`) |
| `--limit <n>` | Maximum number of results to return |
| `--sort <field>` | Sort field: `name` (default) |
| `--order <asc\|desc>` | Sort direction, default `asc` |

## Output

A table on stdout of the registry's offered personas, after filter, sort, and
paging, with a line reporting the applied offset/limit; an entry may summarize the
membership the persona declares.

## Example

```
$ sauron list catalogue persona acme --limit 20
NAME         MEMBERS
backend-dev  skills: go-style, sql-review; agents: code-reviewer
showing 1–1 of 1 (offset 0, limit 20)
```

## Exit codes

| Code | Condition |
|---|---|
| `0` | The catalogue page was produced |
| `2` | Missing/invalid arguments or flags |
| `1` | The registry does not exist or is unreachable |

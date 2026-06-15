# `list registries` — command line

```
sauron list registries [--search <term>] [--fields <list>] [--sort name|transport] [--order asc|desc]
```

Review the configured sources.

## Flags

| Flag | Meaning |
|---|---|
| `--search <term>` | Case-insensitive substring filter on the registry name |
| `--fields <list>` | Columns to display, in order; `name` is always present and first. Valid: `name`, `transport`, `uri`, `timeout` |
| `--sort <field>` | Sort field: `name` (default) or `transport` |
| `--order <asc\|desc>` | Sort direction, default `asc` |

## Output

A table on stdout, one registry per row. An empty registry set prints an empty
result and exits `0`.

## Exit codes

| Code | Condition |
|---|---|
| `0` | The list was produced (including empty) |
| `2` | Invalid flags |
| `1` | `registries.yaml` is unreadable |

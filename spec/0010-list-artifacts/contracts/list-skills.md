# `list skills` — command line

```
sauron list skills [--search <term>] [--fields <list>] [--sort name|registry|updated] [--order asc|desc]
```

List the installed skills.

## Flags

| Flag | Meaning |
|---|---|
| `--search <term>` | Case-insensitive substring filter on the skill name |
| `--fields <list>` | Columns to display, in order; `name` is always present and first. Valid: `name`, `registry`, `version`, `provenance`, `updated` |
| `--sort <field>` | Sort field: `name` (default), `registry`, or `updated` |
| `--order <asc\|desc>` | Sort direction, default `asc` |

## Output

A table on stdout, one installed skill per row. An empty set prints an empty
result and exits `0`.

## Exit codes

| Code | Condition |
|---|---|
| `0` | The list was produced (including empty) |
| `2` | Invalid flags |
| `1` | `track.yaml` is unreadable |

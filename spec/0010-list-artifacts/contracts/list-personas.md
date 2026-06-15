# `list personas` — command line

```
sauron list personas [--search <term>] [--fields <list>] [--sort name|registry|updated] [--order asc|desc]
```

List the installed personas. Unlike skills and agents, a persona row can surface
its resolved membership.

## Flags

| Flag | Meaning |
|---|---|
| `--search <term>` | Case-insensitive substring filter on the persona name |
| `--fields <list>` | Columns to display, in order; `name` is always present and first. Valid: `name`, `registry`, `version`, `members`, `provenance`, `updated` |
| `--sort <field>` | Sort field: `name` (default), `registry`, or `updated` |
| `--order <asc\|desc>` | Sort direction, default `asc` |

## Output

A table on stdout, one installed persona per row. With `members` in `--fields`,
each row summarizes the persona's resolved skills and agents. An empty set prints
an empty result and exits `0`.

## Exit codes

| Code | Condition |
|---|---|
| `0` | The list was produced (including empty) |
| `2` | Invalid flags |
| `1` | `track.yaml` is unreadable |

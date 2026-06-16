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
| `--fields <list>` | Columns to display, in order; `name` is always present and first. Valid: `name`, `registry`, `version`, `members`, `updated` |
| `--sort <field>` | Sort field: `name` (default), `registry`, or `updated` |
| `--order <asc\|desc>` | Sort direction, default `asc` |

## Fields

| Field | Description |
|---|---|
| `name` | The persona's name |
| `registry` | The source registry it was installed from |
| `version` | Optional version of the persona definition; `—` when none |
| `members` | The resolved skills and agents the persona brings in |
| `updated` | When the persona was last re-resolved or updated |

## Output

A table on stdout, one installed persona per row. With `members` in `--fields`,
each row summarizes the persona's resolved skills and agents. An empty set prints
an empty result and exits `0`.

## Example

```
$ sauron list personas --fields name,registry,members
NAME         REGISTRY  MEMBERS
backend-dev  acme      skills: go-style, sql-review; agents: code-reviewer
```

## Exit codes

| Code | Condition |
|---|---|
| `0` | The list was produced (including empty) |
| `2` | Invalid flags |
| `1` | `track.yaml` is unreadable |

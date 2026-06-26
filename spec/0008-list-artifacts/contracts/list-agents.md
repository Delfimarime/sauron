# `list agents` — command line

```
sauron list agents [--search <term>] [--fields <list>] [--sort name|updated] [--order asc|desc]
```

List the installed agents.

## Flags

| Flag | Meaning |
|---|---|
| `--search <term>` | Case-insensitive substring filter on the agent name |
| `--fields <list>` | Columns to display, in order; `name` is always present and first. Valid: `name`, `version`, `updated` |
| `--sort <field>` | Sort field: `name` (default) or `updated` |
| `--order <asc\|desc>` | Sort direction, default `asc` |

## Fields

| Field | Description |
|---|---|
| `name` | The agent's name |
| `version` | Optional human-meaningful version; `—` when none |
| `updated` | When the artifact was last updated |

## Output

A table on stdout, one installed agent per row. An empty set prints an empty
result and exits `0`.

## Example

```
$ sauron list agents
NAME           VERSION  UPDATED
code-reviewer  3af1c2e  2026-06-14
```

## Exit codes

| Code | Condition |
|---|---|
| `0` | The list was produced (including empty) |
| `2` | Invalid flags |
| `1` | `track.yaml` is unreadable |

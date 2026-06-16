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

## Fields

| Field | Description |
|---|---|
| `name` | The skill's name |
| `registry` | The source registry it was installed from |
| `version` | Optional human-meaningful version; `—` when none |
| `provenance` | Why it is installed: `direct` and/or `via persona <name>` |
| `updated` | When the artifact was last updated |

## Output

A table on stdout, one installed skill per row. An empty set prints an empty
result and exits `0`.

## Example

```
$ sauron list skills
NAME        REGISTRY  VERSION  UPDATED
go-style    acme      v1.4.0   2026-06-15
sql-review  acme      —        2026-06-12
```

## Exit codes

| Code | Condition |
|---|---|
| `0` | The list was produced (including empty) |
| `2` | Invalid flags |
| `1` | `track.yaml` is unreadable |

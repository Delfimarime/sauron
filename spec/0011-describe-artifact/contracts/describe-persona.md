# `describe persona` — command line

```
sauron describe persona <name> [--fields <list>]
```

Show one installed persona's full detail, including its resolved membership.

## Arguments

| Argument | Required | Meaning |
|---|---|---|
| `<name>` | yes | The installed persona to describe |

## Flags

| Flag | Meaning |
|---|---|
| `--fields <list>` | Fields to display, in order; `name` is always present and first. Valid: `name`, `registry`, `version`, `digest`, `members`, `installed`, `updated` |

## Output

The persona's fields on stdout, including its resolved membership (the skills and
agents it brings in).

## Exit codes

| Code | Condition |
|---|---|
| `0` | The detail was produced |
| `2` | Missing/invalid arguments or flags |
| `1` | No installed persona of that name exists, or `track.yaml` is unreadable |

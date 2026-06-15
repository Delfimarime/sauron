# `describe skill` — command line

```
sauron describe skill <name> [--fields <list>]
```

Show one installed skill's full detail.

## Arguments

| Argument | Required | Meaning |
|---|---|---|
| `<name>` | yes | The installed skill to describe |

## Flags

| Flag | Meaning |
|---|---|
| `--fields <list>` | Fields to display, in order; `name` is always present and first. Valid: `name`, `registry`, `version`, `digest`, `path`, `provenance`, `installed`, `updated` |

## Output

The skill's fields on stdout.

## Exit codes

| Code | Condition |
|---|---|
| `0` | The detail was produced |
| `2` | Missing/invalid arguments or flags |
| `1` | No installed skill of that name exists, or `track.yaml` is unreadable |

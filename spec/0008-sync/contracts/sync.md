# `sync` — command line

```
sauron sync [skills|agents|personas]... [--dry-run]
```

Fully reconcile the installed set against its sources.

## Arguments

| Argument | Required | Meaning |
|---|---|---|
| kind list | no | Any of `skills`, `agents`, `personas` to scope the run; omitted means all |

## Flags

| Flag | Meaning |
|---|---|
| `--dry-run` | Print the reconcile plan without changing the environment or the track file |

## Output

The plan grouped under `skills:`, `agents:`, and `personas:`, prefixed `+`
(added), `~` (updated), or `-` (removed), followed by a summary count when applied.
An already-current set prints no changes and exits `0`. Per-artifact failures are
reported without stopping the run.

## Exit codes

| Code | Condition |
|---|---|
| `0` | The set was reconciled, already current, or `--dry-run` |
| `2` | Invalid arguments or flags |
| `1` | No provider is set, or the track file could not be read or written |

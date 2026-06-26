# `sync` — command line

```
sauron sync [skills|agents]... [--dry-run]
```

Fully reconcile the installed set against the registry.

## Arguments

| Argument | Required | Meaning |
|---|---|---|
| kind list | no | Any of `skills`, `agents` to scope the run; omitted means all |

## Flags

| Flag | Meaning |
|---|---|
| `--dry-run` | Print the reconcile plan without changing the environment or the track file |

## Output

The plan grouped under `skills:` and `agents:`, prefixed `+` (added), `~`
(updated), or `-` (removed), followed by a summary count when applied. An
already-current set prints no changes and exits `0`. Per-artifact failures are
reported without stopping the run.

## Example

```
$ sauron sync
skills:
  ~ sauron-acme-go-style
  - sauron-acme-old-skill
agents:
  + sauron-acme-new-reviewer
1 added, 1 updated, 1 removed
```

## Exit codes

| Code | Condition |
|---|---|
| `0` | The set was reconciled, already current, or `--dry-run` |
| `2` | Invalid arguments or flags |
| `1` | No provider is set, or the track file could not be read or written |

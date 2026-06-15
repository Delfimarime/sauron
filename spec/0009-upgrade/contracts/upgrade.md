# `upgrade` — command line

```
sauron upgrade [skills|agents|personas]... [--dry-run]
```

Non-destructively refresh the installed set to the latest from its sources.

## Arguments

| Argument | Required | Meaning |
|---|---|---|
| kind list | no | Any of `skills`, `agents`, `personas` to scope the run; omitted means all |

## Flags

| Flag | Meaning |
|---|---|
| `--dry-run` | Print the upgrade plan without changing the environment or the track file |

## Output

The plan grouped under `skills:`, `agents:`, and `personas:`, prefixed `+` (added)
or `~` (updated), followed by a summary count when applied. No `-` lines ever
appear: upgrade never removes. An already-current set prints no changes and exits
`0`. Per-artifact failures are reported without stopping the run.

## Example

```
$ sauron upgrade skills personas
skills:
  ~ sauron-acme-go-style
personas:
  + new-helper-skill (via backend-dev)
1 added, 1 updated
```

## Exit codes

| Code | Condition |
|---|---|
| `0` | The set was refreshed, already current, or `--dry-run` |
| `2` | Invalid arguments or flags |
| `1` | No provider is set, or the track file could not be read or written |

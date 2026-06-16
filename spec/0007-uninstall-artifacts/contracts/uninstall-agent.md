# `uninstall agent` — command line

```
sauron uninstall agent <registry> <name>... [--dry-run]
```

Remove named installed agents.

## Arguments

| Argument | Required | Meaning |
|---|---|---|
| `<registry>` | yes | The registry the agents were installed from |
| `<name>...` | yes | One or more agent names to remove |

## Flags

| Flag | Meaning |
|---|---|
| `--dry-run` | Print the removal plan without changing the environment or the track file |

## Output

The plan under an `agents:` heading, prefixed `-`, with a summary count when
applied. Uninstalling something not installed reports nothing was removed and
exits `0`.

## Example

```
$ sauron uninstall agent acme code-reviewer
agents:
  - sauron-acme-code-reviewer
1 removed
```

## Exit codes

| Code | Condition |
|---|---|
| `0` | The agents were removed, none were installed, or `--dry-run` |
| `2` | Missing/invalid arguments or flags |
| `1` | Track file unreadable, or a removal could not be persisted |

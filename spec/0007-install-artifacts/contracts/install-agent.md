# `install agent` — command line

```
sauron install agent <name>...
```

Install named agents from the registry into the active provider.

## Arguments

| Argument | Required | Meaning |
|---|---|---|
| `<name>...` | yes | One or more agent names to install |

## Output

The plan under an `agents:` heading, `+` for additions and `~` for updates,
followed by a summary count. Per-name failures (a name the registry does not
offer) are reported without stopping the run.

## Example

```
$ sauron install agent code-reviewer
agents:
  + sauron-code-reviewer
1 added
```

## Exit codes

| Code | Condition |
|---|---|
| `0` | The named agents were installed or already current |
| `2` | Missing/invalid arguments or flags |
| `1` | No provider is set, the registry is unreachable, or the install could not be persisted |

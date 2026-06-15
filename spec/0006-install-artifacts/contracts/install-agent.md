# `install agent` — command line

```
sauron install agent <registry> <name>...
```

Install named agents from a registry into the active provider.

## Arguments

| Argument | Required | Meaning |
|---|---|---|
| `<registry>` | yes | The source registry |
| `<name>...` | yes | One or more agent names to install |

## Output

The plan under an `agents:` heading, `+` for additions and `~` for updates,
followed by a summary count. Per-name failures (a name the registry does not
offer) are reported without stopping the run.

## Exit codes

| Code | Condition |
|---|---|
| `0` | The named agents were installed or already current |
| `2` | Missing/invalid arguments or flags |
| `1` | No provider is set, the registry is unreachable, or the install could not be persisted |

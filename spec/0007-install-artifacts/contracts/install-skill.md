# `install skill` — command line

```
sauron install skill <name>...
```

Install named skills from the registry into the active provider.

## Arguments

| Argument | Required | Meaning |
|---|---|---|
| `<name>...` | yes | One or more skill names to install |

## Output

The plan under a `skills:` heading, `+` for additions and `~` for updates,
followed by a summary count. Per-name failures (a name the registry does not
offer) are reported without stopping the run.

## Example

```
$ sauron install skill go-style sql-review
skills:
  + sauron-go-style
  + sauron-sql-review
2 added
```

## Exit codes

| Code | Condition |
|---|---|
| `0` | The named skills were installed or already current |
| `2` | Missing/invalid arguments or flags |
| `1` | No provider is set, the registry is unreachable, or the install could not be persisted |

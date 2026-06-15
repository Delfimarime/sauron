# `describe provider` — command line

```
sauron describe provider [--fields <list>]
```

Show the active provider's detail.

## Flags

| Flag | Meaning |
|---|---|
| `--fields <list>` | Fields to display, in order. Valid: `name`, `labels` |

## Output

The active provider on stdout. When no provider is set, a line reporting that none
is set; exits `0`.

## Exit codes

| Code | Condition |
|---|---|
| `0` | The provider detail was produced, or none is set |
| `2` | Invalid flags |
| `1` | `settings.yaml` is unreadable |

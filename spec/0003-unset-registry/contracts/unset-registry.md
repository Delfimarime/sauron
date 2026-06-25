# `unset registry` — command line

```
sauron unset registry [--dry-run]
```

Remove the configured source. Installed artifacts are preserved.

## Flags

| Flag | Meaning |
|---|---|
| `--dry-run` | Report what would be unset without changing the environment or state |

## Output

One confirmation line on stdout. Installed artifacts are left in place. Unsetting
when no registry is configured reports that nothing was unset and exits `0`.

## Example

```
$ sauron unset registry
registry unset; installed artifacts preserved
```

## Exit codes

| Code | Condition |
|---|---|
| `0` | The registry was removed, or none was configured, or `--dry-run` |
| `2` | Missing/invalid arguments or flags |
| `1` | The state file is unreadable, or the removal could not be persisted |

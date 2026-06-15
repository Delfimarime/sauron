# `uninstall persona` — command line

```
sauron uninstall persona <registry> <name>... [--dry-run]
```

Remove named installed personas. Uninstalling a persona removes the members it
brought in, keeping any member also installed directly or brought in by another
persona.

## Arguments

| Argument | Required | Meaning |
|---|---|---|
| `<registry>` | yes | The registry the personas were installed from |
| `<name>...` | yes | One or more persona names to remove |

## Flags

| Flag | Meaning |
|---|---|
| `--dry-run` | Print the removal plan without changing the environment or the track file |

## Output

The plan grouped under `personas:` and the `skills:`/`agents:` headings for the
members removed, prefixed `-`, with a summary count when applied. Uninstalling
something not installed reports nothing was removed and exits `0`.

## Exit codes

| Code | Condition |
|---|---|
| `0` | The personas were removed, none were installed, or `--dry-run` |
| `2` | Missing/invalid arguments or flags |
| `1` | Track file unreadable, or a removal could not be persisted |

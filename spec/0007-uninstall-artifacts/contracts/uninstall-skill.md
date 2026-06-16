# `uninstall skill` — command line

```
sauron uninstall skill <registry> <name>... [--dry-run]
```

Remove named installed skills.

## Arguments

| Argument | Required | Meaning |
|---|---|---|
| `<registry>` | yes | The registry the skills were installed from |
| `<name>...` | yes | One or more skill names to remove |

## Flags

| Flag | Meaning |
|---|---|
| `--dry-run` | Print the removal plan without changing the environment or the track file |

## Output

The plan under a `skills:` heading, prefixed `-`, with a summary count when
applied. Uninstalling something not installed reports nothing was removed and
exits `0`.

## Example

```
$ sauron uninstall skill acme go-style
skills:
  - sauron-acme-go-style
1 removed
```

## Exit codes

| Code | Condition |
|---|---|
| `0` | The skills were removed, none were installed, or `--dry-run` |
| `2` | Missing/invalid arguments or flags |
| `1` | Track file unreadable, or a removal could not be persisted |

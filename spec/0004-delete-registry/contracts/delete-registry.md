# `delete registry` — command line

```
sauron delete registry <name> [--dry-run]
```

Unregister a source and cascade-uninstall every artifact it delivered.

## Arguments

| Argument | Required | Meaning |
|---|---|---|
| `<name>` | yes | The registry to delete |

## Flags

| Flag | Meaning |
|---|---|
| `--dry-run` | Print the cascade plan without changing the environment or the track file |

## Output

The plan of removed artifacts grouped under `skills:`, `agents:`, and `personas:`,
prefixed `-`, with a summary count when applied. Deleting a registry that does not
exist reports that nothing was deleted and exits `0`.

## Exit codes

| Code | Condition |
|---|---|
| `0` | The registry and its artifacts were removed, or it did not exist, or `--dry-run` |
| `2` | Missing/invalid arguments or flags |
| `1` | Configuration or track file unreadable, or the registry removal could not be persisted |

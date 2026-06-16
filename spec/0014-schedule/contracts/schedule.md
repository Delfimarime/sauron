# `schedule` — command line

```
sauron schedule (sync|upgrade) <expression>
```

Register the OS-crontab automation for a reconcile operation.

## Arguments

| Argument | Required | Meaning |
|---|---|---|
| operation noun | yes | `sync` or `upgrade` |
| `<expression>` | yes | The cron expression |

## Output

A confirmation of the operation and expression registered. Re-scheduling an
operation replaces its existing entry rather than adding a second.

## Example

```
$ sauron schedule sync "0 */6 * * *"
scheduled sync: 0 */6 * * *
```

## Exit codes

| Code | Condition |
|---|---|
| `0` | The schedule was registered or replaced |
| `2` | Missing/invalid arguments, or an invalid cron expression |
| `1` | The OS crontab could not be read or written |

# `unschedule` — command line

```
sauron unschedule (sync|upgrade)
```

Remove the OS-crontab automation for a reconcile operation.

## Arguments

| Argument | Required | Meaning |
|---|---|---|
| operation noun | yes | `sync` or `upgrade` |

## Output

A confirmation of removal. Unscheduling an operation that is not scheduled reports
that nothing was removed and exits `0`.

## Example

```
$ sauron unschedule sync
unscheduled sync
```

## Exit codes

| Code | Condition |
|---|---|
| `0` | The schedule was removed, or none was scheduled |
| `2` | Missing/invalid arguments |
| `1` | The OS crontab could not be read or written |

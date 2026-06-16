# `set provider` — command line

```
sauron set provider <claude|zencoder>
```

Set the single global provider destination; migrate installed artifacts on change.

## Arguments

| Argument | Required | Meaning |
|---|---|---|
| `<provider>` | yes | `claude` or `zencoder` |

## Output

On a change, the migration plan grouped under `skills:`, `agents:`, and
`personas:` with a summary count. Setting the already-active provider reports no
change and exits `0`.

## Example

```
$ sauron set provider zencoder
skills:
  ~ sauron-acme-go-style
agents:
  ~ sauron-acme-code-reviewer
provider set to "zencoder"; 2 artifacts migrated
```

## Exit codes

| Code | Condition |
|---|---|
| `0` | The provider was set or migrated, or was already active |
| `2` | Missing argument, or an unsupported provider name |
| `1` | The setting could not be persisted, or a migration step failed |

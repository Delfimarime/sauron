# `describe provider` — command line

```
sauron describe provider [--fields <list>]
```

Show the active provider's detail.

## Flags

| Flag | Meaning |
|---|---|
| `--fields <list>` | Fields to display, in order; `name` is always first. Valid: `name`, `directory`, `labels`, `createdAt`, `lastUpdatedAt`, `lastSyncedAt`, `lastSyncAttemptAt` |

## Output

The active provider on stdout, as an aligned `label: value` view. `directory` is
derived from the provider name (`claude` → `~/.claude`, `zencoder` →
`~/.zencoder`); `labels` renders as an indented, key-sorted section; the audit
(`createdAt`/`lastUpdatedAt`) and sync (`lastSyncedAt`/`lastSyncAttemptAt`) timestamps render
only when present. When no provider is set, a line reporting that none is set;
exits `0`.

## Example

```
$ sauron describe provider
name:           claude
directory:      ~/.claude
createdAt:      2026-06-21T07:30:00Z
lastUpdatedAt:  2026-06-21T07:30:00Z
```

## Exit codes

| Code | Condition |
|---|---|
| `0` | The provider detail was produced, or none is set |
| `2` | Invalid flags |
| `1` | `settings.yaml` is unreadable |

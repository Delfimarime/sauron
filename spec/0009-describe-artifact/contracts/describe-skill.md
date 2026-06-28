# `describe skill` — command line

```
sauron describe skill <name> [--fields <list>]
```

Show one installed skill's full detail.

## Arguments

| Argument | Required | Meaning |
|---|---|---|
| `<name>` | yes | The installed skill to describe |

## Flags

| Flag | Meaning |
|---|---|
| `--fields <list>` | Fields to display, in order; `name` is always present and first. Valid: `name`, `version`, `path`, `installedAt`, `lastUpdatedAt` |

## Fields

| Field | Description |
|---|---|
| `name` | The skill's name |
| `version` | Optional human-meaningful version; `—` when none |
| `path` | Exact installed location under the provider |
| `installedAt` | When it was first installed |
| `lastUpdatedAt` | When it was last updated |

## Output

The skill's fields on stdout.

## Example

```
$ sauron describe skill go-style
name:           go-style
version:        v1.4.0
path:           skills/sauron-acme-go-style
installedAt:    2026-06-10T09:00:00Z
lastUpdatedAt:  2026-06-15T10:00:00Z
```

## Exit codes

| Code | Condition |
|---|---|
| `0` | The detail was produced |
| `2` | Missing/invalid arguments or flags |
| `1` | No installed skill of that name exists, or `track.yaml` is unreadable |

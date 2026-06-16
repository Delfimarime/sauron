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
| `--fields <list>` | Fields to display, in order; `name` is always present and first. Valid: `name`, `registry`, `version`, `digest`, `path`, `provenance`, `installed`, `updated` |

## Fields

| Field | Description |
|---|---|
| `name` | The skill's name |
| `registry` | The source registry it was installed from |
| `version` | Optional human-meaningful version; `—` when none |
| `digest` | Content identity used to detect change and local drift |
| `path` | Exact installed location under the provider |
| `provenance` | Why it is installed: `direct` and/or `via persona <name>` |
| `installed` | When it was first installed |
| `updated` | When it was last updated |

## Output

The skill's fields on stdout.

## Example

```
$ sauron describe skill go-style
name:        go-style
registry:    acme
version:     v1.4.0
digest:      sha256:1a2b…
path:        skills/sauron-acme-go-style
provenance:  direct, via persona backend-dev
installed:   2026-06-10T09:00:00Z
updated:     2026-06-15T10:00:00Z
```

## Exit codes

| Code | Condition |
|---|---|
| `0` | The detail was produced |
| `2` | Missing/invalid arguments or flags |
| `1` | No installed skill of that name exists, or `track.yaml` is unreadable |

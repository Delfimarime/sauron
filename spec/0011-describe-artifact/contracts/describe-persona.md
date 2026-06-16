# `describe persona` — command line

```
sauron describe persona <name> [--fields <list>]
```

Show one installed persona's full detail, including its resolved membership.

## Arguments

| Argument | Required | Meaning |
|---|---|---|
| `<name>` | yes | The installed persona to describe |

## Flags

| Flag | Meaning |
|---|---|
| `--fields <list>` | Fields to display, in order; `name` is always present and first. Valid: `name`, `registry`, `version`, `digest`, `members`, `installed`, `updated` |

## Fields

| Field | Description |
|---|---|
| `name` | The persona's name |
| `registry` | The source registry it was installed from |
| `version` | Optional version of the persona definition; `—` when none |
| `digest` | Content identity of the persona definition |
| `members` | The resolved skills and agents the persona brings in |
| `installed` | When the persona was first installed |
| `updated` | When the persona was last re-resolved or updated |

## Output

The persona's fields on stdout, including its resolved membership (the skills and
agents it brings in).

## Example

```
$ sauron describe persona backend-dev
name:      backend-dev
registry:  acme
version:   9f4d2a1
digest:    sha256:7e1f…
members:
  skills:  go-style, sql-review
  agents:  code-reviewer
installed: 2026-06-10T09:00:00Z
updated:   2026-06-15T10:00:00Z
```

## Exit codes

| Code | Condition |
|---|---|
| `0` | The detail was produced |
| `2` | Missing/invalid arguments or flags |
| `1` | No installed persona of that name exists, or `track.yaml` is unreadable |

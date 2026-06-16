# `describe agent` — command line

```
sauron describe agent <name> [--fields <list>]
```

Show one installed agent's full detail.

## Arguments

| Argument | Required | Meaning |
|---|---|---|
| `<name>` | yes | The installed agent to describe |

## Flags

| Flag | Meaning |
|---|---|
| `--fields <list>` | Fields to display, in order; `name` is always present and first. Valid: `name`, `registry`, `version`, `digest`, `path`, `provenance`, `installed`, `updated` |

## Fields

| Field | Description |
|---|---|
| `name` | The agent's name |
| `registry` | The source registry it was installed from |
| `version` | Optional human-meaningful version; `—` when none |
| `digest` | Content identity used to detect change and local drift |
| `path` | Exact installed location under the provider |
| `provenance` | Why it is installed: `direct` and/or `via persona <name>` |
| `installed` | When it was first installed |
| `updated` | When it was last updated |

## Output

The agent's fields on stdout.

## Example

```
$ sauron describe agent code-reviewer
name:        code-reviewer
registry:    acme
version:     3af1c2e
digest:      sha256:9c4d…
path:        agents/sauron-acme-code-reviewer
provenance:  via persona backend-dev
installed:   2026-06-12T08:30:00Z
updated:     2026-06-14T11:15:00Z
```

## Exit codes

| Code | Condition |
|---|---|
| `0` | The detail was produced |
| `2` | Missing/invalid arguments or flags |
| `1` | No installed agent of that name exists, or `track.yaml` is unreadable |

# `describe registry` — command line

```
sauron describe registry [--fields <list>]
```

Show the configured registry's full detail.

## Flags

| Flag | Meaning |
|---|---|
| `--fields <list>` | Fields to display, in order; `source` is always present and first. Valid: `source`, `transport`, `revision`, `credentials`, `tls`, `sshKey`, `timeout`, `createdAt`, `lastUpdatedAt` |

## Output

The registry's fields on stdout. Credential fields render as their stored
environment reference, never a resolved secret.

## Example

```
$ sauron describe registry
source:         git@github.com:acme/artifacts.git
transport:      git
credentials:
  username:     ${env:ACME_USER}
  password:     ${env:ACME_TOKEN}
timeout:        30s
createdAt:      2026-06-21T07:30:00Z
lastUpdatedAt:  2026-06-21T07:30:00Z
```

## Exit codes

| Code | Condition |
|---|---|
| `0` | The registry detail was produced |
| `2` | Missing/invalid arguments or flags |
| `1` | No registry is configured, or `settings.yaml` is unreadable |

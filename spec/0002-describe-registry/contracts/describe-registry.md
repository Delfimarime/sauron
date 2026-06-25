# `describe registry` — command line

```
sauron describe registry [--fields <list>]
```

Show the configured registry's full detail.

## Flags

| Flag | Meaning |
|---|---|
| `--fields <list>` | Fields to display, in order; `uri` is always present and first. Valid: `uri`, `transport`, `ref`, `auth`, `tls`, `sshKey`, `timeout`, `creationTimestamp`, `lastUpdatedTimestamp` |

## Output

The registry's fields on stdout. Credential fields render as their stored
environment reference, never a resolved secret.

## Example

```
$ sauron describe registry
uri:                   git@github.com:acme/artifacts.git
transport:             git
auth:
  username:            ${env:ACME_USER}
  password:            ${env:ACME_TOKEN}
timeout:               30s
creationTimestamp:     2026-06-21T07:30:00Z
lastUpdatedTimestamp:  2026-06-21T07:30:00Z
```

## Exit codes

| Code | Condition |
|---|---|
| `0` | The registry detail was produced |
| `2` | Missing/invalid arguments or flags |
| `1` | No registry is configured, or `settings.yaml` is unreadable |

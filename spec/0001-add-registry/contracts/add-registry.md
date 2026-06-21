# `add registry` — command line

```
sauron add registry [--kind git|http|filesystem] [--ref <ref>] [--timeout <duration>]
                     [--username <value>] [--password <ref>] [--skip-tls-verify]
                     [--ca-cert <path>] [--client-cert <path>] [--client-key <path>]
                     [--ssh-key <path>]
                     <name> <uri>
```

Register an artifact source of any transport.

## Arguments

| Argument | Required | Meaning |
|---|---|---|
| `<name>` | yes | Unique, path-safe registry name; the namespacing segment in `sauron-<registry>-<name>` |
| `<uri>` | yes | The source location, interpreted per the transport |

## Flags

| Flag | Meaning |
|---|---|
| `--kind <kind>` | Transport: `git`, `http` (default), or `filesystem`; persisted as `spec.transport` |
| `--timeout <duration>` | Bound on the validation network operation (default `30s`) |
| `--username <value>` | Auth username; a literal value or an environment reference (`${env:VAR}`) |
| `--password <ref>` | Auth password/token, as an environment reference (`${env:VAR}`); a literal secret is a usage error |
| `--skip-tls-verify` | Skip TLS certificate verification (http/git) |
| `--ca-cert <path>` | CA certificate path (http/git) |
| `--client-cert <path>` | Client certificate path (http) |
| `--client-key <path>` | Client key path (http) |
| `--ssh-key <path>` | SSH private key path (git over SSH) |
| `--ref <ref>` | Git ref (branch, tag, or commit) to resolve artifacts from; persisted as `spec.ref`; git only |

Auth and TLS flags are accepted only for the transports that use them; the
applicable set per transport is fixed by the
[git](../capabilities/git.md), [http](../capabilities/http.md), and
[filesystem](../capabilities/filesystem.md) capabilities.

## Output

On success, one confirmation line on stdout naming the registered registry and its
transport. No output is written to stdout on failure.

## Example

```
$ sauron add registry --kind git acme git@github.com:acme/artifacts.git
registered registry "acme" (git)
```

## Exit codes

| Code | Condition |
|---|---|
| `0` | The registry was validated and persisted |
| `2` | Missing/invalid arguments or flags, or a non-path-safe name |
| `1` | A registry of that name exists, the source is unreachable, or it hosts no artifact |

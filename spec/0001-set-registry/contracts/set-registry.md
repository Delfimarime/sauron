# `set registry` — command line

```
sauron set registry [--transport git|http] [--revision <revision>] [--timeout <duration>]
                     [--username <value>] [--password <ref>] [--skip-tls-verify]
                     [--ca-cert <path>] [--client-cert <path>] [--client-key <path>]
                     [--ssh-key <path>]
                     <source>
```

Configure the single artifact source, of any transport, replacing any registry
already set.

## Arguments

| Argument | Required | Meaning |
|---|---|---|
| `<source>` | yes | The source location, interpreted per the transport |

## Flags

| Flag | Meaning |
|---|---|
| `--transport <kind>` | Transport: `git` or `http` (default); persisted as `spec.transport` |
| `--timeout <duration>` | Bound on the validation network operation (default `30s`) |
| `--username <value>` | Auth username; a literal value or an environment reference (`${env:VAR}`) |
| `--password <ref>` | Auth password/token, as an environment reference (`${env:VAR}`); a literal secret is a usage error |
| `--skip-tls-verify` | Skip TLS certificate verification (http/git) |
| `--ca-cert <path>` | CA certificate path (http/git) |
| `--client-cert <path>` | Client certificate path (http) |
| `--client-key <path>` | Client key path (http) |
| `--ssh-key <path>` | SSH private key path (git over SSH) |
| `--revision <revision>` | Git revision (branch, tag, or commit) to resolve artifacts from; persisted as `spec.revision`; git only |

Auth and TLS flags are accepted only for the transports that use them; the
applicable set per transport is fixed by the
[git](../capabilities/git.md) and [http](../capabilities/http.md) capabilities.

## Output

On success, one confirmation line on stdout naming the source and its
transport. No output is written to stdout on failure.

## Example

```
$ sauron set registry --transport git git@github.com:acme/artifacts.git
registry set to git@github.com:acme/artifacts.git (git)
```

## Exit codes

| Code | Condition |
|---|---|
| `0` | The registry was validated and persisted (creating or replacing) |
| `2` | Missing/invalid arguments or flags |
| `1` | The source is unreachable, it hosts no artifact, the `--revision` cannot be resolved, or persisting it fails (IO) |

# Contract: Command Line — Backend

Conventions: [CLI conventions](../../AUTHORING.md).

**Spec**: [Backend](../spec.md)

This feature owns two commands: `set backend` configures the singleton
backend, and `unset backend` tears it down.

## Synopsis

```
sauron set backend [--kind <http|filesystem|git>]
  [--username ${env:VAR}] [--password ${env:VAR}] [--timeout <duration>]
  <uri>

sauron unset backend [--keep-artifacts]
```

Command hierarchy: `sauron` (root) → `set` / `unset` (group) →
`backend` (subcommand). `--kind` defaults to `http`; `filesystem` and
`git` must be selected explicitly. The backend is a singleton — re-running
`set backend` overrides the previous configuration.

## Arguments

| Argument | Command | Required | Description |
|----------|---------|----------|-------------|
| `<uri>` | `set backend` | Yes | The persona-definition backend, by kind: an `http`/`https` URL ([http](../capabilities/http.md)), a directory path resolved to an absolute, symlink-resolved path before use ([filesystem](../capabilities/filesystem.md)), or an SSH-based git URI — scp-like or `ssh://` ([git](../capabilities/git.md)). Realizes [spec](../spec.md) FR-004, FR-013, FR-017. |

## Flags

A kind-scoped flag passed with a different kind is a usage error (exit `2`).

| Flag | Command | Scope | Required | Default | Values | Description |
|------|---------|-------|----------|---------|--------|-------------|
| `--kind` | `set backend` | all | No | `http` | `http`, `filesystem`, `git` | Backend kind; defaults to `http`, so `filesystem` and `git` must be given explicitly. Realizes [spec](../spec.md) FR-004, FR-014. |
| `--username` | `set backend` | http | No | — | `${env:VAR}` only | Basic-auth user, env reference only. Realizes [http](../capabilities/http.md) FR-004, FR-008, FR-009. |
| `--password` | `set backend` | http | No | — | `${env:VAR}` only | Basic-auth password, env reference only. Realizes [http](../capabilities/http.md) FR-004, FR-008, FR-009. |
| `--timeout` | `set backend` | http, git | No | `30s` | duration | Bounds network operations: the HTTP `HEAD` probe, `git ls-remote`. Realizes [http](../capabilities/http.md) FR-005, FR-010 and [git](../capabilities/git.md) FR-004, FR-007. |
| `--keep-artifacts` | `unset backend` | all | No | false | — | Remove the backend connection, the installed personas, and install records but leave the delivered artifacts in place. Realizes [spec](../spec.md) FR-010. |

## Output

- **`set backend` success**: a single confirmation line on stdout
  naming the configured backend. Realizes [spec](../spec.md) FR-007.
- **`unset backend` success**: a report of what was torn down — the
  backend connection, the installed personas, and install records; when artifacts
  are removed (no `--keep-artifacts`), they are listed in the shared plan/report
  format, grouped under `skills:` and `agents:` headings, one artifact per line
  prefixed `-`, followed by a summary count line. When nothing is configured, a
  line noting nothing was removed. Realizes [spec](../spec.md) FR-008, FR-009,
  FR-010, FR-012.
- **Failure**: a single human-readable message on stderr. No partial output.

## Exit codes

Exit-status meanings are owned by the [CLI conventions](../../AUTHORING.md);
this table refines which conditions map to each code.

| Code | Meaning | Realizes |
|------|---------|----------|
| `0` | Backend set; or torn down; or no backend was configured (idempotent teardown) | [spec](../spec.md) FR-005, FR-007, FR-008, FR-012 |
| `2` | Usage error — missing `<uri>`, unknown `--kind`, a raw (non-env) credential, a malformed `uri` for the kind, an invalid `--timeout`, a kind-scoped flag on a different kind | [spec](../spec.md) FR-013, FR-014, FR-015, FR-017, FR-019, FR-020; [http](../capabilities/http.md) FR-006, FR-008, FR-010; [git](../capabilities/git.md) FR-005, FR-007 |
| `1` | Runtime error — backend unreachable, an unset `${env:VAR}`, an unreadable `backend.yaml`/`personas.yaml`/`track.yaml`, or a failed artifact removal during teardown | [spec](../spec.md) FR-016, FR-018, FR-021; [http](../capabilities/http.md) FR-007; [filesystem](../capabilities/filesystem.md) FR-005; [git](../capabilities/git.md) FR-006 |

## Examples

```
# Set an http backend (kind defaults to http)
$ sauron set backend https://personas.example.com
Configured backend: https://personas.example.com

# Re-running overrides the previous backend (singleton upsert)
$ sauron set backend https://personas-v2.example.com
Configured backend: https://personas-v2.example.com

# http with basic auth via environment references
$ sauron set backend \
    --username '${env:PERSONAS_USER}' --password '${env:PERSONAS_PASS}' \
    https://secure-personas.example.com
Configured backend: https://secure-personas.example.com

# filesystem (explicit kind required)
$ sauron set backend --kind filesystem ./team-personas
Configured backend: /home/user/team-personas

# git over SSH (explicit kind required)
$ sauron set backend --kind git git@github.com:acme/personas.git
Configured backend: git@github.com:acme/personas.git

# Unknown kind (usage error, exit 2)
$ sauron set backend --kind ftp ftp://personas.example.com
Error: unknown --kind 'ftp' (expected http, filesystem, or git)

# Raw credential rejected (usage error, exit 2)
$ sauron set backend --username admin https://personas.example.com
Error: only the ${env:VAR} form is supported for credentials

# Unreachable backend (validation error, exit 1)
$ sauron set backend https://down.example.com
Error: backend cannot be reached: https://down.example.com

# Tear down everything (connection + installed personas + install records + artifacts)
$ sauron unset backend
skills:
- code-review
agents:
- software-engineer
Removed backend, installed personas, and install records; cleared 2 artifacts (1 skill, 1 agent).

# Tear down but keep delivered artifacts
$ sauron unset backend --keep-artifacts
Removed backend, installed personas, and install records; kept delivered artifacts.

# Nothing configured (idempotent, exit 0)
$ sauron unset backend
No backend configured; nothing to remove.
```

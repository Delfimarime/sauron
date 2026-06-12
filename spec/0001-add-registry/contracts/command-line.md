# Contract: Command Line — Add Registry

Conventions: [CLI conventions](../../AUTHORING.md).

**Spec**: [Add Registry](../spec.md)

Defines the command-line interface for registering a registry of any kind
(`http`, `filesystem`, or `git`) as a source of artifacts. This is the
user-facing contract only; kind-specific validation behavior is defined by the
[http](../capabilities/http.md), [filesystem](../capabilities/filesystem.md),
and [git](../capabilities/git.md) capabilities.

## Synopsis

```
sauron add registry [--kind <http|filesystem|git>] [--priority <n>]
  [--username ${env:VAR}] [--password ${env:VAR}] [--skip-tls-verify]
  [--ca-cert <path>] [--client-cert <path>] [--client-key <path>]
  [--ssh-key <path>] [--timeout <duration>]
  <name> <location>
```

Command hierarchy: `sauron` (root) → `add` (group) → `registry`
(subcommand). `--kind` defaults to `http`; `filesystem` and `git` must be
selected explicitly.

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `<name>` | Yes | Unique slug `^[a-z0-9]+(-[a-z0-9]+)*$`, unique across all kinds. Realizes [spec](../spec.md) FR-002, FR-004, FR-005, FR-012, FR-013, FR-016. |
| `<location>` | Yes | The artifact source, by kind: an `http`/`https` URL ([http](../capabilities/http.md)), a directory path resolved to an absolute, symlink-resolved path before use ([filesystem](../capabilities/filesystem.md)), or an SSH-based git URI — scp-like or `ssh://` ([git](../capabilities/git.md)). Realizes [spec](../spec.md) FR-004, FR-014, FR-018. |

## Flags

A kind-scoped flag passed with a different kind is a usage error (exit `2`).
Realizes [spec](../spec.md) FR-019.

| Flag | Scope | Required | Default | Values | Description |
|------|-------|----------|---------|--------|-------------|
| `--kind` | all | No | `http` | `http`, `filesystem`, `git` | Registry kind; defaults to `http`, so `filesystem` and `git` must be given explicitly. Realizes [spec](../spec.md) FR-004. |
| `--priority` | all | No | first repo `0`, else `max + 1` | non-negative int | Optional; unique across all kinds, lower = higher precedence. The first registry takes `0`; an omitted value appends at the end (`max + 1`). See the [unified priority model](../../AUTHORING.md#priority-model). Realizes [spec](../spec.md) FR-003, FR-004, FR-009, FR-010, FR-015, FR-017, FR-022. |
| `--username` | http | No | — | `${env:VAR}` only | Basic-auth user, env reference only. Realizes [http](../capabilities/http.md) FR-004, FR-009, FR-010. |
| `--password` | http | No | — | `${env:VAR}` only | Basic-auth password, env reference only. Realizes [http](../capabilities/http.md) FR-004, FR-009, FR-010. |
| `--skip-tls-verify` | http | No | false | — | Skip server cert verification. Realizes [http](../capabilities/http.md) FR-003. |
| `--ca-cert` | http | No | — | path | CA bundle to trust the server. Realizes [http](../capabilities/http.md) FR-003, FR-011. |
| `--client-cert` | http | No | — | path | Client cert for mutual TLS; requires `--client-key`. Realizes [http](../capabilities/http.md) FR-003, FR-008, FR-011. |
| `--client-key` | http | No | — | path | Client key for mutual TLS; requires `--client-cert`. Realizes [http](../capabilities/http.md) FR-003, FR-008, FR-011. |
| `--ssh-key` | git | No | system SSH | path | Private key to authenticate; omitted = system SSH credentials (agent, `~/.ssh/config`, default keys). Realizes [git](../capabilities/git.md) FR-007, FR-009. |
| `--timeout` | http, git | No | `30s` | duration | Bounds network operations: the HTTP `HEAD` probe and fetches, `git ls-remote`. Realizes [http](../capabilities/http.md) FR-005, FR-012 and [git](../capabilities/git.md) FR-004, FR-008. |

## Output

- **Success**: a single confirmation line on stdout naming the registered
  registry name and location. Realizes [spec](../spec.md) FR-008.
- **Failure**: a single human-readable message on stderr. No partial output,
  no stack traces. Realizes [spec](../spec.md) FR-021.

## Exit codes

Exit-status meanings are owned by the [CLI conventions](../../AUTHORING.md);
this table refines which conditions map to each code.

| Code | Meaning | Realizes |
|------|---------|----------|
| `0` | Registry registered | [spec](../spec.md) FR-006, FR-008 |
| `2` | Usage error — missing `<name>`/`<location>`, invalid name, a `--priority` that is not a non-negative integer, a non-`0` `--priority` for the first registry, unknown `--kind`, invalid `--timeout`, a kind-scoped flag on a different kind; **http**: invalid (non-`http`/`https`) URL, raw (non-env) credential, unpaired `--client-cert`/`--client-key`; **git**: non-SSH or malformed git URI | [spec](../spec.md) FR-012, FR-013, FR-014, FR-015, FR-019, FR-020, FR-022; [http](../capabilities/http.md) FR-006, FR-008, FR-009, FR-012; [git](../capabilities/git.md) FR-005, FR-008 |
| `1` | Validation error — duplicate name, duplicate priority; **http**: server unreachable (connection, TLS, or non-success `HEAD`), unset `${env:VAR}`, unreadable CA/cert/key file; **filesystem**: directory not accessible, no artifacts under `.skills/` or `.agents/`; **git**: remote unreachable or authentication failed, unreadable `--ssh-key` file | [spec](../spec.md) FR-016, FR-017; [http](../capabilities/http.md) FR-007, FR-010, FR-011; [filesystem](../capabilities/filesystem.md) FR-005, FR-006; [git](../capabilities/git.md) FR-006, FR-007 |

## Examples

```
# http (kind defaults to http)
$ sauron add registry --priority 1 team-http https://skills.example.com
Registered registry 'team-http': https://skills.example.com

# First registry, --priority omitted (assigned 0)
$ sauron add registry team-base https://base.example.com
Registered registry 'team-base': https://base.example.com

# Later registry, --priority omitted (appended at the end, max + 1)
$ sauron add registry --kind filesystem team-extra ./team-extra
Registered registry 'team-extra': /home/user/team-extra

# http with basic auth via environment references
$ sauron add registry --priority 2 \
    --username '${env:SKILLS_USER}' --password '${env:SKILLS_PASS}' \
    team-secure https://secure.example.com
Registered registry 'team-secure': https://secure.example.com

# filesystem (explicit kind required)
$ sauron add registry --kind filesystem --priority 3 team-skills ./team-skills
Registered registry 'team-skills': /home/user/team-skills

# git over SSH (explicit kind required)
$ sauron add registry --kind git --priority 4 team-git git@github.com:acme/skills.git
Registered registry 'team-git': git@github.com:acme/skills.git

# Kind-scoped flag on the wrong kind (usage error, exit 2)
$ sauron add registry --kind filesystem --priority 5 --ssh-key ~/.ssh/id_ed25519 t ./t
Error: --ssh-key applies only to git registries

# Non-SSH git URI rejected (usage error, exit 2)
$ sauron add registry --kind git --priority 6 web https://github.com/acme/skills.git
Error: an SSH-based git URI is required (http(s)://, git://, file:// are not supported)

# Unreachable server (validation error, exit 1)
$ sauron add registry --priority 7 down https://down.example.com
Error: registry cannot be reached: https://down.example.com
```

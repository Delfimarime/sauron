# Contract: Command Line — Add HTTP Server Repository

**Spec**: `../spec.md` (Add HTTP Server Repository)
**Status**: Draft

Defines the command-line interface for registering an HTTP(S) server as a repository. This is the user-facing contract only. `http` is the default kind; other kinds are covered by their own features.

## Synopsis

```
sauron add repository [--kind http] --priority <n> \
  [--username ${env:VAR}] [--password ${env:VAR}] [--skip-tls-verify] \
  [--ca-cert <path>] [--client-cert <path>] [--client-key <path>] \
  [--timeout <duration>] \
  <name> <url>
```

Command hierarchy: `sauron` (root) → `add` (group) → `repository` (subcommand).

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `<name>` | Yes | Unique slug `^[a-z0-9]+(-[a-z0-9]+)*$`, unique across all kinds. Realizes FR-004, FR-008, FR-015, FR-016, FR-018. |
| `<url>` | Yes | http/https URL of the server. Realizes FR-004, FR-005, FR-012, FR-013. |

## Flags

| Flag | Required | Default | Values | Description |
|------|----------|---------|--------|-------------|
| `--kind` | No | http | `http` | Repository kind; defaults to http. Realizes FR-004. |
| `--priority` | Yes | — | positive int | Unique, lower = higher precedence. Realizes FR-004, FR-017, FR-019. |
| `--username` | No | — | `${env:VAR}` only | Basic-auth user, env reference only. Realizes FR-007, FR-021, FR-022. |
| `--password` | No | — | `${env:VAR}` only | Basic-auth password, env reference only. Realizes FR-007, FR-021, FR-022. |
| `--skip-tls-verify` | No | false | — | Skip server cert verification. Realizes FR-006. |
| `--ca-cert` | No | — | path | CA bundle to trust the server. Realizes FR-006, FR-023. |
| `--client-cert` | No | — | path | Client cert for mutual TLS; requires `--client-key`. Realizes FR-006, FR-020, FR-023. |
| `--client-key` | No | — | path | Client key for mutual TLS; requires `--client-cert`. Realizes FR-006, FR-020, FR-023. |
| `--timeout` | No | 30s | duration | Bounds the HEAD probe and fetches. Realizes FR-006, FR-025, FR-026. |

## Output

- **Success**: a single confirmation line on stdout naming the registered name and URL.
- **Failure**: a single human-readable message on stderr. No partial output, no stack traces.

## Exit codes

| Code | Meaning | Realizes |
|------|---------|----------|
| `0` | Repository registered | FR-009, FR-010 |
| `2` | Usage error — missing `<name>`/`<url>`, invalid name, invalid URL, missing/invalid `--priority`, invalid `--timeout`, unsupported `--kind`, raw (non-env) credential, unpaired `--client-cert`/`--client-key`, or an http-only flag on another kind | FR-004, FR-012, FR-013, FR-015, FR-016, FR-017, FR-020, FR-021, FR-024, FR-026 |
| `1` | Validation error — server unreachable, duplicate name, duplicate priority, unset `${env:VAR}`, or unreadable cert/key file | FR-014, FR-018, FR-019, FR-022, FR-023 |

## Examples

```
# Success (kind defaults to http)
$ sauron add repository --priority 1 team-http https://skills.example.com
Registered repository 'team-http': https://skills.example.com

# Basic auth via environment references
$ sauron add repository --priority 2 \
    --username '${env:SKILLS_USER}' --password '${env:SKILLS_PASS}' \
    team-secure https://secure.example.com
Registered repository 'team-secure': https://secure.example.com

# Mutual TLS
$ sauron add repository --priority 3 \
    --ca-cert ./ca.pem --client-cert ./client.pem --client-key ./client.key \
    team-mtls https://mtls.example.com
Registered repository 'team-mtls': https://mtls.example.com

# Raw credential rejected (usage error, exit 2)
$ sauron add repository --priority 4 --username admin t https://x.example.com
Error: --username supports only the ${env:VAR} form

# Invalid URL (usage error, exit 2)
$ sauron add repository --priority 5 bad ftp://nope.example.com
Error: invalid URL; must be http:// or https://

# Custom timeout
$ sauron add repository --priority 6 --timeout 10s slow https://slow.example.com
Registered repository 'slow': https://slow.example.com

# Unreachable server (validation error, exit 1)
$ sauron add repository --priority 7 down https://down.example.com
Error: repository cannot be reached: https://down.example.com

# Unset env reference (validation error, exit 1)
$ sauron add repository --priority 7 --username '${env:MISSING}' t https://x.example.com
Error: environment variable 'MISSING' is not set
```

# Contract: Command Line — Add Git Repository

**Spec**: `../spec.md` (Add Git Repository)

**Status**: Draft

Defines the command-line interface for registering an SSH git remote as a repository. This is the user-facing contract only. `git` must be selected explicitly via `--kind git`; `http` is the default kind, and other kinds are covered by their own features.

## Synopsis

```
sauron add repository --kind git --priority <n> \
  [--ssh-key <path>] [--timeout <duration>] \
  <name> <uri>
```

Command hierarchy: `sauron` (root) → `add` (group) → `repository` (subcommand).

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `<name>` | Yes | Unique slug `^[a-z0-9]+(-[a-z0-9]+)*$`, unique across all kinds. Realizes FR-004, FR-008, FR-016, FR-017, FR-019. |
| `<uri>` | Yes | SSH-based git URI (scp-like or `ssh://`). Realizes FR-004, FR-005, FR-013, FR-014. |

## Flags

| Flag | Required | Default | Values | Description |
|------|----------|---------|--------|-------------|
| `--kind` | Yes | — | `git` | Selects the git kind; required because http is the default. Realizes FR-004, FR-023. |
| `--priority` | Yes | — | positive int | Unique, lower = higher precedence. Realizes FR-004, FR-018, FR-020. |
| `--ssh-key` | No | system SSH | path | Private key to authenticate; omitted = system SSH credentials. Realizes FR-007, FR-021, FR-024. |
| `--timeout` | No | 30s | duration | Bounds network operations (e.g. `git ls-remote`). Realizes FR-006, FR-011, FR-022. |

## Output

- **Success**: a single confirmation line on stdout naming the registered name and URI.
- **Failure**: a single human-readable message on stderr. No partial output, no stack traces.

## Exit codes

| Code | Meaning | Realizes |
|------|---------|----------|
| `0` | Repository registered | FR-009, FR-010 |
| `2` | Usage error — missing `<name>`/`<uri>`, invalid name, invalid SSH URI, missing/invalid `--priority`, missing/unsupported `--kind`, invalid `--timeout`, or `--ssh-key` on a non-git kind | FR-004, FR-013, FR-014, FR-016, FR-017, FR-018, FR-022, FR-023, FR-024 |
| `1` | Validation error — remote unreachable / auth failed, duplicate name, duplicate priority, or unreadable `--ssh-key` file | FR-015, FR-019, FR-020, FR-021 |

## Examples

```
# Success (system SSH credentials)
$ sauron add repository --kind git --priority 1 team-git git@github.com:acme/skills.git
Registered repository 'team-git': git@github.com:acme/skills.git

# Explicit key and ssh:// URL
$ sauron add repository --kind git --priority 2 \
    --ssh-key ~/.ssh/deploy_ed25519 \
    team-deploy ssh://git@github.com/acme/agents.git
Registered repository 'team-deploy': ssh://git@github.com/acme/agents.git

# Custom timeout
$ sauron add repository --kind git --priority 3 --timeout 10s \
    team-slow git@git.example.com:team/repo.git
Registered repository 'team-slow': git@git.example.com:team/repo.git

# Non-SSH URI rejected (usage error, exit 2)
$ sauron add repository --kind git --priority 4 web https://github.com/acme/skills.git
Error: an SSH-based git URI is required (http(s)://, git://, file:// are not supported)

# Unreachable / auth failed (validation error, exit 1)
$ sauron add repository --kind git --priority 5 down git@github.com:acme/missing.git
Error: repository cannot be reached: git@github.com:acme/missing.git
```

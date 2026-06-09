# Contract: Command Line — Add Filesystem Repository

**Spec**: `../spec.md` (Add Filesystem Repository)
**Status**: Draft

Defines the command-line interface for registering a filesystem directory as a repository. This is the user-facing contract only.

## Synopsis

```
sauron add repository --kind filesystem --priority <n> <name> <path>
```

Command hierarchy: `sauron` (root) → `add` (group) → `repository` (subcommand).

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `<name>` | Yes | Unique slug identifying the repository. Pattern `^[a-z0-9]+(-[a-z0-9]+)*$`. Must be unique across all repositories regardless of kind. Realizes FR-002, FR-005, FR-017, FR-018, FR-021. |
| `<path>` | Yes | Directory to register. Resolved to an absolute, symlink-resolved path before use. Realizes FR-002, FR-009. |

## Flags

| Flag | Required | Values | Description |
|------|----------|--------|-------------|
| `--kind` | Yes | `filesystem` | Repository kind. Only `filesystem` is accepted in this feature; any other value is a usage error. Realizes FR-002, FR-010. |
| `--priority` | Yes | positive integer | Repository priority. Lower is higher precedence. Must be unique across all repositories regardless of kind. Realizes FR-002, FR-019, FR-022. |

## Output

- **Success**: a single confirmation line on stdout naming the registered repository name and absolute path. Realizes FR-007.
- **Failure**: a single human-readable message on stderr. No partial output, no stack traces.

## Exit codes

| Code | Meaning | Realizes |
|------|---------|----------|
| `0` | Repository registered | FR-006, FR-007 |
| `2` | Usage error — missing `<name>`, missing `<path>`, missing/unsupported `--kind`, missing/invalid `--priority`, or invalid name format | FR-002, FR-009, FR-010, FR-017, FR-018, FR-019 |
| `1` | Validation error — directory not accessible, no artifacts found, duplicate name, or duplicate priority | FR-011, FR-012, FR-021, FR-022 |

## Examples

```
# Success
$ sauron add repository --kind filesystem --priority 1 team-skills ./team-skills
Registered repository 'team-skills': /home/user/team-skills

# Missing path (usage error, exit 2)
$ sauron add repository --kind filesystem --priority 1 team-skills
Error: a directory path is required

# Unsupported kind (usage error, exit 2)
$ sauron add repository --kind something-else --priority 1 team-skills ./team-skills
Error: 'filesystem' is the only supported kind

# Invalid name (usage error, exit 2)
$ sauron add repository --kind filesystem --priority 1 "Team Skills" ./team-skills
Error: invalid name; use lowercase letters, digits, and hyphens

# Not accessible (validation error, exit 1)
$ sauron add repository --kind filesystem --priority 1 team-skills ./missing
Error: repository cannot be accessed: ./missing

# No artifacts (validation error, exit 1)
$ sauron add repository --kind filesystem --priority 1 team-skills ./empty
Error: repository has no skills or agents

# Duplicate name (validation error, exit 1)
$ sauron add repository --kind filesystem --priority 9 team-skills ./other
Error: a repository named 'team-skills' already exists

# Duplicate priority (validation error, exit 1)
$ sauron add repository --kind filesystem --priority 1 other-skills ./other
Error: priority 1 is already in use
```

# Contract: Command Line — Describe Registry

Conventions: [CLI conventions](../../AUTHORING.md).

**Spec**: [Describe Registry](../spec.md)

Defines the command-line interface for describing one registered registry. This
is the user-facing contract only. Describing is read-only and offline: it reads
the named registry from the settings and contacts no external resource.

## Synopsis

```
sauron describe registry <name> [--fields <list>]
```

Command hierarchy: `sauron` (root) → `describe` (group) → `registry`
(subcommand).

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `<name>` | Yes | Name of the registry to describe. Realizes [spec](../spec.md) FR-001, FR-003, FR-008, FR-009. |

## Flags

| Flag | Required | Default | Values | Description |
|------|----------|---------|--------|-------------|
| `--fields` | No | all fields | comma-separated `name`, `kind`, `location`, `priority`, `auth`, `tls`, `timeout`, `ssh-key` | Fields to display, in order; `name` is always present and first. Realizes [spec](../spec.md) FR-013, FR-012. |

## Output

- **Success**: the registry's fields on stdout, one per line as `field: value`,
  with `name` first. Fields that do not apply to the resolved kind, or that have
  no value, are shown with an empty value. `auth` shows the credential's
  `${env:VAR}` reference; a resolved secret is rendered `REDACTED`. Realizes
  [spec](../spec.md) FR-003, FR-005, FR-006, FR-007.
- **Failure**: a single human-readable message on stderr. Realizes
  [spec](../spec.md) FR-009, FR-011.

## Exit codes

Exit-status meanings are owned by the [CLI conventions](../../AUTHORING.md);
this table refines which conditions map to each code.

| Code | Meaning | Realizes |
|------|---------|----------|
| `0` | The registry was found and described | [spec](../spec.md) FR-003 |
| `2` | Usage error — `<name>` missing, or an unknown `--fields` value | [spec](../spec.md) FR-008, FR-012 |
| `1` | Runtime error — the registry was not found, or the settings cannot be read or parsed | [spec](../spec.md) FR-009, FR-010, FR-011 |

## Examples

```
# Describe an http registry (password shown as its env reference; resolved secret REDACTED)
$ sauron describe registry team-secure
name: team-secure
kind: http
location: https://secure.example.com
priority: 2
auth: ${env:SKILLS_PASS}
tls: caCert=/home/user/.sauron/ca.pem
timeout: 30s
ssh-key:

# A subset of fields, in the given order (name always first)
$ sauron describe registry team-secure --fields location,priority
name: team-secure
location: https://secure.example.com
priority: 2

# Registry not found (runtime error, exit 1)
$ sauron describe registry nope
Error: registry "nope" not found

# Missing name (usage error, exit 2)
$ sauron describe registry
Error: a registry name is required

# Unknown field (usage error, exit 2)
$ sauron describe registry team-secure --fields color
Error: --fields must name only: name, kind, location, priority, auth, tls, timeout, ssh-key
```

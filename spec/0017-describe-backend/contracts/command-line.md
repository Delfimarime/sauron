# Contract: Command Line — Describe Backend

Conventions: [CLI conventions](../../AUTHORING.md).

**Spec**: [Describe Backend](../spec.md)

Defines the command-line interface for describing the singleton backend. This is
the user-facing contract only. Describing is read-only and offline: it reads the
singleton backend connection from
[backend.yaml](../../contracts/configuration.md#backendyaml) and contacts no
external resource. The command takes no name, because there is exactly one
backend per instance.

## Synopsis

```
sauron describe backend [--fields <list>]
```

Command hierarchy: `sauron` (root) → `describe` (group) → `backend`
(subcommand).

## Arguments

This command takes no positional arguments. Supplying a name is a usage error.
Realizes [spec](../spec.md) FR-001, FR-009.

## Flags

| Flag | Required | Default | Values | Description |
|------|----------|---------|--------|-------------|
| `--fields` | No | all fields | comma-separated `uri`, `kind`, `auth`, `timeout`, `last-synced` | Fields to display, in order; `uri` is always present and first. Realizes [spec](../spec.md) FR-013, FR-010. |

## Output

- **Success (configured)**: the backend's fields on stdout, one per line as
  `field: value`, with `uri` first. Fields that have no value are shown with an
  empty value. `auth` is rendered `REDACTED`; `last-synced` is the
  `last_synced_at` timestamp. Realizes [spec](../spec.md) FR-003, FR-005, FR-006,
  FR-007.
- **Success (none configured)**: a single line on stdout reporting that no backend
  is configured; exit 0. Realizes [spec](../spec.md) FR-008.
- **Failure**: a single human-readable message on stderr. Realizes
  [spec](../spec.md) FR-011, FR-012.

## Exit codes

Exit-status meanings are owned by the [CLI conventions](../../AUTHORING.md);
this table refines which conditions map to each code.

| Code | Meaning | Realizes |
|------|---------|----------|
| `0` | The backend was described, or no backend is configured (idempotent no-op) | [spec](../spec.md) FR-003, FR-008 |
| `2` | Usage error — a name was supplied, or an unknown `--fields` value | [spec](../spec.md) FR-009, FR-010 |
| `1` | Runtime error — the backend connection cannot be read or parsed | [spec](../spec.md) FR-011 |

## Examples

```
# Describe the configured backend (auth always REDACTED)
$ sauron describe backend
uri: https://secure-personas.example.com
kind: http
auth: REDACTED
timeout: 30s
last-synced: 2026-06-12T09:30:00Z

# A subset of fields, in the given order (uri always first)
$ sauron describe backend --fields kind,last-synced
uri: https://secure-personas.example.com
kind: http
last-synced: 2026-06-12T09:30:00Z

# No backend configured (idempotent no-op, exit 0)
$ sauron describe backend
No backend configured.

# A name was supplied (usage error, exit 2)
$ sauron describe backend my-backend
Error: describe backend takes no name

# Unknown field (usage error, exit 2)
$ sauron describe backend --fields color
Error: --fields must name only: uri, kind, auth, timeout, last-synced
```

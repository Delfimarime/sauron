# Contract: Command Line — Sync Personas

Conventions: [CLI conventions](../../AUTHORING.md).

**Spec**: [Sync Personas](../spec.md)

Defines the command-line interface for refreshing the local catalog of persona
definitions from the configured backend. This is the user-facing
contract only. These commands pull persona *definitions*; they do not deliver
artifacts — artifact delivery is [sync](../../0006-sync-artifacts/spec.md), an independent
command.

## Synopsis

```
sauron sync personas [--force]
sauron sync persona <name> [--force]
```

Command hierarchy: `sauron` (root) → `sync` (command) → `personas` / `persona`
(subcommand).

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `<name>` | Yes (for `sync persona`) | The persona definition to refresh, by name. Realizes [spec](../spec.md) FR-007, FR-013. |

## Flags

| Flag | Required | Default | Values | Description |
|------|----------|---------|--------|-------------|
| `--force` | No | false | — | Re-pull authoritatively, ignoring the "unchanged" short-circuit, and hard-reconcile by uninstalling installed personas no longer present in the registry. Realizes [spec](../spec.md) FR-006, FR-008. |

## Output

- **Catalog changes** (always printed): the catalog entries added, updated, and
  removed; when nothing changed, an up-to-date message.
- **Kept installs**: without `--force`, installed personas that have vanished
  upstream are reported and kept; with `--force`, they are uninstalled and
  reported as removed.
- **Failure**: a single human-readable message on stderr.

## Exit codes

Exit-status meanings are owned by the [CLI conventions](../../AUTHORING.md);
this table refines which conditions map to each code.

| Code | Meaning | Realizes |
|------|---------|----------|
| `0` | Catalog refreshed, or already up to date | [spec](../spec.md) FR-002, FR-004, FR-007, FR-009 |
| `2` | Usage error — `sync persona` with no name, or a malformed flag | — |
| `1` | No backend configured; the registry unreachable; the settings or the catalog unreadable; or `sync persona` names a persona the registry does not offer | [spec](../spec.md) FR-010, FR-011, FR-012, FR-013 |

## Examples

```
# Refresh the whole catalog
$ sauron sync personas
added:
+ data-engineer
updated:
+ backend-developer
Catalog refreshed: 1 added, 1 updated, 0 removed.

# Up to date (exit 0)
$ sauron sync personas
Catalog is already up to date.

# An installed persona vanished upstream — reported, kept (exit 0)
$ sauron sync personas
removed:
- legacy-ops (still installed; kept — use --force to uninstall)
Catalog refreshed: 0 added, 0 updated, 1 removed.

# Authoritative re-pull — hard-reconcile installed personas
$ sauron sync personas --force
removed:
- legacy-ops (uninstalled)
Catalog refreshed: 0 added, 0 updated, 1 removed.

# Refresh one persona
$ sauron sync persona backend-developer
updated:
+ backend-developer
Catalog refreshed: 0 added, 1 updated, 0 removed.

# Unknown persona (exit 1)
$ sauron sync persona ghost
Error: persona 'ghost' is not offered by the backend

# No registry configured (exit 1)
$ sauron sync personas
Error: no backend configured; set a backend first
```

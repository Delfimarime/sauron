# Contract: Command Line — Sync Personas

Conventions: [CLI conventions](../../AUTHORING.md).

**Spec**: [Sync Personas](../spec.md)

Defines the command-line interface for refreshing the stored definitions of the
installed personas from the configured backend. This is the user-facing contract
only. These commands refresh persona *definitions*; they do not deliver artifacts
— artifact delivery is [sync](../../0006-sync-artifacts/spec.md), an independent
command. There is no persisted catalog; the *available* personas are a
[live view](../../contracts/configuration.md#live-persona-view) assembled at
command time.

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
| `<name>` | Yes (for `sync persona`) | The installed persona whose definition to refresh, by name. Realizes [spec](../spec.md) FR-007, FR-013. |

## Flags

| Flag | Required | Default | Values | Description |
|------|----------|---------|--------|-------------|
| `--force` | No | false | — | Re-pull definitions authoritatively, ignoring the "unchanged" short-circuit, and hard-reconcile by uninstalling installed personas the backend no longer offers. Realizes [spec](../spec.md) FR-006, FR-008. |

## Output

- **Refresh report** (always printed): the installed personas whose stored
  definitions were refreshed and those left unchanged; when nothing changed, an
  up-to-date message.
- **Kept installs**: without `--force`, installed personas the backend no longer
  offers are reported and kept; with `--force`, they are uninstalled and reported
  as removed.
- **Failure**: a single human-readable message on stderr.

## Exit codes

Exit-status meanings are owned by the [CLI conventions](../../AUTHORING.md);
this table refines which conditions map to each code.

| Code | Meaning | Realizes |
|------|---------|----------|
| `0` | Installed personas refreshed, or already up to date | [spec](../spec.md) FR-002, FR-004, FR-007, FR-009 |
| `2` | Usage error — `sync persona` with no name, or a malformed flag | — |
| `1` | No backend configured; the backend unreachable; `personas.yaml` or `backend.yaml` unreadable; or `sync persona` names a persona that is not installed | [spec](../spec.md) FR-010, FR-011, FR-012, FR-013 |

## Examples

```
# Refresh every installed persona's stored definition
$ sauron sync personas
refreshed:
+ backend-developer
+ data-engineer
Personas refreshed: 2 refreshed, 0 unchanged.

# Up to date (exit 0)
$ sauron sync personas
Installed personas are already up to date.

# An installed persona vanished upstream — reported, kept (exit 0)
$ sauron sync personas
- legacy-ops (no longer offered; kept — use --force to uninstall)
Personas refreshed: 0 refreshed, 1 unchanged.

# Authoritative re-pull — hard-reconcile installed personas
$ sauron sync personas --force
- legacy-ops (uninstalled)
Personas refreshed: 0 refreshed, 1 removed.

# Refresh one installed persona
$ sauron sync persona backend-developer
refreshed:
+ backend-developer
Personas refreshed: 1 refreshed, 0 unchanged.

# Persona not installed (exit 1)
$ sauron sync persona ghost
Error: persona 'ghost' is not installed

# No backend configured (exit 1)
$ sauron sync personas
Error: no backend configured; set a backend first
```

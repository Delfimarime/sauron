# Command Line Interface Reference

The compiled reference for the `sauron` CLI: every command with its synopsis,
intent, and key flags, each linking to the feature contract that owns its full
behavior. Every command obeys the same command grammar, shared flags,
exit-status semantics, and output discipline.

## add registry

```
sauron add registry [--kind <kind>] [--priority <n>] [kind-scoped flags] <name> <uri>
```

Register an artifact source of any kind.

- Key flags: `--kind` (default `http`), `--priority` (optional; first repo `0`, else `max + 1`), `--timeout`
  (http/git); kind-scoped auth/TLS flags (`--username`/`--password`,
  `--skip-tls-verify`, `--ca-cert`, `--client-cert`/`--client-key`, `--ssh-key`).
- Full contract → [0001-add-registry](../0001-add-registry/contracts/command-line.md).

## list registries

```
sauron list registries [--search <term>] [--fields <list>] [--sort <name|priority|kind>] [--order <asc|desc>]
```

Review configured sources.

- Key flags: `--search`, `--fields`, `--sort`, `--order`.
- Full contract → [0002-list-registries](../0002-list-registries/contracts/command-line.md).

## delete registry

```
sauron delete registry <name>
```

Unregister a source, keeping installed artifacts.

- Key flags: none.
- Full contract → [0003-delete-registry](../0003-delete-registry/contracts/command-line.md).

## prune

```
sauron prune (artifacts|skills|agents) [--dry-run]
```

Remove artifacts orphaned by unregistered registries.

- Required noun: `artifacts` (both), `skills`, or `agents`. Key flags: `--dry-run`.
- Full contract → [0004-prune](../0004-prune/contracts/command-line.md).

## list personas

```
sauron list personas [--search <term>] [--tag <tag>]... [--installed <true|false>] [--fields <list>] [--sort <name|installed|priority|last-updated|last-synced>] [--order <asc|desc>]
```

Review the available personas (installed plus those the backend offers live) and which are installed.

- Key flags: `--search`, `--tag` (repeatable), `--installed`, `--fields`, `--sort`, `--order`.
- Full contract → [0005-list-personas](../0005-list-personas/contracts/command-line.md).

## sync artifacts

```
sauron sync artifacts [--persona <name>] [--dry-run]
```

Reconcile the provider with registries and installed personas.

- Key flags: `--persona`, `--dry-run`.
- Full contract → [0006-sync-artifacts](../0006-sync-artifacts/contracts/command-line.md).

## set priority persona

```
sauron set priority persona <name> <value>
```

Adjust an installed persona's precedence.

- Key flags: none.
- Full contract → [0007-set-persona-priority](../0007-set-persona-priority/contracts/command-line.md).

## set priority registry

```
sauron set priority registry <name> <value>
```

Reorder registry precedence for conflict resolution.

- Key flags: none.
- Full contract → [0008-set-registry-priority](../0008-set-registry-priority/contracts/command-line.md).

## set provider

```
sauron set provider <claude|zencoder> [--copy-only]
```

Choose the provider destination.

- Key flags: `--copy-only`.
- Full contract → [0009-set-provider](../0009-set-provider/contracts/command-line.md).

## delete artifacts

```
sauron delete (artifacts|skills|agents) [--persona <name>] [--dry-run]
```

Remove all tracked artifacts Sauron installed (optionally scoped to one persona);
unlike `prune`, which removes only orphans from unregistered registries, this
removes everything in scope.

- Required noun: `artifacts` (both), `skills`, or `agents`. Key flags: `--persona`, `--dry-run`.
- Full contract → [0010-delete-artifacts](../0010-delete-artifacts/contracts/command-line.md).

## schedule sync artifacts

```
sauron schedule sync artifacts <expression>
sauron unschedule sync artifacts
sauron unschedule sync
```

Schedule `sauron sync artifacts` via the OS crontab; `unschedule sync artifacts`
removes it, and `unschedule sync` (no operation) removes every managed sync schedule.

- Required arg: `<expression>` (cron) for `schedule`.
- Full contract → [0011-schedule-sync](../0011-schedule-sync/contracts/command-line.md).

## set backend

```
sauron set backend [--kind <http|filesystem|git>] [--username ${env:VAR}] [--password ${env:VAR}] [--timeout <duration>] <uri>
```

Configure the singleton backend that owns persona definitions (upsert).

- Key flags: `--kind` (default `http`), `--username`/`--password` (env refs only), `--timeout`.
- Full contract → [0012-backend](../0012-backend/contracts/command-line.md).

## unset backend

```
sauron unset backend [--keep-artifacts]
```

Tear down the backend (cascades to installed personas, installs, and artifacts).

- Key flags: `--keep-artifacts` (preserve delivered artifacts).
- Full contract → [0012-backend](../0012-backend/contracts/command-line.md).

## sync personas

```
sauron sync personas [--force]
sauron sync persona <name> [--force]
```

Refresh the definitions of the installed personas from the backend.

- Key flags: `--force` (authoritative re-pull + hard-reconcile).
- Full contract → [0013-sync-personas](../0013-sync-personas/contracts/command-line.md).

## set persona

```
sauron set persona <name>...
```

Declare the installed persona set (full-replace; priority = argument order).

- Key flags: none.
- Full contract → [0014-select-personas](../0014-select-personas/contracts/command-line.md).

## unset persona

```
sauron unset persona [<name>...]
```

Uninstall named personas, or all when no name is given.

- Key flags: none.
- Full contract → [0014-select-personas](../0014-select-personas/contracts/command-line.md).

## describe registry

```
sauron describe registry <name> [--fields <list>]
```

Show one registry's full detail.

- Key flags: `--fields`.
- Full contract → [0015-describe-registry](../0015-describe-registry/contracts/command-line.md).

## describe persona

```
sauron describe persona <name> [--fields <list>]
```

Show one persona's full detail (installed or available).

- Key flags: `--fields`.
- Full contract → [0016-describe-persona](../0016-describe-persona/contracts/command-line.md).

## describe backend

```
sauron describe backend [--fields <list>]
```

Show the singleton backend's detail.

- Key flags: `--fields`.
- Full contract → [0017-describe-backend](../0017-describe-backend/contracts/command-line.md).

## describe provider

```
sauron describe provider [--fields <list>]
```

Show the active provider's detail.

- Key flags: `--fields`.
- Full contract → [0018-describe-provider](../0018-describe-provider/contracts/command-line.md).

## schedule sync personas

```
sauron schedule sync personas <expression>
sauron unschedule sync personas
```

Schedule `sauron sync personas` via the OS crontab; `unschedule sync personas`
removes it. (Remove every schedule at once with `unschedule sync`.)

- Required arg: `<expression>` (cron) for `schedule`.
- Full contract → [0019-schedule-sync-personas](../0019-schedule-sync-personas/contracts/command-line.md).

## pin skill / pin agent

```
sauron pin skill <name> <registry> [--reconcile]
sauron pin agent <name> <registry> [--reconcile]
sauron unpin skill <name> [--reconcile] [--dry-run]
sauron unpin agent <name> [--reconcile] [--dry-run]
```

Bind an artifact to a registry, overriding priority; `unpin` removes the binding.

- Key flags: `--reconcile` (apply now via a scoped sync), `--dry-run` (`unpin` preview).
- Full contract → [0020-pin-artifact](../0020-pin-artifact/contracts/command-line.md).

## list skills / list agents

```
sauron list skills [--available] [--registry <name>] [--search <term>] [--fields <list>] [--sort <name|registry|type>] [--order <asc|desc>]
sauron list agents [--available] [--registry <name>] [--search <term>] [--fields <list>] [--sort <name|registry|type>] [--order <asc|desc>]
```

List managed skills/agents; `--available` shows a registry's offerings or the resolved catalog.

- Key flags: `--available`, `--registry`, `--search`, `--fields` (incl. `source`, `pinned`), `--sort`, `--order`.
- Full contract → [0021-list-artifacts](../0021-list-artifacts/contracts/command-line.md).

## describe skill / describe agent

```
sauron describe skill <name> [--fields <list>]
sauron describe agent <name> [--fields <list>]
```

Show one managed skill's or agent's detail.

- Key flags: `--fields` (incl. `source`, `pinned`).
- Full contract → [0022-describe-artifact](../0022-describe-artifact/contracts/command-line.md).

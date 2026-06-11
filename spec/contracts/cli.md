# Command Line Interface Reference

The compiled reference for the `sauron` CLI: every command with its synopsis,
intent, and key flags, each linking to the feature contract that owns its full
behavior. Every command obeys the same command grammar, shared flags,
exit-status semantics, and output discipline.

## add repository

```
sauron add repository [--kind <kind>] --priority <n> [kind-scoped flags] <name> <location>
```

Register an artifact source of any kind.

- Key flags: `--kind` (default `http`), `--priority` (required), `--timeout`
  (http/git); kind-scoped auth/TLS flags (`--username`/`--password`,
  `--skip-tls-verify`, `--ca-cert`, `--client-cert`/`--client-key`, `--ssh-key`).
- Full contract → [0001-add-repository](../0001-add-repository/contracts/command-line.md).

## list repositories

```
sauron list repositories [--search <term>] [--sort <name|priority|kind>] [--order <asc|desc>]
```

Review configured sources.

- Key flags: `--search`, `--sort`, `--order`.
- Full contract → [0002-list-repositories](../0002-list-repositories/contracts/command-line.md).

## delete repository

```
sauron delete repository <name>
```

Unregister a source, keeping installed artifacts.

- Key flags: none.
- Full contract → [0003-delete-repository](../0003-delete-repository/contracts/command-line.md).

## prune

```
sauron prune [skills|agents] [--dry-run]
```

Remove artifacts orphaned by unregistered repositories.

- Key flags: `--dry-run`; optional `skills`|`agents` positional narrows scope.
- Full contract → [0004-prune](../0004-prune/contracts/command-line.md).

## import persona

```
sauron import persona [--priority <n>] <path>
```

Define a named artifact set for a group.

- Key flags: `--priority`.
- Full contract → [0005-import-persona](../0005-import-persona/contracts/command-line.md).

## delete persona

```
sauron delete persona <name>
```

Remove a persona, keeping installed artifacts.

- Key flags: none.
- Full contract → [0006-delete-persona](../0006-delete-persona/contracts/command-line.md).

## list personas

```
sauron list personas [--search <term>] [--tag <tag>]... [--sort <name|priority>] [--order <asc|desc>]
```

Review defined personas.

- Key flags: `--search`, `--tag` (repeatable), `--sort`, `--order`.
- Full contract → [0007-list-personas](../0007-list-personas/contracts/command-line.md).

## update persona

```
sauron update persona <path>
```

Revise a persona's definition.

- Key flags: none.
- Full contract → [0008-update-persona](../0008-update-persona/contracts/command-line.md).

## sync

```
sauron sync [--persona <name>] [--dry-run]
```

Reconcile the target with repositories and personas.

- Key flags: `--persona`, `--dry-run`.
- Full contract → [0009-sync](../0009-sync/contracts/command-line.md).

## set priority persona

```
sauron set priority persona <name> <value>
```

Reorder persona precedence.

- Key flags: none.
- Full contract → [0010-set-persona-priority](../0010-set-persona-priority/contracts/command-line.md).

## set priority repository

```
sauron set priority repository <name> <value>
```

Reorder repository precedence for conflict resolution.

- Key flags: none.
- Full contract → [0011-set-repository-priority](../0011-set-repository-priority/contracts/command-line.md).

## set target

```
sauron set target <claude|zencoder> [--copy-only]
```

Choose the provider destination.

- Key flags: `--copy-only`.
- Full contract → [0012-set-target](../0012-set-target/contracts/command-line.md).

## clear

```
sauron clear [--persona <name>] [--dry-run]
```

Remove all Sauron-installed artifacts.

- Key flags: `--persona`, `--dry-run`.
- Full contract → [0013-clear](../0013-clear/contracts/command-line.md).

## cron sync

```
sauron cron sync <expression>
sauron cron sync --disable
```

Schedule automatic sync via the OS crontab.

- Key flags: `--disable` (mutually exclusive with `<expression>`).
- Full contract → [0014-cron-sync](../0014-cron-sync/contracts/command-line.md).

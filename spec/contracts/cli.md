# Command Line Interface Contract

Global, normative contract for the `sauron` CLI. This document owns the
conventions every command follows; each command's full behavior (flags,
outputs, exit conditions) is owned by its feature contract, linked from the
[command index](#command-index). Authoring conventions live in
[spec/AUTHORING.md](../AUTHORING.md); project principles in
[CONSTITUTION.md](../../CONSTITUTION.md).

## Command grammar

```
sauron <verb> [<noun> [<noun>]] [flags] <args...>
```

- Verb–noun hierarchy: `add repository`, `list personas`,
  `set priority repository`, `cron sync`.
- Flags are GNU-style long options: `--flag` for booleans, `--flag <value>`
  otherwise. Repeatable flags are marked `...` in synopses.
- Positional arguments follow flags in synopses and are written `<name>`.

## Shared flags

These flags mean the same thing in every command that accepts them. A feature
contract may narrow a shared flag (e.g. restrict `--sort` values) but may not
contradict it.

| Flag | Meaning |
|---|---|
| `--dry-run` | Print the plan without changing the environment or the track file |
| `--priority <n>` | Integer precedence; lower value wins; unique within its namespace |
| `--kind <kind>` | Repository kind: `http` (default), `filesystem`, or `git` |
| `--search <term>` | Case-insensitive substring filter |
| `--sort <field>` | Sort field for list output |
| `--order <asc\|desc>` | Sort direction, default `asc` |
| `--persona <name>` | Scope the operation to one persona's artifacts |
| `--timeout <duration>` | Bound on network operations, default `30s` |

## Exit status (normative)

| Code | Meaning |
|---|---|
| `0` | Success — including idempotent no-ops: deleting an absent resource, an empty list, an already-up-to-date sync, an already-set value, and any `--dry-run` run |
| `2` | Usage error — invalid or missing arguments/flags; nothing was executed |
| `1` | Runtime error — validation failure, unreadable configuration or track file, unreachable external resource, or a failed artifact operation |

Feature contracts may only refine *which conditions* map to each code; they
never redefine these meanings.

## Output discipline

- Results (confirmations, tables, plans) go to stdout.
- A failing command writes exactly one human-readable message to stderr and
  produces no partial output.
- Commands that apply changes in bulk (`sync`, `prune`, `clear`, `set target`)
  print a shared plan/report format: artifacts grouped under `skills:` and
  `agents:` headings, one artifact per line, prefixed `+` for additions/updates
  and `-` for removals, followed by a summary count line when changes are
  applied. Per-artifact failures are reported without stopping the run.

## Command index

| Command | Synopsis | Intent | Contract |
|---|---|---|---|
| add repository | `sauron add repository [--kind <kind>] --priority <n> [kind-scoped flags] <name> <location>` | Register an artifact source | [contract](../0001-add-repository/contracts/command-line.md) |
| list repositories | `sauron list repositories [--search <term>] [--sort <name\|priority\|kind>] [--order <asc\|desc>]` | Review configured sources | [contract](../0002-list-repositories/contracts/command-line.md) |
| delete repository | `sauron delete repository <name>` | Unregister a source, keeping installed artifacts | [contract](../0003-delete-repository/contracts/command-line.md) |
| prune | `sauron prune [skills\|agents] [--dry-run]` | Remove artifacts orphaned by unregistered repositories | [contract](../0004-prune/contracts/command-line.md) |
| import persona | `sauron import persona [--priority <n>] <path>` | Define a named artifact set for a group | [contract](../0005-import-persona/contracts/command-line.md) |
| delete persona | `sauron delete persona <name>` | Remove a persona, keeping installed artifacts | [contract](../0006-delete-persona/contracts/command-line.md) |
| list personas | `sauron list personas [--search <term>] [--tag <tag>]... [--sort <name\|priority>] [--order <asc\|desc>]` | Review defined personas | [contract](../0007-list-personas/contracts/command-line.md) |
| update persona | `sauron update persona <path>` | Revise a persona's definition | [contract](../0008-update-persona/contracts/command-line.md) |
| sync | `sauron sync [--persona <name>] [--dry-run]` | Reconcile the target with repositories and personas | [contract](../0009-sync/contracts/command-line.md) |
| set priority persona | `sauron set priority persona <name> <value>` | Reorder persona precedence | [contract](../0010-set-persona-priority/contracts/command-line.md) |
| set priority repository | `sauron set priority repository <name> <value>` | Reorder repository precedence for conflict resolution | [contract](../0011-set-repository-priority/contracts/command-line.md) |
| set target | `sauron set target <claude\|zencoder> [--copy-only]` | Choose the provider destination | [contract](../0012-set-target/contracts/command-line.md) |
| clear | `sauron clear [--persona <name>] [--dry-run]` | Remove all Sauron-installed artifacts | [contract](../0013-clear/contracts/command-line.md) |
| cron sync | `sauron cron sync <expression>` / `sauron cron sync --disable` | Schedule automatic sync via OS crontab | [contract](../0014-cron-sync/contracts/command-line.md) |

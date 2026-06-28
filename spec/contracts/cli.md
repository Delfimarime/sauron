# CLI Contract

The normative conventions every `sauron` command obeys: command grammar, shared
flags, exit-status semantics, and output discipline. This is a binding contract,
alongside the [terminal UI contract](terminal-ui.md) (the interactive surface),
the [architecture contract](architecture.md), and the
[state data contract](state.md). The CLI is sauron's **headless surface** — for
scripting, CI, and OS-crontab scheduling; the TUI is the interactive surface.
Both render the same presentation-agnostic use-case results and share one
field/flag vocabulary, so a name means the same thing on either surface. Each
feature owns one or more commands; every command's `contracts/<verb>-<noun>.md`
conforms to the rules here. A feature contract may refine *which conditions* map
to which exit code; it may not redefine the meanings.

Authoring mechanics (spec structure, numbering, EARS, the canonical
requirement boilerplate) live in [AUTHORING.md](../AUTHORING.md), which points
here for the command conventions.

## Command grammar

```
sauron <verb> [<noun> [<noun>]] [flags] <args...>
```

- Verb–noun hierarchy: `set registry`, `describe registry`, `unset registry`,
  `list catalogue skill`, `install skill`, `uninstall agent`, `sync`, `upgrade`,
  `set provider`, `describe skill`. `unset` is the inverse of `set` for the
  registry; `uninstall` is the inverse of `install`; `describe` shows a single
  resource's detail.
- A feature may own a **family** of sibling commands that differ by their noun
  (e.g. `list skills` / `list agents`); each such command is a distinct command
  with its own contract file.
- Flags are GNU-style long options: `--flag` for booleans, `--flag <value>`
  otherwise. Repeatable arguments are marked `...` in synopses.
- Positional arguments follow flags in synopses and are written `<name>`.

## Shared flags

These flags mean the same thing in every command that accepts them. A feature
contract may narrow a shared flag (e.g. restrict `--sort` values) but may not
contradict it.

| Flag | Meaning |
|---|---|
| `--dry-run` | Print the plan without changing the environment or the track file |
| `--transport <kind>` | Registry transport: `git` or `http` (persisted as `spec.transport`) |
| `--search <term>` | Case-insensitive substring filter |
| `--sort <field>` | Sort field for list output |
| `--order <asc\|desc>` | Sort direction, default `asc` |
| `--page <n>` | Page number, 1-based (catalogue paging), default `1`; the client computes the backend offset as `(page−1)·limit` |
| `--limit <n>` | Page size (catalogue paging), default `20` |
| `--fields <list>` | Comma-separated columns to display, in order, for `list` and `describe`; the identity field is always present and first. Each contract defines its valid field set |
| `--timeout <duration>` | Bound on network operations, default `30s` |
| `--theme <name>` | Color theme: `sauron` (dark, default) or `light`; honored by both surfaces. The `NO_COLOR` environment variable disables color entirely |

## Exit status

| Code | Meaning |
|---|---|
| `0` | Success — including idempotent no-ops: uninstalling an absent artifact, an empty list, an already-current sync/upgrade, an already-set value, and any `--dry-run` run |
| `2` | Usage error — invalid or missing arguments/flags; nothing was executed |
| `1` | Runtime error — validation failure, unreadable state or track file, unreachable registry, or a failed artifact operation |

Feature contracts may only refine *which conditions* map to each code; they never
redefine these meanings.

## Output discipline

- Results (confirmations, tables, plans) go to stdout.
- A failing command writes exactly one human-readable message to stderr and
  produces no partial output.
- Commands that apply changes in bulk (`install`, `uninstall`, `sync`, `upgrade`,
  `set provider`) print a shared plan/report format: artifacts grouped under
  `skills:` and `agents:` headings, one artifact per line, prefixed `+` (added),
  `~` (updated in place), or `-` (removed), followed by a summary count line when
  changes are applied. Per-artifact failures are reported without stopping the run.

### Canonical rendering

Every command contract presents an `## Example` instantiating one of these
formats; the formats themselves are fixed here.

**Tables** (`list`) — aligned columns, uppercase headers, `—` for an absent
optional value:

```
NAME        VERSION  UPDATED
go-style    v1.4.0   2026-06-15
sql-review  —        2026-06-12
```

**Catalogue** (`list catalogue`) — a table followed by one paging line:

```
NAME         KIND
go-style     skill
code-helper  skill
showing 1–20 (page 1, limit 20)
```

**Detail** (`describe`) — left-aligned `key:` values:

```
name:     go-style
version:  9f4d2a1
path:     sauron-go-style
updated:  2026-06-15
```

**Plan / report** (bulk operations) — kind headings, two-space indent, `+`/`~`/`-`
and the target name, then a summary line:

```
skills:
  + sauron-go-style
  ~ sauron-sql-review
agents:
  - sauron-old-reviewer
1 added, 1 updated, 1 removed
```

Under `--dry-run` the same plan is printed beneath a `(dry run)` header and nothing
is applied.

**Confirmation** (`set registry`, `unset registry`) — a single line:

```
registry set to git@github.com:acme/artifacts.git (git)
```

**Failure** — exactly one `error:`-prefixed line to stderr, no partial stdout:

```
error: registry source unreachable
```

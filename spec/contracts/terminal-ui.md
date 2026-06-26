# Terminal UI Contract

**Status:** Deferred — this contract is binding for the eventual TUI, but no TUI
is implemented or supported until the headless CLI is complete (see
[ADR-0005](../architecture/ADR-0005-terminal-ui-deferred.md)).

The normative contract for sauron's interactive **terminal UI (TUI)** — the
full-screen application launched when `sauron` is run interactively. It is a
binding contract alongside the [CLI contract](cli.md) (the headless command
surface), the [state data contract](state.md), and the
[architecture contract](architecture.md).

The TUI and the headless CLI are **two presentations of the same use cases**: a
use case holds all behavior and returns a presentation-agnostic result; the CLI
renders it through its `view_<name>.go`, the TUI renders it on screen. Neither
surface introduces behavior the other lacks, and both draw from **one shared
vocabulary** of field, flag, and value names — defined here and in the
[CLI contract](cli.md), never diverging. [terminal-ui.html](terminal-ui.html) is
the non-normative visual companion that renders the palette and layouts.

## Surface selection

- With a terminal attached and no command, `sauron` launches the TUI.
- With a verb-noun command (e.g. `sauron sync`), or with no TTY (a pipe, cron,
  CI), sauron runs headless per the [CLI contract](cli.md).
- `--theme <name>` and the `NO_COLOR` environment variable apply to both surfaces.

## Shared vocabulary

Every value the TUI displays uses these canonical names, identical to the
[CLI contract](cli.md). The identity field is always present and shown first.

| Field | Meaning |
|---|---|
| `name` | artifact / provider identity |
| `source` | registry location (the registry's identity) |
| `transport` | registry transport: `git`, `http`, or `filesystem` |
| `revision` | git branch, tag, or commit (git transport only) |
| `credentials` | `username` / `password`, environment references only (`${env:VAR}`) |
| `tls` | TLS settings |
| `sshKey` | SSH private-key reference |
| `timeout` | bound on network operations |
| `version` | artifact version |
| `digest` | content identity |
| `path` | installed location (`sauron-<name>`) |
| `kind` | artifact kind: `skill` or `agent` |
| `created` | document creation time |
| `updated` | document last-change time |
| `installed` | artifact install time |

No view ever prints the raw on-disk audit-timestamp keys; the
[state data contract](state.md) owns them and views render the display names
above (`created`, `updated`, `installed`).

## Interaction model

- The TUI manages the single global registry and single global provider (see the
  [state data contract](state.md)).
- It opens on the **resource browser** when a registry is configured, or the
  **empty state** otherwise.
- Navigation is modal: a list view, a detail view, a diff/plan view, and modal
  forms. `esc` always returns to the previous view; `q` from the root exits.
- Every interactive affordance maps to the same concept as a headless flag, so
  the two surfaces stay aligned:

| Concept | Headless flag | TUI affordance |
|---|---|---|
| Filter substring | `--search <term>` | `/` search |
| Choose columns | `--fields <list>` | `f` fields picker |
| Sort | `--sort <field>` `--order` | column sort |
| Paging | `--page <n>` `--limit <n>` | catalogue paging |
| Preview only | `--dry-run` | the diff view shown before `↵ apply` |
| Theme | `--theme <name>` | `m` cycle |

The `tab` kind toggle and the installed/catalogue distinction correspond to the
headless commands' noun selection (`list skills` / `list agents` /
`list catalogue`), not to flags.

## Global chrome

- **Status bar** — the configured `source` (registry) and the active `theme`.
- **Key-binding overlay** (`?`) — lists every binding; `esc`/`q` closes.
- **Empty state** — when no registry is configured, the browser is replaced by a
  prompt to set one (`E set registry`); nothing can be browsed, installed,
  synced, or upgraded until then.
- **Loading** — a `⟳` indicator while a network operation runs.
- **Error line** — a failure surfaces as a single in-app `error:`-prefixed line
  (the same message the headless surface writes to stderr); exit-status semantics
  are a headless concern owned by the [CLI contract](cli.md).

## Key bindings

```
j/k move    tab kind     ↵ describe    / search    f fields
i installed s sync       u upgrade     r registry  p provider
m theme     ? help                                  esc · q close
```

Bindings are view-scoped: the same key may act differently per view (e.g. `↵`
describes in the resource browser and installs in the catalogue).

## Views

Each view's columns or `key:` fields are exactly the shared-vocabulary names.
`◀` marks the cursor.

### Resource browser

The home view: a table of installed skills and agents.

```
┌─ sauron ──────────────────────────────────────────── theme: sauron ─┐
│ ┤ RESOURCE ├                                        registry: acme   │
│                                                                      │
│ NAME            VERSION   UPDATED      KIND                          │
│ go-style        v1.4.0    2026-06-15   skill   ◀                     │
│ sql-review      —         2026-06-12   skill                         │
│ code-reviewer   3af1c2e   2026-06-14   agent                         │
│                                                                      │
│ 3 installed                                                          │
├──────────────────────────────────────────────────────────────────  ┤
│ j/k move · tab kind · ↵ describe · / search · f fields · i installed │
│ s sync · u upgrade · r registry · p provider · m theme · ? keys      │
└──────────────────────────────────────────────────────────────────  ┘
```

### Catalogue browser

What the registry offers, with the install entry point. Follows the catalogue
paging format of the [CLI contract](cli.md).

```
┌─ sauron · catalogue ──────────────────────────────── theme: sauron ─┐
│ ┤ CATALOGUE ├                                       registry: acme   │
│                                                                      │
│ NAME           KIND                                                  │
│ go-style       skill   ◀  ✓ installed                               │
│ code-helper    skill                                                 │
│ code-reviewer  agent                                                 │
│                                                                      │
│ showing 1–3 (page 1, limit 20)                                       │
├──────────────────────────────────────────────────────────────────  ┤
│ space select · ↵ install · / search · n/p page · esc back            │
└──────────────────────────────────────────────────────────────────  ┘
```

### Describe registry

```
┌─ ┤ DESCRIBE · REGISTRY ├ ───────────────────────────────────────────┐
│                                                                      │
│ source:       git@github.com:acme/artifacts.git                     │
│ transport:    git                                                    │
│ revision:     main            (branch, tag, or commit)              │
│ credentials:                                                         │
│   username:   ${env:ACME_USER}                                       │
│   password:   ${env:ACME_TOKEN}                                      │
│ timeout:      30s                                                    │
│ created:      2026-06-21 07:30                                       │
│ updated:      2026-06-21 07:30                                       │
│                                                                      │
│ E edit registry   D unset      edit re-resolves HEAD · unset detaches│
└──────────────────────────────────────────────────────────────────  ┘
```

### Describe provider

```
┌─ ┤ DESCRIBE · PROVIDER ├ ───────────────────────────────────────────┐
│                                                                      │
│ name:    claude            ● active                                  │
│                                                                      │
│ The single global provider — where sauron installs artifacts.       │
│ Switching migrates tracked artifacts to the new provider's dir.     │
│                                                                      │
│ SWITCH TO:   claude (current) · zencoder                            │
│                                                                      │
│ ⟳ sync   ⬆ upgrade                                       esc to close │
└──────────────────────────────────────────────────────────────────  ┘
```

### Describe artifact (skill / agent)

```
┌─ ┤ DESCRIBE · SKILL ├ ──────────────────────────────────────────────┐
│                                                                      │
│ name:       go-style                                                 │
│ version:    v1.4.0                                                   │
│ digest:     sha256:1a2b…                                             │
│ path:       skills/sauron-acme-go-style                             │
│ installed:  2026-06-10 09:00                                         │
│ updated:    2026-06-15 10:00                                         │
│                                                                      │
│                                                  esc · back to list  │
└──────────────────────────────────────────────────────────────────  ┘
```

### Diff / plan

Shown for `install`, `uninstall`, `sync`, `upgrade`, and provider switch. It is a
preview (a `--dry-run` in headless terms): nothing is written until `↵ apply`. The
body uses the plan/report format of the [CLI contract](cli.md).

```
┌─ sauron · sync (dry run) ───────────────────────────────────────────┐
│ computing the diff — nothing is written during a dry-run            │
│                                                                      │
│ skills:                                                              │
│   ~ sauron-acme-go-style                                             │
│   - sauron-acme-old-skill                                            │
│ agents:                                                              │
│   + sauron-acme-new-reviewer                                         │
│                                                                      │
│ 1 added, 1 updated, 1 removed                                        │
├──────────────────────────────────────────────────────────────────  ┤
│ ↵ apply (run) · esc cancel                                           │
└──────────────────────────────────────────────────────────────────  ┘
```

When the diff is empty, the body shows `✓ already current — nothing to change.`
and there is nothing to apply.

## Forms

Form field labels are the canonical field names.

### Set registry

```
┌─ SET REGISTRY ──────────────────────────────────────────────────────┐
│                                                                      │
│ transport:   ( git )  http   filesystem          (--transport)      │
│ source:      [ git@github.com:acme/artifacts.git            ]       │
│ revision:    [ main                          ]  optional            │
│ credentials: [ ${env:ACME_USER} ] / [ ${env:ACME_TOKEN} ]          │
│                                                                      │
│                                   set registry        cancel  ◀     │
└──────────────────────────────────────────────────────────────────  ┘
```

A literal secret in `credentials` is rejected, matching the headless rule
(environment references only).

### Switch provider

```
┌─ SWITCH PROVIDER? ──────────────────────────────────────────────────┐
│                                                                      │
│   claude  →  zencoder                                                │
│                                                                      │
│ ⚠ A sync is required afterwards to reconcile the new provider's     │
│   directory with the registry.                                       │
│                                                                      │
│ switch & migrate        esc cancel                                   │
└──────────────────────────────────────────────────────────────────  ┘
```

### Unset registry

```
┌─ UNSET REGISTRY? ───────────────────────────────────────────────────┐
│ Detaches the source; installed artifacts are preserved.            │
│                                          ↵ confirm · esc cancel      │
└──────────────────────────────────────────────────────────────────  ┘
```

### Fields picker

Selects which columns a list view shows (`--fields`); the identity field is
locked first.

```
┌─ FIELDS · columns to display (--fields) ────────────────────────────┐
│ [x] name      (identity — always first)                            │
│ [x] version                                                         │
│ [x] updated                                                         │
│ [ ] digest                                                          │
│ [ ] path                                                            │
│                                                        esc to close  │
└──────────────────────────────────────────────────────────────────  ┘
```

## Theme

- Two themes: `sauron` (dark, the default) and `light`.
- `m` cycles the active theme in the TUI and persists the choice; `--theme <name>`
  overrides it for a single invocation on either surface; the `NO_COLOR`
  environment variable disables color entirely (structure is still drawn in the
  terminal's default foreground).
- The persisted preference is owned by the [state data contract](state.md).
- The palette is defined as semantic roles, each mapped to the two palettes in
  [terminal-ui.html](terminal-ui.html):

| Role | `sauron` (dark) | `light` |
|---|---|---|
| background | `#0d0907` | `#fbf7f0` |
| surface | `#150b06` | `#f1eadd` |
| border | `#3a2a18` | `#ddd0bb` |
| text | `#cdbba6` | `#2c2316` |
| text-dim | `#6b5949` | `#9a8b76` |
| accent | `#f3b145` | `#b5650f` |
| added `+` | `#5fcf6b` | `#1f7a33` |
| updated `~` | `#ffce6a` | `#d97712` |
| removed `-` | `#e0604f` | `#a82e22` |

## Verification

The TUI is verified at three layers: behavior by the shared use-case unit tests;
its Model→Update→View state machine by unit tests (including a golden assertion
that TUI and CLI render the same canonical labels); and end-to-end by
pseudo-terminal Gherkin scenarios that mirror each interactive feature's headless
scenario. The integration-test conventions are owned by the
[architecture contract](architecture.md) and the `test/e2e` harness; the
pseudo-terminal scenarios are named `terminal_ui_<command>.feature`.

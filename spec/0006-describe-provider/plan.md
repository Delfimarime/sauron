# Describe Provider — implementation plan

How [`describe provider`](spec.md) is built. The executable task breakdown lives
in [TASKS.md](TASKS.md). Code follows the
[architecture contract](../contracts/architecture.md) and the
`sauron-implementing-architecture` conventions; the end-to-end suite follows
`sauron-implementing-integration-tests`.

## TDD discipline (mandatory order)

The implementation is **test-driven, end-to-end first**:

1. **Start with `test/e2e`.** Author the Gherkin feature and its controller
   *before any production code*. It must run **red** — `describe provider` does
   not exist yet, so `task gate-integration` fails on the missing command. This
   red run is the proof the suite drives real behavior.
2. **Then implement the code** — data model → schema → state contract → use case
   → view → command — each unit written **test-first** (its unit test before its
   production code), turning checkpoints green in order.
3. **Conclude by running `task all`.** The whole contract gate — test, lint,
   build, coverage (≥80%), security, and integration — must end green. The e2e
   feature from step 1 now passes.

## Goal & scope

Add the `describe provider` command: render the active provider's detail through
the shared descriptor view (FR-001) — `name`, the derived `directory`
(`claude` → `~/.claude`, `zencoder` → `~/.zencoder`), `labels`, the audit
`createdAt`/`lastUpdatedAt` timestamps, and the sync `lastSyncedAt`/`lastSyncAttemptAt`
timestamps. `--fields <list>` selects and orders the displayed fields, `name`
always present and first (FR-002). When no provider is set, print
`no provider is set` and exit `0` (FR-003). An unreadable `settings.yaml` is a
runtime error, exit `1` (FR-004). Invalid flags exit `2` before the command runs
(FR-005). The slice mirrors the existing `describe registry`
(use case → `runUseCase` handler → `view_*` renderer → typed store).

**In scope:** giving `Provider` a `spec{lastSyncedAt, lastSyncAttemptAt}` in the
Go type, the **normative** [`Provider.schema.json`](../contracts/schemas/Provider.schema.json),
and the **shared** [state contract](../contracts/state.md); the describe slice;
deriving `directory` for display.

**Out of scope (YAGNI):** *writing* the sync timestamps — that is
[sync](../0011-sync/spec.md), not built yet. Here `lastSyncedAt`/`lastSyncAttemptAt`
are **read-only, render-when-present**: omitted while absent, exactly as
`createdAt`/`lastUpdatedAt` already tolerate absence. No new store —
`ProvidersStore.Get` already exists (from 0005).

## Pre-requirements

- The `Provider` type, its kind constant (`types.KindProvider`), and the singleton
  `ProvidersStore` (`Get`/`Set`) already exist (from
  [set provider](../0005-set-provider/spec.md)).
- The shared descriptor view (`descriptor`/`descriptorField`,
  [`view_descriptor.go`](../../internal/cmd/view_descriptor.go)) already renders
  the aligned `label: value` column with indented nested sections — reused
  untouched.

## Design

Slice parallels `describe registry`, plus a small data-model addition. New and
touched files:

| File | Change |
|---|---|
| `pkg/sauron/types/provider.go` | add `ProviderSpec{LastSyncedAt, LastSyncAttemptAt string}` and a `Spec ProviderSpec` field on `Provider` |
| `spec/contracts/schemas/Provider.schema.json` | add an optional `spec` object — two `date-time` properties, `additionalProperties: false` |
| `spec/contracts/state.md` | Provider per-kind: replace "Provider has none" — note the `spec` carries sync timestamps **written by `sync` (0011)**, tolerated absent on read |
| `spec/0006-describe-provider/spec.md` | FR-002 valid fields → `name, directory, labels, createdAt, lastUpdatedAt, lastSyncedAt, lastSyncAttemptAt` |
| `spec/0006-describe-provider/contracts/describe-provider.md` | `--fields` table + example updated to the full field set |
| `spec/0006-describe-provider/data/state.md` | field realization rows: derived `directory`, `spec.lastSyncedAt`, `spec.lastSyncAttemptAt` |
| `internal/usecase/usecase_describe_provider.go` | **new** — `DescribeProviderUseCase.Execute` → `ProvidersStore.Get`; returns `(nil, nil)` when none set, `NewIOError` on read failure |
| `internal/usecase/fx.go` | provide `NewDescribeProviderUseCase` |
| `internal/cmd/view_describe_provider.go` | **new** — provider field set, selector (`name` forced first, dedup), `directory` derivation, `labels` → sorted section, the `no provider is set` line |
| `internal/cmd/cmd_describe_provider.go` | **new** — `DescribeProvider()` builder + cobra-free `describeProvider()` handler; a nil result renders the none-set line and exits `0` |
| `internal/cmd/cmd_describe.go` | `cmd.AddCommand(DescribeProvider())` |
| `test/e2e/testdata/describe_provider.feature` + `.../describe_provider_controller.go` | **new** — first describe, `--fields`, none-set (exit 0), invalid field (exit 2), and a synced-provider doc-string asserting the full descriptor `reads:` block |

### Field set and provider directory

The valid `--fields` set, in default order, is `name`, `directory`, `labels`,
`createdAt`, `lastUpdatedAt`, `lastSyncedAt`, `lastSyncAttemptAt`. `name` is identity — always
present and first. `directory` is **derived** from the provider name in the view
(`claude` → `~/.claude`, `zencoder` → `~/.zencoder`), reusing the
provider→home mapping from set-provider/migrate if a constant exists, else a
two-entry switch — never stored. `labels` renders as an indented section with its
keys sorted for deterministic output. Leaf timestamp fields render only when
present, so the common pre-sync provider shows `name`, `directory`, `createdAt`,
`lastUpdatedAt` and nothing else.

### Success display

The active provider on stdout, via the shared descriptor renderer (value column =
widest leaf label + colon + a two-space gap).

Common case today (set, never synced — sync fields absent, so omitted):

```
$ sauron describe provider
name:           claude
directory:      ~/.claude
createdAt:      2026-06-21T07:30:00Z
lastUpdatedAt:  2026-06-21T07:30:00Z
```

Full detail once a sync has run (column auto-aligns to `lastSyncAttemptAt`):

```
$ sauron describe provider
name:               claude
directory:          ~/.claude
labels:
  team:             backend
createdAt:          2026-06-21T07:30:00Z
lastUpdatedAt:      2026-06-22T08:00:00Z
lastSyncedAt:       2026-06-25T09:15:00Z
lastSyncAttemptAt:  2026-06-26T06:00:00Z
```

`--fields` selects and orders (`name` always first):

```
$ sauron describe provider --fields directory,lastSyncedAt
name:          claude
directory:     ~/.claude
lastSyncedAt:  2026-06-25T09:15:00Z
```

No provider set (FR-003 — plain line, exit `0`, not an error):

```
$ sauron describe provider
no provider is set
```

### `DescribeProviderUseCase.Execute` flow

1. `ProvidersStore.Get` the configured provider.
2. On read failure → `NewIOError` (FR-004, exit 1).
3. When none is set → return `(nil, nil)`; the command renders
   `no provider is set` and exits `0` (FR-003).
4. Otherwise return the `*types.Provider`; field selection and directory
   derivation are presentation concerns resolved by the view.

## Key decisions

- **`Provider` gains a `spec` for sync state.** `lastSyncedAt`/`lastSyncAttemptAt`
  are provider-/sync-specific, not generic audit metadata, so they live in
  `Provider.spec` (mirroring every other kind), not under the shared `metadata`
  envelope. They are written by `sync` (0011) and only read here.
- **`directory` is derived, never stored.** It is a pure function of the provider
  name; storing it would duplicate the name and invite drift.
- **None-set is not an error.** Unlike `describe registry` (which returns
  `NotFoundError`), the use case returns `(nil, nil)`; the command prints the
  none-set line and exits `0` (FR-003).
- **Reuse the descriptor view** (`descriptor`/`descriptorField`/`view_descriptor.go`)
  untouched. Only a provider-specific field selector/projection is new.
- **Mirror `describe registry`** for the command/use-case/view shape.

## Checkpoints

| # | Milestone | Verify |
|---|---|---|
| C1 | e2e feature authored, fails **red** (command missing) | `task gate-integration` fails only on the missing `describe provider` |
| C2 | `Provider.spec` type + schema + state-contract note compile; schema validates a synced doc | `go test ./pkg/...` |
| C3 | use case unit-tested (set / none → nil,nil / read-error) | `go test ./internal/usecase/...` |
| C4 | view + command unit-tested; whole tree builds | `go build ./... && go test ./internal/cmd/...` |
| C5 | full contract gate green | `task all` |

## Execution flow

Sequential chain: **T1 → T2 → T3 → T4 → T5 → T6** (see [TASKS.md](TASKS.md)).
T1 (the `.feature` + controller) touches only `test/e2e/**` and may be **drafted
in a parallel git worktree** (`branch feat/describe-provider-e2e`, merged back
into the working tree uncommitted), but it only *passes* once T5 lands. T6 runs
`task all` as the closing gate.

# List Artifacts — implementation plan

How `list skills`/`list agents` ([spec.md](spec.md)) are built. The executable
task breakdown lives in [TASKS.md](TASKS.md). Code follows the
[architecture contract](../contracts/architecture.md) and the
`sauron-implementing-architecture` conventions; the end-to-end suite follows
`sauron-implementing-integration-tests`.

## Goal & scope

Add `list skills`/`list agents`: read `track.yaml` via the existing
`storage.TrackStore`, filter to the invoked kind, apply `--search` (FR-004) and
`--sort`/`--order` (FR-005), page the result with `--page`/`--limit` (FR-003, no
total count, mirroring `list catalogue`'s paging line), and print a
`name`/`version`/`lastUpdatedAt` table selectable via `--fields`, `name` always
first (FR-001/FR-002). Nothing installed, or a page past the end, prints the empty
paging line and exits 0 (FR-006). An unreadable `track.yaml` is a runtime error
(FR-007); missing/invalid flags exit 2 (FR-008).

**This is the second use of the shared listing seam, not a one-off.** The
[use case contract's "Listing use cases"](../contracts/use-case.md#listing-use-cases)
section — landed ahead of this feature, together with `list catalogue`'s
refactor onto it — fixes the generic `ListUseCase[I, T]`
(`internal/usecase/list.go`, composing `Lister[T]`/`ListWindow`/`ListResult[T]`/
`listWith[T]`) as the standing shape for any use case that lists a collection.
There is no analogous list-specific scaffold at the command layer — every
command, listing or not, builds on the same `newCommand` (`internal/cmd/helper.go`);
a listing leaf command builder (e.g. `newCatalogueCommand(kind
usecase.CatalogueKind, use, short, long string)`) simply closes over its own
kind and a `listFlags`/`fieldsFlags`-based flag struct in the `withFlags`/
`withRunE` options it passes to `newCommand`. This feature supplies the second
adapter at each layer: a `trackLister` composing `storage.TrackStore` instead
of a live registry, and `ListSkills()`/`ListAgents()` built via a
`newListArtifactsCommand(kind, use, short, long)` that closes over `kind`
exactly as `newCatalogueCommand` does, itself calling `newCommand` directly.
No new store method is needed.

**Out of scope (YAGNI):** a total count in the paging line (matches `list
catalogue`'s own convention, by choice rather than constraint — see D4); a `kind`
column (each command is already kind-scoped, per the feature's Notes); any field
beyond `name`/`version`/`lastUpdatedAt`.

## Test-driven approach (Constitution Ch. III, Art. 1)

- **Outer loop — acceptance first.** The `test/e2e` Gherkin scenarios are authored
  first (T1) and fail (the commands don't exist); they turn green at T3.
- **Inner loop — unit first.** `NewListArtifactsUseCase` (T2) and the command
  (T3) each start from a failing test.
- **Mocks at the seams.** The use-case test mocks `storage.TrackStore` (the
  existing `MockBasedTrackStore`); no adapter or network fixture is needed. The
  shared `Lister[T]`/`listWith[T]` seam itself already carries its own unit tests
  (`internal/usecase/list_test.go`) and is not re-verified here — only the
  feature-specific `trackLister` adapter is.

## Pre-requirements (already in place)

- **The shared listing seam** — `internal/usecase/list.go`: `Lister[T]`
  (`List(ctx, opts ...source.Option) ([]T, error)`), `ListWindow`
  (`Search`/`Sort`/`Order`/`Page`/`Limit`, embedded by a listing request),
  `ListResult[T]` (`Items`/`Page`/`Limit`/`Offset`, embedded by a listing
  response), `listWith[T]` (validates the window, calls the `Lister`, wraps the
  page), and the generic `ListUseCase[I, T]` (`resolve func(ctx, in I) (Lister[T],
  ListWindow, error)`, `Execute` calling `listWith`). `NewListCatalogueUseCase`
  returns a `*ListUseCase[ListCatalogueRequest, string]` **directly** — no
  wrapping type, per [no pointless wrappers](../contracts/architecture.md#coding-standards)
  — with `resolve` built as the validate→get→open pipeline ending in a
  `catalogueLister`; `ListCatalogueResponse` is correspondingly a **type
  alias** (`= ListResult[string]`), not a struct, since the caller already
  knows `Kind` before `Execute` runs. This feature's `NewListArtifactsUseCase`
  mirrors that exactly: `resolve` ends in a `trackLister` in place of a
  `catalogueLister`, and it likewise returns `*ListUseCase[ListArtifactsRequest,
  types.Artifact]` directly, with `ListArtifactsResponse` as the equivalent
  type alias.
- **The shared command scaffold** — `internal/cmd/helper.go`'s
  `newCommand(use, short string, opts ...commandOption) *cobra.Command`, with
  its `withLong`/`withArgs`/`withFlags`/`withRunE`/`withSubcommands` options.
  It carries no "kind" or listing-specific notion of its own —
  `newCatalogueCommand(kind usecase.CatalogueKind, use, short, long string)`
  closes over `kind` and a `listFlags` value in the `withFlags`/`withRunE`
  options it passes to `newCommand`, and already builds
  `ListCatalogueSkill()`/`ListCatalogueAgent()` this way; this feature's
  `newListArtifactsCommand` mirrors that call exactly, additionally closing
  over a `fieldsFlags` value for `--fields`.
- **The shared paging renderer** — `internal/cmd/helper_view.go`'s
  `pagingLine(page, limit, offset int64, count int) string`, already used
  inline by `newCatalogueCommand`'s `withRunE` closure in `cmd_catalogue.go`
  (there is no separate per-command view file); this feature's handler reuses
  it verbatim.
- `storage.TrackStore.List(ctx) ([]types.Artifact, error)` — reads Skill+Agent
  documents, each discriminated by `TypeMeta.Kind` (`types.KindSkill`/`KindAgent`).
- `types.Artifact` / `ArtifactSpec` (`Metadata.Name`, `Spec.Version`, `Spec.UpdatedAt`).
- Shared command helpers: `selectFields`, `table`/`buildTable`, `defaultOrder`/
  `validateOrder`, `pagingFlags`/`bindPagingFlags`, `listFlags`/`bindListFlags`,
  `fieldsFlags`/`bindFieldsFlags`, `newCommand`/`usageArgs`/`silenceFlagErrors`/
  `runUseCase`.

## Design (implementation map)

| File | Change |
|---|---|
| `internal/usecase/types.go` | add `ListArtifactsRequest{Kind string; ListWindow}` and `ListArtifactsResponse = ListResult[types.Artifact]` — a **type alias**, not a struct, mirroring `ListCatalogueRequest`/`ListCatalogueResponse` exactly |
| `internal/usecase/usecase_list_artifacts.go` | **new** — a `trackLister{track storage.TrackStore, kind string}` implementing `Lister[types.Artifact]` (`List`: `TrackStore.List` → filter to `kind` → resolve the received `source.Option`s into a local `source.Options` value → `--search` substring on `Metadata.Name` → sort by `name`/`lastUpdatedAt` + order → window at offset/limit); `NewListArtifactsUseCase`'s `resolve` closure validates the kind and window then builds a `trackLister`, and the constructor returns `NewListUseCase(resolve)` **directly** — `*ListUseCase[ListArtifactsRequest, types.Artifact]`, no wrapping type — exactly as `NewListCatalogueUseCase` does |
| `internal/usecase/fx.go` | `fx.Annotate(NewListArtifactsUseCase, fx.As(new(UseCase[ListArtifactsRequest, ListArtifactsResponse])))`, alongside the existing entries |
| `internal/cmd/cmd_list_artifacts.go` | **new** — `ListSkills()`/`ListAgents()` → shared `newListArtifactsCommand(kind, use, short, long)`, itself built directly on `newCommand` (closing over `kind` and a `listFlags`+`fieldsFlags` struct in `withFlags`/`withRunE`, exactly as `newCatalogueCommand` does); the `withRunE` closure calls `listArtifacts` via `fx.Populate`/`runUseCase`, then renders the table (`buildTable`/`selectFields`) and the shared `pagingLine` inline — no separate render function or view file, per [Command structure](../contracts/architecture.md#command-structure) |
| `internal/cmd/cmd_list.go` | register `ListSkills()`, `ListAgents()` alongside `Catalogue()` |
| `test/e2e/internal/gherkin/seed_controller.go` | add a `the track file contains:` doc-string step (mirrors the existing `the settings file contains:`) |
| `test/e2e/internal/gherkin/list_artifacts_controller.go` | **new** — row-match step; reuses the existing generic `the paging line reads (.+)$` step verbatim |
| `test/e2e/internal/gherkin/table_assert.go` | **extract** the row/line matching helpers (`catalogueHasRow`→generalized, `hasLine`, `equalFields`) out of `catalogue_controller.go` into a file shared by both controllers |
| `test/e2e/internal/gherkin/init.go` | register the new controller |
| `test/e2e/testdata/list_skills.feature`, `list_agents.feature` | **new (authored first, T1)** |

### D1 — Kind identity

Reuse `types.KindSkill`/`types.KindAgent` (the document `Kind` already stamped on
every tracked `Artifact`) as the request's `Kind`, rather than a parallel enum.
Contrast `CatalogueKind`: that type exists because the catalogue table prints a
lowercase `kind` *column*. This feature has no such column — each command is
already kind-scoped — so there's nothing for a parallel type to serve.

### D2 — Search & sort run in memory, behind `Lister[types.Artifact]`

`track.yaml` has no remote pushdown (unlike the registry), so `trackLister.List`
resolves the `source.Option`s it receives into a plain `source.Options` value
(applying each option to a local zero value — the same technique
`internal/usecase/list_test.go`'s `fakeLister` uses to assert on them) and then
applies `Search`/`Sort`/`Order`/`Offset`/`Limit` itself, entirely in memory, over
the slice `TrackStore.List` returns: `--search` is a case-insensitive
`strings.Contains` on `Metadata.Name`; `--sort lastUpdatedAt` compares
`Spec.UpdatedAt` as a plain string — both are RFC3339 UTC, which sorts
lexicographically correct with no parsing needed. The *mechanism* (local vs.
remote) stays inside the adapter: `NewListArtifactsUseCase`'s `resolve` closure
and `NewListCatalogueUseCase`'s are structurally identical (validate → build the
feature-specific `Lister[T]` → return it with the window), and both return the
same generic `*ListUseCase[I, T]` directly rather than a wrapping type.

### D3 — Table header casing

List tables render lowercase headers — the field name verbatim (`name`,
`version`, `lastUpdatedAt`) — matching the shipped `list catalogue` view
(`tableHeaderName = "name"`, `tableHeaderKind = "kind"` in `cmd_catalogue.go`)
and this feature's own contract examples. The CLI contract's "uppercase
headers" prose is stale against shipped behavior; this plan follows precedent
rather than the prose. Confirmed by the maintainer.

### D4 — Paging without a total count

`TrackStore.List` returns every tracked artifact in memory, so a total count
*could* be computed cheaply, unlike `list catalogue` (a remote, paged source with
no total). This plan reports no total anyway — a deliberate choice for output
consistency across both `list` families (the CLI contract's canonical "Catalogue"
rendering is a table + one paging line, no total), not a technical constraint.

## Key decisions

- **D1** — `Kind` is `types.KindSkill`/`types.KindAgent` directly; no parallel enum.
- **D2** — search/sort run in-memory inside `trackLister`, behind the shared
  `Lister[T]` seam; no new `TrackStore` method.
- **D3** — lowercase table headers, matching shipped precedent over the CLI
  contract's uppercase prose.
- **D4** — paging reports no total count, by choice, for output consistency with
  `list catalogue`.
- **D5** — `NewListArtifactsUseCase` returns the generic `*ListUseCase[ListArtifactsRequest,
  types.Artifact]` directly (no wrapping type; `ListArtifactsResponse` a type
  alias), and `newListArtifactsCommand` is built directly on `newCommand`
  (closing over its `kind`, never threading it through the shared constructor) —
  the same composition `list catalogue` uses at both layers, not a parallel
  implementation that merely resembles it.
- Mirror `describe provider` for `--fields` selection, `list catalogue` for the
  table/sort/paging-flag shape and the inline (no separate view file) rendering,
  the generic `ListUseCase[I, T]` composition, and the shared `newCommand`
  scaffold — there is no list-specific cobra scaffold to build on.

## Checkpoints (RED → GREEN)

| # | Milestone | Verify |
|---|---|---|
| C1 | acceptance scenarios authored and **failing** (commands absent) | `task gate-integration` fails only on the new list scenarios |
| C2 | `NewListArtifactsUseCase`/`trackLister`: failing test → green (filter/search/sort/paging window/out-of-range page/empty/read-error) | `go test ./internal/usecase/...` |
| C3 | command: failing test → green; C1 scenarios now **pass** | `go test ./internal/cmd/... && task gate-integration` |
| C4 | full gate green; coverage ≥90% | `task all` |

## Execution flow

```
T1  acceptance e2e (outer RED — fails until T3)  ─┐ parallel, worktree-isolated
T2  NewListArtifactsUseCase (RED→GREEN)           ─┘
T3  cmd (RED→GREEN) — turns T1 GREEN
T4  gates (task all, coverage ≥90%)
```

- **T1** is authored first and left failing; touches only `test/e2e/**`, drafted in
  worktree `feat/0008-list-artifacts-e2e`, merged back uncommitted.
- **T2** is file-disjoint from T1 → parallel, worktree `feat/0008-list-artifacts-usecase`.
- **T3 → T4** sequential in the working tree; T3 is where the T1 scenarios flip GREEN.

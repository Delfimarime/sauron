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
[architecture contract's "Listing use cases"](../contracts/architecture.md#listing-use-cases)
section — landed ahead of this feature, together with `list catalogue`'s
refactor onto it — fixes the generic `ListUseCase[I, T]`
(`internal/usecase/list.go`, composing `Lister[T]`/`ListWindow`/`ListResult[T]`/
`listWith[T]`) as the standing shape for any command that lists a collection, and
`newListCommand` (`internal/cmd/helper.go`) as the standing cobra scaffold every
listing leaf command builds on — it carries no notion of "kind" itself, so it is
not generic; a caller closes over its own kind (or any other bound value) in the
`bind`/`run` closures it passes in. This feature supplies the second adapter at
each layer: a `trackLister` composing `storage.TrackStore` instead of a live
registry, and `ListSkills()`/`ListAgents()` built on `newListCommand` exactly as
`ListCatalogueSkill()`/`ListCatalogueAgent()` are. No new store method is needed.

**Out of scope (YAGNI):** a total count in the paging line (matches `list
catalogue`'s own convention, by choice rather than constraint — see D4); a `kind`
column (each command is already kind-scoped, per the feature's Notes); any field
beyond `name`/`version`/`lastUpdatedAt`.

## Test-driven approach (Constitution Ch. III, Art. 1)

- **Outer loop — acceptance first.** The `test/e2e` Gherkin scenarios are authored
  first (T1) and fail (the commands don't exist); they turn green at T3.
- **Inner loop — unit first.** `ListArtifactsUseCase` (T2) and the command/view
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
  ListWindow, error)`, `Execute` calling `listWith`). `ListCatalogueUseCase`
  already **wraps** a `*ListUseCase[ListCatalogueRequest, string]` — its
  constructor builds `resolve` as the validate→get→open pipeline ending in a
  `catalogueLister`, and its own `Execute` projects the inner `ListResult[string]`
  onto `ListCatalogueResponse`. This feature's `ListArtifactsUseCase` mirrors that
  wrapping exactly, with a `trackLister` in place of `catalogueLister`.
- **The shared list-command scaffold** — `internal/cmd/helper.go`'s
  `newListCommand(use, short, long string, bind func(*cobra.Command),
  run func(ctx context.Context, stdout io.Writer) error) *cobra.Command`. It
  carries no "kind" of its own — `newCatalogueCommand(kind usecase.CatalogueKind,
  use, short, long string)` closes over `kind` in the `bind`/`run` closures it
  passes to it, and already builds `ListCatalogueSkill()`/`ListCatalogueAgent()`
  this way; this feature's `newListArtifactsCommand` mirrors that call exactly.
- **The shared paging renderer** — `internal/cmd/view_helper.go`'s
  `pagingLine(page, limit, offset int64, count int) string`, already extracted
  and used by `view_catalogue.go`; this feature's view reuses it verbatim.
- `storage.TrackStore.List(ctx) ([]types.Artifact, error)` — reads Skill+Agent
  documents, each discriminated by `TypeMeta.Kind` (`types.KindSkill`/`KindAgent`).
- `types.Artifact` / `ArtifactSpec` (`Metadata.Name`, `Spec.Version`, `Spec.UpdatedAt`).
- Shared command/view helpers: `selectFields`, `table`, `defaultOrder`/
  `validateOrder`, `pagingFlags`/`bindPagingFlags`, `newGroup`/`usageArgs`/
  `silenceFlagErrors`/`runUseCase`.

## Design (implementation map)

| File | Change |
|---|---|
| `internal/usecase/types.go` | add `ListArtifactsRequest{Kind string; ListWindow}` / `ListArtifactsResponse{Kind string; ListResult[types.Artifact]}` — the same embedding shape as `ListCatalogueRequest`/`Response` |
| `internal/usecase/usecase_list_artifacts.go` | **new** — a `trackLister{track storage.TrackStore, kind string}` implementing `Lister[types.Artifact]` (`List`: `TrackStore.List` → filter to `kind` → resolve the received `source.Option`s into a local `source.Options` value → `--search` substring on `Metadata.Name` → sort by `name`/`lastUpdatedAt` + order → window at offset/limit); `NewListArtifactsUseCase`'s `resolve` closure validates the kind and window then builds a `trackLister`, wrapping a `*ListUseCase[ListArtifactsRequest, types.Artifact]`; `ListArtifactsUseCase.Execute` delegates to the wrapped `ListUseCase.Execute` and projects `ListResult[types.Artifact]` onto `ListArtifactsResponse` — the same shape as `ListCatalogueUseCase.Execute` |
| `internal/usecase/fx.go` | provide `NewListArtifactsUseCase` |
| `internal/cmd/cmd_list_artifacts.go` | **new** — `ListSkills()`/`ListAgents()` → shared `newListArtifactsCommand(kind, use, short, long)`, itself built on `newListCommand` (closing over `kind` in its `bind`/`run` closures), binding search/sort/order/fields plus the existing shared `pagingFlags`; cobra-free `listArtifacts()` handler via `fx.Populate`/`runUseCase` |
| `internal/cmd/view_list_artifacts.go` | **new** — field set (`name`/`version`/`lastUpdatedAt`), `selectListArtifactsFields` (via shared `selectFields`), sort validation, table render, the already-shared `pagingLine` |
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
feature-specific `Lister[T]` → return it with the window), both wrapping the same
generic `*ListUseCase[I, T]`.

### D3 — Table header casing

List tables render lowercase headers — the field name verbatim (`name`,
`version`, `lastUpdatedAt`) — matching the shipped `list catalogue` view
(`headerName = "name"`, `headerKind = "kind"`) and this feature's own contract
examples. The CLI contract's "uppercase headers" prose is stale against shipped
behavior; this plan follows precedent rather than the prose. Confirmed by the
maintainer.

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
- **D5** — `ListArtifactsUseCase` wraps the generic `*ListUseCase[ListArtifactsRequest,
  types.Artifact]`, and `newListArtifactsCommand` is built on `newListCommand`
  (closing over its `kind`, never threading it through the shared constructor) —
  the same composition `list catalogue` uses at both layers, not a parallel
  implementation that merely resembles it.
- Mirror `describe provider` for `--fields` selection, `list catalogue` for the
  table/sort/paging-flag shape, the generic `ListUseCase[I, T]` composition, and
  the (non-generic) `newListCommand` scaffold.

## Checkpoints (RED → GREEN)

| # | Milestone | Verify |
|---|---|---|
| C1 | acceptance scenarios authored and **failing** (commands absent) | `task gate-integration` fails only on the new list scenarios |
| C2 | `ListArtifactsUseCase`/`trackLister`: failing test → green (filter/search/sort/paging window/out-of-range page/empty/read-error) | `go test ./internal/usecase/...` |
| C3 | command + view: failing test → green; C1 scenarios now **pass** | `go test ./internal/cmd/... && task gate-integration` |
| C4 | full gate green; coverage ≥90% | `task all` |

## Execution flow

```
T1  acceptance e2e (outer RED — fails until T3)  ─┐ parallel, worktree-isolated
T2  ListArtifactsUseCase (RED→GREEN)              ─┘
T3  cmd + view (RED→GREEN) — turns T1 GREEN
T4  gates (task all, coverage ≥90%)
```

- **T1** is authored first and left failing; touches only `test/e2e/**`, drafted in
  worktree `feat/0008-list-artifacts-e2e`, merged back uncommitted.
- **T2** is file-disjoint from T1 → parallel, worktree `feat/0008-list-artifacts-usecase`.
- **T3 → T4** sequential in the working tree; T3 is where the T1 scenarios flip GREEN.

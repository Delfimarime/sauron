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

**This is composition, not infrastructure.** `TrackStore.List` already reads and
decodes both `Skill` and `Agent` documents from `track.yaml`; no new store or
adapter is needed. This recombines `list catalogue`'s search/sort/paging/table
shape and `describe provider`'s `--fields` selection over installed (not live)
state.

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
  existing `MockBasedTrackStore`); no adapter or network fixture is needed.

## Pre-requirements (already in place)

- `storage.TrackStore.List(ctx) ([]types.Artifact, error)` — reads Skill+Agent
  documents, each discriminated by `TypeMeta.Kind` (`types.KindSkill`/`KindAgent`).
- `types.Artifact` / `ArtifactSpec` (`Metadata.Name`, `Spec.Version`, `Spec.UpdatedAt`).
- Shared command/view helpers: `selectFields`, `table`, `defaultOrder`/
  `validateOrder`, `pagingFlags`/`bindPagingFlags`, `newGroup`/`usageArgs`/
  `silenceFlagErrors`/`runUseCase`.
- Precedent for the paging shape: `ListCatalogueRequest.offset()` /
  `ListCatalogueResponse{Page, Limit, Offset}` and `view_catalogue.go`'s
  `pagingLine`.

## Design (implementation map)

| File | Change |
|---|---|
| `internal/usecase/types.go` | add `ListArtifactsRequest{Kind, Search, Sort, Order, Page, Limit}` (+ `offset()` helper, mirroring `ListCatalogueRequest`) / `ListArtifactsResponse{Kind, Items []types.Artifact, Page, Limit, Offset int64}` |
| `internal/usecase/usecase_list_artifacts.go` | **new** — `ListArtifactsUseCase`: validate `Page`/`Limit` (mirrors `ListCatalogueUseCase.validate`) → `TrackStore.List` → filter by `Kind` → `--search` substring on `Metadata.Name` → sort by `name`/`lastUpdatedAt` + order → page the in-memory slice at the computed offset |
| `internal/usecase/fx.go` | provide `NewListArtifactsUseCase` |
| `internal/cmd/cmd_list_artifacts.go` | **new** — `ListSkills()`/`ListAgents()` → shared `newListArtifactsCommand(kind, use, short, long)`, binding search/sort/order/fields plus the existing shared `pagingFlags`; cobra-free `listArtifacts()` handler via `fx.Populate`/`runUseCase` |
| `internal/cmd/view_helper.go` | **extract** `pagingLine(page, limit, offset int64, count int) string` out of `view_catalogue.go` into a shared helper so both views render the identical "showing X–Y (page N, limit M)" / "showing 0 results" line |
| `internal/cmd/view_catalogue.go` | switch to the extracted shared `pagingLine` |
| `internal/cmd/view_list_artifacts.go` | **new** — field set (`name`/`version`/`lastUpdatedAt`), `selectListArtifactsFields` (via shared `selectFields`), sort validation, table render, shared `pagingLine` |
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

### D2 — Search & sort run in memory

`track.yaml` has no `source.FileSystem`/`Option` layer (unlike the registry), so
`--search` and `--sort` run over the `[]types.Artifact` slice `TrackStore.List`
returns: `--search` is a case-insensitive `strings.Contains` on `Metadata.Name`;
`--sort lastUpdatedAt` compares `Spec.UpdatedAt` as a plain string — both are
RFC3339 UTC, which sorts lexicographically correct with no parsing needed.

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
`--page`/`--limit` window the already filtered+sorted slice locally, reusing the
same `offset()`/`Page`/`Limit`/`Offset` shape as `ListCatalogueRequest`/
`Response` and the shared `pagingLine` renderer.

## Key decisions

- **D1** — `Kind` is `types.KindSkill`/`types.KindAgent` directly; no parallel enum.
- **D2** — search/sort run in-memory over `TrackStore.List`'s result; no new store method.
- **D3** — lowercase table headers, matching shipped precedent over the CLI
  contract's uppercase prose.
- **D4** — paging reports no total count, by choice, for output consistency with
  `list catalogue`; `pagingLine` is extracted into a shared helper.
- Mirror `describe provider` for `--fields` selection, `list catalogue` for the
  table/sort/paging-flag shape.

## Checkpoints (RED → GREEN)

| # | Milestone | Verify |
|---|---|---|
| C1 | acceptance scenarios authored and **failing** (commands absent) | `task gate-integration` fails only on the new list scenarios |
| C2 | `ListArtifactsUseCase`: failing test → green (filter/search/sort/paging window/out-of-range page/empty/read-error) | `go test ./internal/usecase/...` |
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

# List Artifacts — tasks

Executable breakdown for [plan.md](plan.md), structured **test-first** per the
Constitution (Ch. III, Art. 1). Order: **T1 ∥ T2 → T3 → T4**. T1 (acceptance e2e)
and T2 (use case) open in parallel, each in its own worktree, merged back
uncommitted. The T1 scenarios are **expected to fail** until T3 lands the command.

The shared listing foundation this feature builds on — `internal/usecase/list.go`
(`Lister[T]`, `ListWindow`, `ListResult[T]`, `listWith[T]`, and the generic
`ListUseCase[I, T]`), `internal/cmd/helper.go`'s `newListCommand` (not generic —
carries no "kind" of its own; a caller closes over its own bound values in the
`bind`/`run` closures it passes in), the
[architecture contract's "Listing use cases"](../contracts/architecture.md#listing-use-cases)
section, `list catalogue`'s refactor onto both, and the shared `pagingLine` in
`internal/cmd/view_helper.go` — is **already landed**, ahead of this task
breakdown. No task here builds it; T2 wraps `ListUseCase[I, T]`, T3 calls
`newListCommand`, mirroring `ListCatalogueUseCase`/`newCatalogueCommand`'s call
sites exactly.

| Convention | Owner |
|---|---|
| production Go | `sauron-developer` / `sauron-implementing-architecture` |
| end-to-end suite | `sauron-integration-test-developer` / `sauron-implementing-integration-tests` |

---

## T1 — Acceptance scenarios (outer RED)

**Files:** `test/e2e/testdata/list_skills.feature`, `list_agents.feature`,
`test/e2e/internal/gherkin/list_artifacts_controller.go` (new, + registration in
`init.go`), `test/e2e/internal/gherkin/table_assert.go` (new — extract the
row/line matching helpers out of `catalogue_controller.go` so both controllers
share them), `test/e2e/internal/gherkin/seed_controller.go` (add the
`the track file contains:` step).

**RED:** author the scenarios, seeding `track.yaml` directly via doc-string (each
file's exact content designed in the feature, never synthesized in controller
code): lists every installed skill/agent (FR-001); default shows `name` only,
`--fields` reorders/selects columns (FR-002); paging with `--page`/`--limit`
windows the rows and reports a paging line with no total, and a page past the
end reports `showing 0 results` (FR-003); `--search` substring filter (FR-004);
`--sort lastUpdatedAt --order desc` (FR-005); nothing installed → empty result +
paging line, exit 0 (FR-006); an unreadable/malformed `track.yaml` → runtime
error, exit 1 (FR-007); a missing/invalid flag → exit 2 (FR-008). Run the suite →
the list scenarios **fail** (commands absent); the rest stay green.
**GREEN:** achieved at T3 — no production code here.

**Depends on:** none. Worktree `feat/0008-list-artifacts-e2e`, merged back uncommitted.
**Verify:** `task gate-integration` fails **only** on the new list scenarios.

---

## T2 — Use case: `trackLister` + `ListArtifactsUseCase`

**Files:** `internal/usecase/types.go`, `internal/usecase/usecase_list_artifacts.go`
(new), `usecase_list_artifacts_test.go` (new), `internal/usecase/fx.go`.

**RED:** write `usecase_list_artifacts_test.go` first (mocking `storage.TrackStore`
via `MockBasedTrackStore`, driving `trackLister` and `ListArtifactsUseCase`
exactly as `usecase_list_catalogue_test.go` drives `catalogueLister`): a mixed
Skill/Agent track filters to the requested kind; `--search` keeps only matching
names (case-insensitive); `--sort name` and `--sort lastUpdatedAt`, each with
`asc`/`desc`, order correctly; `--page`/`--limit` window the filtered+sorted
result and report the applied `Page`/`Limit`/`Offset` (via the shared
`ListResult[types.Artifact]`); a page past the end returns an empty window with
no error; `Page` or `Limit` below `1` is a usage error (already covered by
`listWith`'s own test — assert only that `Execute` surfaces it, not re-derive
the boundary); no installed artifact of the kind → empty `Items`, no error; a
`TrackStore.List` failure → classified io error. It fails (use case absent).
**GREEN:** implement `ListArtifactsRequest{Kind string; ListWindow}` /
`ListArtifactsResponse{Kind string; ListResult[types.Artifact]}` in `types.go`;
implement `trackLister{track storage.TrackStore, kind string}` satisfying
`Lister[types.Artifact]` (`List`: `TrackStore.List` → `ioErr` on failure (FR-007)
→ filter to `kind` (FR-001) → resolve the received `source.Option`s into a local
`source.Options` → `--search` substring (FR-004) → sort by `name`/`lastUpdatedAt`
+ order (FR-005) → window at offset/limit (FR-003)); `NewListArtifactsUseCase`
builds a `resolve` closure (validate kind → `in.ListWindow.validate()` → build
the `trackLister`) and wraps it as `*ListUseCase[ListArtifactsRequest,
types.Artifact]` via `NewListUseCase`, exactly as `NewListCatalogueUseCase` does;
`ListArtifactsUseCase.Execute` calls the wrapped `ListUseCase.Execute` and
projects the returned `ListResult[types.Artifact]` onto `ListArtifactsResponse`.
Provide via `fx.go`. gocognit ≤15.

**Depends on:** none (parallel with T1; depends only on the already-landed
`internal/usecase/list.go`). Worktree `feat/0008-list-artifacts-usecase`, merged
back uncommitted.
**Verify:** `go test ./internal/usecase/...` — green.

---

## T3 — Command + view: `list skills`/`list agents` (turns T1 GREEN)

**Files:** `internal/cmd/cmd_list_artifacts.go` (new), `view_list_artifacts.go`
(new), `cmd_list_artifacts_test.go`, `view_list_artifacts_test.go`,
`internal/cmd/cmd_list.go`.

**RED:** write the command/view tests first — `list skills`/`list agents` render
the `name`/`version`/`lastUpdatedAt` table (default = `name` only, per FR-002),
followed by the shared `pagingLine`; `--fields` reorders/selects columns, `name`
always first; `--search`/`--sort`/`--order`/`--page`/`--limit` reach the use
case; an unknown `--fields`/`--sort`/`--order` value or an out-of-range
`--page`/`--limit` is a usage error (exit 2, FR-008); an empty result renders no
table row, only the paging line (FR-006). They fail (command absent).
**GREEN:** `ListSkills()`/`ListAgents()` → `newListArtifactsCommand(kind, use,
short, long)`, itself a thin wrapper over `newListCommand` (`bind` wires
search/sort/order/fields plus the existing shared `pagingFlags`; `run` closes
over `kind` and delegates to `listArtifacts`) — exactly how `newCatalogueCommand`
calls `newListCommand`; cobra-free `listArtifacts(ctx, kind, flags, stdout)` via
`fx.Populate`/`runUseCase`; `view_list_artifacts.go` renders the table + the
already-shared `pagingLine` per
[list-skills](contracts/list-skills.md)/[list-agents](contracts/list-agents.md);
register both under `List()` in `cmd_list.go`.

**Depends on:** T2.
**Verify:** `go build ./... && go test ./internal/cmd/...` green, **and** the T1
scenarios now pass: `task gate-integration`.

---

## T4 — Gates

**Files:** none (verification only); address any lint/coverage findings above.

**Depends on:** T3.
**Verify:** `task all` (build, test, gate-coverage ≥90%/floor 80%, gate-lint,
gate-security, gate-integration) — fully green.

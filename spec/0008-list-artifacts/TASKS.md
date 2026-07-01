# List Artifacts — tasks

Executable breakdown for [plan.md](plan.md), structured **test-first** per the
Constitution (Ch. III, Art. 1). Order: **T1 ∥ T2 → T3 → T4**. T1 (acceptance e2e)
and T2 (use case) open in parallel, each in its own worktree, merged back
uncommitted. The T1 scenarios are **expected to fail** until T3 lands the command.

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

## T2 — Use case: `ListArtifactsUseCase`

**Files:** `internal/usecase/types.go`, `internal/usecase/usecase_list_artifacts.go`
(new), `usecase_list_artifacts_test.go` (new), `internal/usecase/fx.go`.

**RED:** write `usecase_list_artifacts_test.go` first (mocking `storage.TrackStore`
via `MockBasedTrackStore`): a mixed Skill/Agent track filters to the requested
kind; `--search` keeps only matching names (case-insensitive); `--sort name` and
`--sort lastUpdatedAt`, each with `asc`/`desc`, order correctly; `--page`/
`--limit` window the filtered+sorted result and report the applied
`Page`/`Limit`/`Offset`; a page past the end returns an empty window with no
error; `Page` or `Limit` below `1` is a usage error; no installed artifact of the
kind → empty `Items`, no error; a `TrackStore.List` failure → classified io
error. It fails (use case absent).
**GREEN:** implement `ListArtifactsRequest`/`ListArtifactsResponse` and
`ListArtifactsUseCase.Execute` per the [design](plan.md#design-implementation-map):
validate `Page`/`Limit` → `TrackStore.List` → classified io error on failure
(FR-007) → filter by `Kind` → `--search` substring (FR-004) → sort by
`name`/`lastUpdatedAt` + order (FR-005) → page at the computed offset (FR-003).
Provide via `fx.go`. gocognit ≤15.

**Depends on:** none (parallel with T1). Worktree `feat/0008-list-artifacts-usecase`,
merged back uncommitted.
**Verify:** `go test ./internal/usecase/...` — green.

---

## T3 — Command + view: `list skills`/`list agents` (turns T1 GREEN)

**Files:** `internal/cmd/cmd_list_artifacts.go` (new), `view_list_artifacts.go`
(new), `cmd_list_artifacts_test.go`, `view_list_artifacts_test.go`,
`internal/cmd/cmd_list.go`, `internal/cmd/view_helper.go` (extract the shared
`pagingLine`), `internal/cmd/view_catalogue.go` (switch to it),
`internal/cmd/view_catalogue_test.go` (unchanged behavior, still green).

**RED:** write the command/view tests first — `list skills`/`list agents` render
the `name`/`version`/`lastUpdatedAt` table (default = `name` only, per FR-002),
followed by the paging line; `--fields` reorders/selects columns, `name` always
first; `--search`/`--sort`/`--order`/`--page`/`--limit` reach the use case; an
unknown `--fields`/`--sort`/`--order` value or an out-of-range `--page`/`--limit`
is a usage error (exit 2, FR-008); an empty result renders no table row, only the
paging line (FR-006). They fail (command absent).
**GREEN:** `ListSkills()`/`ListAgents()` → `newListArtifactsCommand(kind, use,
short, long)` (`Args: usageArgs(cobra.NoArgs)`), binding search/sort/order/fields
plus the existing shared `pagingFlags`/`bindPagingFlags`; cobra-free
`listArtifacts(ctx, kind, flags, stdout)` via `fx.Populate`/`runUseCase`;
`view_list_artifacts.go` renders the table + shared `pagingLine` per
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

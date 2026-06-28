# Install Artifacts — tasks

Executable breakdown for [plan.md](plan.md), structured **test-first** per the
Constitution (Ch. III, Art. 1): each task writes its failing test before the code
that makes it pass. Order: **T1 ∥ T2 → (T3 ∥ T4) → T5 → T6 → T7**. T1 (acceptance
e2e) and T2 (port) open in parallel; T3 (git) and T4 (http) run in parallel git
worktrees, each merged back uncommitted. The T1 acceptance scenarios are **expected to
fail** until T6 lands the command.

| Convention | Owner |
|---|---|
| production Go | `sauron-developer` / `sauron-implementing-architecture` |
| end-to-end suite | `sauron-integration-test-developer` / `sauron-implementing-integration-tests` |

Each task below names its **RED** (the failing test to write first), its **GREEN**
(the minimum code to pass), and the **Verify** command. The registry roots are already
`skills/`/`agents/` (dot-less) — no task renames them.

---

## T1 — Acceptance scenarios (outer RED)

**Files:** `test/e2e/testdata/install_skill.feature`, `install_agent.feature`,
`test/e2e/internal/gherkin/install_controller.go` (+ registration in `init.go`),
`test/e2e/internal/runtime/httpregistry/server.go` (confirm the `/{kind}/{name}/content`
archive shape the adapter will unpack).

**RED:** author the scenarios over the in-process http fixture, driving the built
binary (`SAURON_BIN`), asserting via `pkg/`: install a skill/agent (`+`, tracked with
`version`/`path`), re-install unchanged (no-op) and changed (`~`), a name the registry
does not offer (reported, run continues, others install), no provider set (runtime
error, exit 1), missing argument (exit 2). Run the suite → the install scenarios
**fail** (command absent); the rest stay green.
**GREEN:** achieved at T6 — do not implement production code here.

**Depends on:** none. Worktree `feat/0007-install-e2e`, merged back uncommitted.
**Verify:** `task gate-integration` fails **only** on the new install scenarios (everything else still passes).

---

## T2 — Port: `Fetch` contract + provider-path helper (foundation)

**Files:** `pkg/sauron/source/source.go`, `pkg/sauron/source/mock_based_file_system.go`,
`internal/usecase/helper.go`, `internal/usecase/helper_test.go`.

**RED:** write `helper_test.go` asserting `installPath(KindSkill,"go") == "skills/sauron-go"`
and the agent case; it fails to compile (`installPath` absent).
**GREEN:** add `Fetch(ctx, uri) ([]File, error)` to `source.FileSystem` (a fetched
`File.Name()` is artifact-relative); extend the source mock; add `installPath(kind,
name)` shared with `migrate`'s `providerDirs`.

**Depends on:** none (parallel with T1).
**Verify:** `go build ./... && go test ./internal/usecase/... -run InstallPath` — helper passes; the source mock satisfies `FileSystem`.

---

## T3 — Git transport: in-memory go-git source (parallel)

**Files:** `internal/infrastructure/repository/registry/git_filesystem.go`,
`git_tree_source.go` (new), `api/paging.go` (new), `fx.go`, `*_test.go`; **delete**
`api/directory.go` + `api/directory_test.go`.

**RED:** write the git source tests first — `List` paging/search/sort parity over
`skills/`/`agents/`, `Fetch` returns the full tree with artifact-relative paths,
`Version()` equals the artifact dir's tree-object hash; they fail (source not built).
**GREEN:** clone into go-git in-memory storage + memfs (remove `MkdirTemp`/`RemoveAll`/
`cleanupWhenDone` and the pond pool if now unused); resolve the revision; implement the
git-tree-backed `source.FileSystem` (`List` paged via the extracted `api/paging.go`
helper, `Fetch` recursive blob walk, `Version()` from `tree.Hash`). gocognit ≤15; add a
`ponytail:` comment naming the in-memory ceiling.

**Depends on:** T2. Worktree `feat/0007-install-git`, merged back uncommitted.
**Verify:** `go test ./internal/infrastructure/repository/registry/...` — green; no temp dir or goroutine remains.

---

## T4 — HTTP transport: archive fetch + header version (parallel)

**Files:** `pkg/sauron/marketplace/resources.go`, `client.go`, `types.go`,
`marketplace_test.go`; `internal/infrastructure/repository/registry/rest_filesystem.go`,
`rest_filesystem_test.go`.

**RED:** write the tests first — `Content` issues `GET /{kind}/{name}/content` and reads
the `Artifact-Version` header (via `httptest`); `restFileSystem.Fetch` unpacks the
archive to the expected tree; an artifact with no declared version is surfaced so
install can skip it. They fail (methods not implemented).
**GREEN:** add `Content(ctx, name) (archive, version, error)` on the `ArtifactClient`;
implement `restFileSystem.Fetch` (download, unpack gzip/tar into artifact-relative
`File`s, carry the header version). gocognit ≤15.

**Depends on:** T2. Worktree `feat/0007-install-http`, merged back uncommitted.
**Verify:** `go test ./internal/infrastructure/repository/registry/... ./pkg/sauron/marketplace/...` — green.

---

## T5 — Use case: `InstallUseCase`

**Files:** `internal/usecase/usecase_install.go` (new), `usecase_install_test.go`
(new), `internal/usecase/fx.go`.

**RED:** write `usecase_install_test.go` first (mocking open-registry, `source.FileSystem`,
`ProvidersStore`, `TrackStore`): fresh name adds and records the source `version` +
`<kind>/sauron-<name>` path; a changed `version` updates and bumps `updatedAt`;
unchanged `version` is a no-op; no provider → runtime error and nothing written; an
unoffered name is recorded as failed while siblings install. It fails (use case absent).
**GREEN:** implement `InstallUseCase` (generic `UseCase` shape) per the
[Execute flow](plan.md): provider presence (FR-005), open registry (FR-007), per-name
resolve `<skills|agents>/<name>` (FR-006), version (http none → skip), `Fetch` + write
under the `name:"provider"` fs at `installPath`, `TrackStore.Update`, reconcile on
`version`; return `InstallResult{Added, Updated, Failures}`. Provide via `fx.go`. gocognit ≤15.

**Depends on:** T3, T4.
**Verify:** `go test ./internal/usecase/...` — green.

---

## T6 — Command + view: `install skill|agent` (turns T1 GREEN)

**Files:** `internal/cmd/cmd_install.go` (new), `view_install.go` (new),
`cmd_install_test.go`, `view_install_test.go`, `internal/cmd/cmd_root.go`.

**RED:** write `cmd_install_test.go`/`view_install_test.go` first — `install skill a b`
renders the `+` plan + `2 added`; an already-current name renders no change; missing
name → exit 2; the kind heading (`skills:`/`agents:`) matches the invoked command. They
fail (command absent).
**GREEN:** `Install()` parent + `InstallSkill()`/`InstallAgent()` →
`newInstallCommand(kind, use, short, long)` (`Args: usageArgs(cobra.MinimumNArgs(1))`)
+ cobra-free `install(ctx, kind, names, stdout)` via `fx.Populate`/`runUseCase`;
`view_install.go` renders the plan per the
[install-skill](contracts/install-skill.md)/[install-agent](contracts/install-agent.md)
contracts; register under the root command.

**Depends on:** T5.
**Verify:** `go build ./... && go test ./internal/cmd/...` green, **and** the T1
acceptance scenarios now pass: `task gate-integration`.

---

## T7 — Gates

**Files:** none (verification only); address any lint/coverage findings above.

**Depends on:** T6.
**Verify:** `task all` (build, test, gate-coverage ≥90%/floor 80%, gate-lint, gate-security, gate-integration) — fully green.

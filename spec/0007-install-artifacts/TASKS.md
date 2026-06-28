# Install Artifacts — tasks

Executable breakdown for [plan.md](plan.md). Each task owns its files, states a
single pass/fail verification, and lists its dependencies. Order:
**T1 → (T2 ∥ T3) → T4 → T5 → T6 → T7**. T2 (git) and T3 (http) are file-disjoint and
run in parallel git worktrees, each merged back into the working tree uncommitted. T6
touches only `test/e2e/**` and verifies after T5.

| Convention | Owner |
|---|---|
| production Go | `sauron-developer` / `sauron-implementing-architecture` |
| end-to-end suite | `sauron-integration-test-developer` / `sauron-implementing-integration-tests` |

---

## T1 — Port: `Fetch` contract + provider-path helper

**Files:** `pkg/sauron/source/source.go`, `pkg/sauron/source/mock_based_file_system.go`,
`internal/usecase/helper.go`, `internal/usecase/helper_test.go`.

Add `Fetch(ctx, uri) ([]File, error)` to `source.FileSystem` (a fetched `File.Name()`
is its path relative to the artifact directory); regenerate/extend the mock. Add
`installPath(kind, name) => "<kind>/sauron-<name>"` shared with `migrate`'s
`providerDirs`.

**Depends on:** none.
**Verify:** `go build ./...` — interface and helper compile; the source mock satisfies `FileSystem`.

---

## T2 — Git transport: in-memory go-git source (parallel)

**Files:** `internal/infrastructure/repository/registry/git_filesystem.go`,
`git_tree_source.go` (new), `api/paging.go` (new), `fx.go`, `*_test.go`; **delete**
`api/directory.go` + `api/directory_test.go`.

Clone into go-git in-memory storage + memfs (remove `MkdirTemp`/`RemoveAll`/
`cleanupWhenDone` and the pond pool if now unused); resolve the revision. New
git-tree-backed `source.FileSystem`: `List` (immediate children, paged via the
extracted `api/paging.go` `page`/`filter`/sort helper), `Fetch` (recursive blob walk →
artifact-relative `File`s), entries whose `Version()` returns the relevant `tree.Hash`.
gocognit ≤15. Add a `ponytail:` comment naming the in-memory ceiling.

**Depends on:** T1. Worktree `feat/0007-install-git`, merged back uncommitted.
**Verify:** `go test ./internal/infrastructure/repository/registry/...` — List preserves the old paging/search/sort parity; `Fetch` returns the full tree with relative paths; `Version()` equals the artifact dir's tree-object hash; no temp dir or goroutine remains.

---

## T3 — HTTP transport: archive fetch + header version (parallel)

**Files:** `pkg/sauron/marketplace/resources.go`, `client.go`, `types.go`,
`marketplace_test.go`; `internal/infrastructure/repository/registry/rest_filesystem.go`,
`rest_filesystem_test.go`.

Add `Content(ctx, name) (archive, version, error)` on the `ArtifactClient`
(`GET /{kind}/{name}/content`, read the `Artifact-Version` header). Implement
`restFileSystem.Fetch`: download the archive, unpack gzip/tar into artifact-relative
`File`s, carry the header version. gocognit ≤15.

**Depends on:** T1. Worktree `feat/0007-install-http`, merged back uncommitted.
**Verify:** `go test ./internal/infrastructure/repository/registry/... ./pkg/sauron/marketplace/...` — `Content` issues the right request and reads the version header; `Fetch` unpacks the archive to the expected tree; an artifact with no declared version is surfaced so install can skip it.

---

## T4 — Use case: `InstallUseCase`

**Files:** `internal/usecase/usecase_install.go` (new), `usecase_install_test.go`
(new), `internal/usecase/fx.go`.

`InstallUseCase` (generic `UseCase` shape) composing `OpenRegistryUseCase` +
`ProvidersStore.Get` + `TrackStore`: provider absent → runtime error, install nothing
(FR-005); registry unreachable → runtime error (FR-007); per name resolve
`<.skills|.agents>/<name>`, not offered → per-name failure, continue (FR-006); read
`version` (http none → skip); `Fetch` and write under the `name:"provider"` fs at
`installPath(kind,name)`; `TrackStore.Update` with `version`/`path`/`installedAt`/
`updatedAt`; reconcile on `version` (add `+` / update `~` / unchanged no-op).
Returns `InstallResult{Added, Updated, Failures}`. gocognit ≤15.

**Depends on:** T2, T3.
**Verify:** `go test ./internal/usecase/...` — fresh name adds and records the source `version` + `<kind>/sauron-<name>` path; a re-install with a changed `version` updates and bumps `updatedAt`; unchanged `version` is a no-op; no provider → runtime error and nothing written; an unoffered name is recorded as failed while siblings install.

---

## T5 — Command + view: `install skill|agent`

**Files:** `internal/cmd/cmd_install.go` (new), `view_install.go` (new),
`cmd_install_test.go`, `view_install_test.go`, `internal/cmd/cmd_root.go`.

`Install()` parent + `InstallSkill()`/`InstallAgent()` → `newInstallCommand(kind, use,
short, long)` (`Args: usageArgs(cobra.MinimumNArgs(1))`) + cobra-free `install(ctx,
kind, names, stdout)` via `fx.Populate`/`runUseCase`. `view_install.go` renders the
`+`/`~` plan under the kind heading (`skills:` or `agents:`) with the summary count,
per the [install-skill](contracts/install-skill.md)/[install-agent](contracts/install-agent.md)
contracts. Missing args → exit 2 (FR-008). Register under the root command.

**Depends on:** T4.
**Verify:** `go build ./... && go test ./internal/cmd/...` — `install skill a b` renders the `+` plan + `2 added`; an already-current name renders no change; missing name → exit 2; the kind heading matches the invoked command.

---

## T6 — End-to-end: `install_*.feature`

**Files:** `test/e2e/testdata/install_skill.feature`, `install_agent.feature`,
`test/e2e/internal/gherkin/install_controller.go` (+ registration in `init.go`),
`test/e2e/internal/runtime/httpregistry/server.go` (ensure the `/{kind}/{name}/content`
archive matches the adapter's unpack).

Scenarios over the in-process http fixture, driving the built binary (`SAURON_BIN`),
asserting via `pkg/`: install a skill/agent (`+`, tracked with `version`/`path`),
re-install unchanged (no-op) and changed (`~`), a name the registry does not offer
(reported, run continues, others install), no provider set (runtime error, exit 1),
missing argument (exit 2).

**Depends on:** T5 (passes only once T5 lands). May be drafted in a worktree
(`feat/0007-install-e2e`, merged back uncommitted).
**Verify:** `task gate-integration`.

---

## T7 — Gates

**Files:** none (verification only); address any lint/coverage findings above.

**Depends on:** T6.
**Verify:** `task all` (build, test, gate-coverage ≥90%/floor 80%, gate-lint, gate-security, gate-integration).

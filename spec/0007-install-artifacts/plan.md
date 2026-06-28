# Install Artifacts — implementation plan

How [`install`](spec.md) is built. The executable task breakdown lives in
[TASKS.md](TASKS.md). Code follows the
[architecture contract](../contracts/architecture.md) and the
`sauron-implementing-architecture` conventions; the end-to-end suite follows
`sauron-implementing-integration-tests`.

## Goal & scope

Add `install skill|agent <name>...`: fetch each named artifact's full directory from
the configured registry, place it under the active provider at `<kind>/sauron-<name>`,
and record a `Skill`/`Agent` track document carrying its source-read `version`,
`path`, and `installedAt`/`updatedAt` (FR-001/FR-002). Reconcile an already installed
artifact on `version` change (FR-003), print a `+`/`~` plan with a summary count
(FR-004), fail with a runtime error and install nothing when no provider is set
(FR-005), report-and-continue on a name the registry does not offer (FR-006), fail
with a runtime error when the registry is unreachable (FR-007), and exit 2 on
missing/invalid arguments (FR-008).

**This is infrastructure work, not composition.** Both transports today only `List`
immediate entries — `Get`/`Read` are `ErrNotImplemented`, the marketplace client has
no download method, there is no recursive content fetch, and `Stat.Version()` returns
`""`. Install must build the **content-fetch path** (git: walk the in-memory clone's
tree; http: download + unpack the `/{kind}/{name}/content` archive) and **version
derivation** (git: tree-object hash; http: `Artifact-Version` header) for both
transports. The use case and command compose on top.

**Out of scope (YAGNI):** `Describe`/single-file `Get` on the port beyond what fetch
needs; sync/upgrade (0011/0012); local-drift detection (dropped with `digest`);
persona; multi-registry; any provider beyond `claude`/`zencoder`.

## Pre-requirements

- `types.Artifact`/`ArtifactSpec` with the single `version` identity already exist
  ([`pkg/sauron/types/artifact.go`](../../pkg/sauron/types/artifact.go)).
- `TrackStore.Update` (upsert by `(kind,name)`), `ProvidersStore.Get`,
  `OpenRegistryUseCase` (live source open; unreachable surfaces as a runtime error),
  the `name:"provider"` afero.Fs and `providerDirs` map, and the `marketplace` client
  `List` already exist. The catalogue roots `.skills`/`.agents` are reused.

## Design

| File | Change |
|---|---|
| `pkg/sauron/source/source.go` | add `Fetch(ctx, uri) ([]File, error)` to `FileSystem`; a fetched `File.Name()` is its path **relative to the artifact directory** |
| `internal/usecase/helper.go` | `installPath(kind, name) => "<kind>/sauron-<name>"` (`kind` ∈ `skills`,`agents`), shared with `migrate`'s `providerDirs` |
| `internal/infrastructure/repository/registry/git_filesystem.go` | clone into go-git **in-memory** storage + memfs (drop `MkdirTemp`/`RemoveAll`/`cleanupWhenDone` and the pond pool if unused); resolve the revision; return a git-tree-backed source |
| `internal/infrastructure/repository/registry/git_tree_source.go` | **new** — `source.FileSystem` over go-git tree objects: `List` (paged via the shared helper), `Fetch` (recursive blob walk), entries whose `Version()` is the relevant `tree.Hash` |
| `internal/infrastructure/repository/registry/api/directory.go` + test | **delete** (sole caller was git); **extract** `page`/`filter`/sort into `api/paging.go` over `[]source.File` |
| `internal/infrastructure/repository/registry/fx.go` | drop the pool injection into the git factory if unused |
| `pkg/sauron/marketplace/resources.go`, `client.go`, `types.go` | add `Content(ctx, name) (archive, version, error)` (`GET /{kind}/{name}/content`, read `Artifact-Version`) |
| `internal/infrastructure/repository/registry/rest_filesystem.go` | implement `Fetch` (download via client, unpack gzip/tar → relative `File`s); version from the header |
| `internal/usecase/usecase_install.go` | **new** — `InstallUseCase` composing open-registry + provider presence + per-name fetch/version/write/track + reconcile |
| `internal/usecase/fx.go` | provide `NewInstallUseCase` |
| `internal/cmd/cmd_install.go` | **new** — `Install()` parent + `InstallSkill()`/`InstallAgent()` → `newInstallCommand(kind,…)` + cobra-free `install()` handler via `fx.Populate` |
| `internal/cmd/view_install.go` | **new** — `+`/`~` plan under the kind heading + summary count (per the CLI contracts) |
| `internal/cmd/cmd_root.go` | register `Install()` |
| `test/e2e/testdata/install_*.feature` + `.../gherkin/` | **new** — install/reconcile/not-offered/no-provider/exit codes over the http fixture |

### Provider directory layout (D4)

claude → `~/.claude`, zencoder → `~/.zencoder`; each artifact lives under its kind's
subdir (`skills/` or `agents/`) as `sauron-<name>`. `spec.path = "<kind>/sauron-<name>"`
is recorded relative to the provider home, so `migrate` moves `<home>/<path>` verbatim
and install writes through the same `name:"provider"` fs.

### Git in-memory source (D2)

The git factory clones into go-git's in-memory storage + memfs (no OS temp dir),
retains the repo, and backs the git `source.FileSystem` with go-git tree objects:
`Stat.Version()` returns the artifact directory's `tree.Hash` (native — no afero, no
recompute), and `Fetch` walks the tree's blobs. This deletes the afero
`api/directory.go`, the temp-dir handling, and the `cleanupWhenDone` pond goroutine.
The catalogue's List/paging/search/sort behaviour is preserved by the extracted
slicing helper.

### Content-fetch port shape (D1)

`Fetch(ctx, uri) ([]source.File, error)` returns the artifact's full tree as
artifact-relative `File`s. Keeps the `pkg/` port `afero`-free (rejected: returning an
`afero.Fs` — cleaner copy code but leaks infrastructure into the port).

### `InstallUseCase.Execute` flow

1. `ProvidersStore.Get` the active provider; absent → runtime error, install nothing
   (FR-005).
2. Open the registry via `OpenRegistryUseCase`; unreachable → runtime error (FR-007).
3. For each name: resolve the source dir `<.skills|.agents>/<name>`; not offered →
   record a per-name failure and continue (FR-006); read `version` (http with none →
   skip per [versioning FR-005](capabilities/artifact-versioning.md)); `Fetch` the
   tree and write it under the provider fs at `<kind>/sauron-<name>`; `TrackStore.Update`
   with `version`/`path`/`installedAt`/`updatedAt`.
4. Reconcile on `version` (D5): absent in track → `+` add; present & changed → `~`
   update (rewrite + bump `updatedAt`); present & unchanged → no-op (current).
5. Return `InstallResult{Added, Updated, Failures}` for the view.

## Key decisions

- **D1 content-fetch port** — `Fetch([]source.File)`, port stays afero-free.
- **D2 git version** — in-memory go-git; `tree.Hash` is the native tree-object hash;
  deletes the afero `Directory`, temp dir, and cleanup goroutine.
- **D3 http version** — declared object version (`Artifact-Version` header / listing);
  an http artifact with none is reported and skipped.
- **D4 provider layout** — `spec.path = "<kind>/sauron-<name>"`, reusing the
  documented claude/zencoder `skills/`+`agents/` layout.
- **D5 reconcile** — key `(kind,name)`; add/update on `version`, no-op when unchanged;
  per-name failure continues.
- **Mirror `catalogue`** for the command/use-case/view shape
  (`Parent()` + `<Verb><Kind>()` + `new<Verb>Command`).

## Notes

- **Memory ceiling (D2).** The in-memory clone holds the repo in RAM; a shallow
  depth-1 clone of a text registry is small. Mark with a `ponytail:` comment naming
  the ceiling (move to a bounded on-disk clone only if a registry is ever large).
- **FR-002 is now `version`, not `digest`.** The state contract and schema record a
  single source-read `version`; install owns `spec.path` and the timestamps.

## Checkpoints

| # | Milestone | Verify |
|---|---|---|
| C1 | port `Fetch` + path helper compile | `go build ./...` |
| C2 | git in-memory source: List parity, `Fetch` tree, `Version`==tree hash | `go test ./internal/infrastructure/repository/registry/...` |
| C3 | http: client `Content`, `Fetch` unpack, version from header | `go test ./internal/infrastructure/repository/registry/... ./pkg/sauron/marketplace/...` |
| C4 | `InstallUseCase` unit-tested (add/update/no-op, no-provider, per-name failure, version recorded) | `go test ./internal/usecase/...` |
| C5 | command + view wired; whole tree builds; style/coverage held | `go build ./... && go test ./internal/cmd/... && task gate-lint && task gate-coverage` |
| C6 | `install_*.feature` pass end-to-end; full gate green | `task all` |

## Execution flow

Sequential chain with one parallel fan-out:
**T1 → (T2 ∥ T3) → T4 → T5 → T6 → T7** (see [TASKS.md](TASKS.md)). T2 (git) and T3
(http) are file-disjoint and run in **parallel git worktrees** (`feat/0007-install-git`,
`feat/0007-install-http`), each merged back into the working tree uncommitted. T6
touches only `test/e2e/**` and may be drafted in a worktree but passes only once T5
lands.

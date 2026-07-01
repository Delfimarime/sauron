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

## Test-driven approach (Constitution Ch. III, Art. 1)

Every behaviour is pinned by a test **written to fail before** the code that makes it
pass; coverage target 90%, hard floor 80%.

- **Outer loop — acceptance first.** The user-observable behaviour is pinned by the
  `test/e2e` Gherkin scenarios (Art. 6). They are authored **first** (T1) and **fail**
  (`install` does not yet exist); `task gate-integration` stays RED for these scenarios
  until the command lands at T6, then goes GREEN. This is the acceptance harness the
  inner loops drive toward.
- **Inner loops — unit first.** Each implementation task (T2–T6) starts by writing the
  failing unit test(s) that pin its slice (port contract, git source, http source, use
  case, command/view), watching them fail (RED), then writing the minimum code to pass
  (GREEN), then refactoring under green. No production line is added before a test
  demands it.
- **Mocks at the seams.** Use-case tests mock the `source.FileSystem`, stores, and the
  open-registry step (`mock_based_*`); the transport tests use in-memory/`httptest`
  fixtures — never the network or the real FS.

## Pre-requirements (already in place)

- The registry's artifact roots are `skills/` and `agents/` (dot-less); both transports
  key on `rootSkills`/`rootAgents`. Install reuses them verbatim.
- `types.Artifact`/`ArtifactSpec` with the single `version` identity exist
  ([`pkg/sauron/types/artifact.go`](../../pkg/sauron/types/artifact.go)).
- `TrackStore.Update` (upsert by `(kind,name)`), `ProvidersStore.Get`,
  `OpenRegistryUseCase` (live source open; unreachable surfaces as a runtime error),
  the `name:"provider"` afero.Fs and `providerDirs` map, and the `marketplace` client
  `List` exist.

## Design (implementation map)

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
| `test/e2e/testdata/install_*.feature` + `.../gherkin/` | **new (authored first, T1)** — install/reconcile/not-offered/no-provider/exit codes over the http fixture |

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
3. For each name: resolve the source dir `<skills|agents>/<name>`; not offered →
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
- **FR-002 is `version`, not `digest`.** The state contract and schema record a single
  source-read `version`; install owns `spec.path` and the timestamps.

## Checkpoints (RED → GREEN)

| # | Milestone | Verify |
|---|---|---|
| C1 | install acceptance scenarios authored and **failing** (command absent) | `task gate-integration` fails only on the new install scenarios |
| C2 | port `Fetch` + path helper: failing test → green; tree builds | `go build ./... && go test ./pkg/sauron/source/... ./internal/usecase/...` (helper) |
| C3 | git in-memory source: failing tests → green (List parity, `Fetch` tree, `Version`==tree hash) | `go test ./internal/infrastructure/repository/registry/...` |
| C4 | http: failing tests → green (`Content`, `Fetch` unpack, header version) | `go test ./internal/infrastructure/repository/registry/... ./pkg/sauron/marketplace/...` |
| C5 | `InstallUseCase`: failing test → green (add/update/no-op, no-provider, per-name failure) | `go test ./internal/usecase/...` |
| C6 | command + view: failing test → green; the C1 acceptance scenarios now **pass** | `go test ./internal/cmd/... && task gate-integration` |
| C7 | full gate green; coverage ≥90% | `task all` |

## Execution flow

```
T1  acceptance e2e (outer RED — fails until T6)   ─┐ may run alongside T2
T2  port Fetch + path helper (RED→GREEN)          ─┘ (foundation)
        ├─ T3  git in-memory source (RED→GREEN)    ┐ parallel,
        └─ T4  http archive fetch  (RED→GREEN)     ┘ worktree-isolated
T5  InstallUseCase (RED→GREEN)
T6  cmd + view (RED→GREEN) — turns T1 GREEN
T7  gates (task all, coverage ≥90%)
```

- **T1** (acceptance) is authored first and left failing; touches only `test/e2e/**`
  and may be drafted in a worktree (`feat/0007-install-e2e`), merged back uncommitted.
- **T2** is the foundation (the `Fetch` port shape) everything depends on.
- **T3** and **T4** are file-disjoint (git vs http/marketplace) → **parallel git
  worktrees** (`feat/0007-install-git`, `feat/0007-install-http`), each merged back
  uncommitted.
- **T5 → T6 → T7** sequential in the working tree; T6 is where the T1 acceptance
  scenarios flip to GREEN.

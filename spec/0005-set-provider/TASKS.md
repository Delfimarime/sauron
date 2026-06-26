# Set Provider — tasks

Executable breakdown for [plan.md](plan.md). Each task owns its files, states a
single pass/fail verification, and lists its dependencies. Order: **T1 → T2 → T3
→ T4 → T5 → T6 → T7**. T6 may be drafted in parallel (worktree on `test/e2e/**`)
but verifies only after T5.

| Convention | Owner |
|---|---|
| production Go | `sauron-developer` / `sauron-implementing-architecture` |
| end-to-end suite | `sauron-integration-test-developer` / `sauron-implementing-integration-tests` |

---

## T1 — Types: unify Skill/Agent into `Artifact`

**Files:** `pkg/sauron/types/artifact.go` (new), `pkg/sauron/types/manifest.go`,
`pkg/sauron/types/types_test.go`; delete `skill.go`, `agent.go`.

Single `Artifact` (TypeMeta + Metadata + `ArtifactSpec`); move `ArtifactSpec`
here. Skill vs agent is read from the embedded `TypeMeta.Kind` — no extra
discriminator field. Keep `KindSkill`/`KindAgent` and the two JSON schemas
unchanged.

**Depends on:** none.
**Verify:** `go test ./pkg/...` — `Artifact` round-trips through YAML preserving `kind: Skill|Agent`.

---

## T2 — Storage: routes, `Upsert`, `ProvidersStore`, `TrackStore`

**Files:** `internal/infrastructure/repository/storage/store.go`,
`providers_store.go` (new), `track_store.go` (new), `fx.go`,
`mock_based_track_store.go` (new), plus `_test.go` for each store and
`upsert_test.go`.

- `store.go`: add `KindProvider → settings.yaml`, `KindSkill`/`KindAgent →
  track.yaml`; add `Upsert(ctx, kind, name, doc)` (locked replace-matching-or-append).
- `ProvidersStore`: `Get` (via `First`) / `Set` (via `Replace`, stamping `TypeMeta`).
- `TrackStore`: `List(ctx) ([]types.Artifact, error)` (each `Artifact` carries its
  `TypeMeta.Kind`) and `Update(ctx, Artifact)` (route by `artifact.Kind`, `Upsert`).
- `fx.go`: provide `NewProvidersStore`, `NewTrackStore`.

**Depends on:** T1.
**Verify:** `go test ./internal/infrastructure/repository/storage/...` — Provider round-trips through `settings.yaml`; `List` returns `[]Artifact` from a mixed track stream; `Update`/`Upsert` replace one `(kind,name)` and leave siblings intact.

---

## T3 — Provider filesystem + config

**Files:** `internal/config/configuration.go`, `internal/config/fx_test.go`,
`internal/infrastructure/repository/agent/fx.go`,
`internal/infrastructure/repository/agent/fx_test.go`.

Add `Configuration.UserHomeDirectory` (real `~` via `os.UserHomeDir()`). Provide a
second `afero.Fs` tagged `name:"provider"` = `BasePathFs(OsFs, UserHomeDirectory)`
in the agent fx — separate from storage's unnamed `$SAURON_HOME` fs. Confirm
`agent.NewFxOptions()` is in the app graph.

**Depends on:** T1.
**Verify:** `go test ./internal/config/... ./internal/infrastructure/repository/agent/...` — the named `provider` fs resolves from the graph and is distinct from storage's.

---

## T4 — Use case: `MigrateUseCase`

**Files:** `internal/usecase/usecase_migrate.go` (new),
`internal/usecase/mock_based_migrate.go` (new), `usecase_migrate_test.go` (new),
`internal/usecase/fx.go`.

`MigrateUseCase` struct, injected elsewhere as the generic
`UseCase[MigrateInput, MigrateResult]` (injects `TrackStore`, `afero.Fs`
`name:"provider"`, logger). `Execute(MigrateInput{From,To})`: for each
`TrackStore.List` artifact move `<providerDir(From)>/<path>` →
`<providerDir(To)>/<path>` (claude→`.claude`, zencoder→`.zencoder`) on the provider
fs, bump `spec.updatedAt`, `Update`; per-artifact failure recorded, migration
continues (FR-005). gocognit ≤15. Provided via `fx.As(new(UseCase[MigrateInput, MigrateResult]))`.

**Depends on:** T2, T3.
**Verify:** `go test ./internal/usecase/...` — move relocates `.claude/skills/sauron-x` → `.zencoder/...` and updates track; a missing source is recorded as failed while siblings still migrate; empty track → zero, no error.

---

## T5 — Use case + command: `set provider`

**Files:** `internal/usecase/usecase_set_provider.go` (new),
`usecase_set_provider_test.go` (new), `internal/cmd/cmd_set_provider.go` (new),
`internal/cmd/view_set_provider.go` (new), `cmd_set_provider_test.go`,
`view_set_provider_test.go`, `internal/cmd/cmd_set.go`, `internal/usecase/fx.go`.

`SetProviderUseCase` composes `UseCase[MigrateInput, MigrateResult]`: validate name (FR-004), `Get` current,
idempotent no-op (FR-003), else run migration on a real switch then persist the
`Provider` even on partial failure (FR-001/FR-002/FR-005). `SetProvider()` builder
(`Use: "provider <claude|zencoder>"`, `Args: usageArgs(cobra.ExactArgs(1))`) +
cobra-free `setProvider()` via `runUseCase`. `renderSetProvider()`: `skills:`/`agents:`
groups + summary on a populated migration, bare `provider set to "<name>"` /
`provider already set to "<name>"` otherwise.

**Depends on:** T4.
**Verify:** `go build ./... && go test ./internal/usecase/... ./internal/cmd/...` — first set persists; re-set of the active provider is `Unchanged` (no `Set`, no migrate); switch composes migrate then `Set`; unknown name → `TypeUsage`/exit-2.

---

## T6 — End-to-end: `set_provider.feature`

**Files:** `test/e2e/testdata/set_provider.feature`,
`test/e2e/internal/gherkin/set_provider_controller.go` (+ registration in `init.go`).

Scenarios driving the built binary (`SAURON_BIN`), asserting via `pkg/`: first set,
idempotent re-set (no change, exit 0), switch (zero migrated — nothing installed),
unknown provider name (exit 2), missing argument (exit 2).

**Depends on:** T5 (passes only once T5 lands). May be drafted in a worktree
(`feat/set-provider-e2e`, merged back uncommitted).
**Verify:** `task gate-integration`.

---

## T7 — Gates

**Files:** none (verification only); address any lint/coverage findings above.

**Depends on:** T6.
**Verify:** `task gate-lint && task gate-coverage && task gate-security` (or `task all`).

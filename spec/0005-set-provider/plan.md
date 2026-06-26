# Set Provider — implementation plan

How [`set provider`](spec.md) is built. The executable task breakdown lives in
[TASKS.md](TASKS.md). Code follows the
[architecture contract](../contracts/architecture.md) and the
`sauron-implementing-architecture` conventions; the end-to-end suite follows
`sauron-implementing-integration-tests`.

## Goal & scope

Add the `set provider` command: record the single global `Provider` document in
`settings.yaml` (FR-001), migrate every installed artifact to the new provider's
directories on a change (FR-002), no-op when the chosen provider is already active
(FR-003), reject unknown provider names and missing arguments as usage errors
(FR-004, FR-006), and on a per-artifact migration failure report it and continue,
leaving the provider set and the track file consistent with what migrated
(FR-005). The command mirrors the existing `set registry` slice
(use case → `fx.Populate` handler → `view_*` renderer → typed store), and composes
a dedicated migration use case.

**In scope:** all six requirements. Migration performs **real file moves** between
provider directories over an injected, provider-scoped filesystem, and is fully
unit-tested with an in-memory filesystem.

**Out of scope (YAGNI):** producing the installed artifacts that migration moves —
that is [install artifacts](../0007-install-artifacts/spec.md), not built yet. Until
install exists, `track.yaml` holds no artifacts, so an end-to-end provider switch
migrates zero; the migration machinery is exercised by unit tests that seed the
filesystem and track directly.

## Pre-requirements

- The `Provider` type ([`pkg/sauron/types/provider.go`](../../pkg/sauron/types/provider.go)),
  its kind constant (`types.KindProvider`), and its
  [schema](../contracts/schemas/Provider.schema.json) already exist.
- The kind-agnostic `Store` already provides `First`/`Replace`/`FindAll`; it gains
  the kind→file routes for `Provider`/`Skill`/`Agent` and a single-document
  `Upsert`.

## Design

Slice parallels `set registry`, plus a migration use case and a provider-scoped
filesystem. New and touched files:

| File | Change |
|---|---|
| `pkg/sauron/types/artifact.go` | **new** — single `Artifact` type (replaces the separate `Skill`/`Agent` Go structs); `ArtifactSpec` moved here. Skill vs agent is read from the embedded `TypeMeta.Kind` (`Skill`/`Agent`) — no extra discriminator field |
| `internal/infrastructure/repository/storage/store.go` | add `KindProvider → settings.yaml`, `KindSkill`/`KindAgent → track.yaml` routes; add `Upsert` (replace-matching-`(kind,name)`-or-append, atomic, lock-guarded) |
| `internal/infrastructure/repository/storage/providers_store.go` | **new** — `ProvidersStore` (singleton `Get`/`Set`), mirroring `registries_store.go` |
| `internal/infrastructure/repository/storage/track_store.go` | **new** — `TrackStore.List(ctx) ([]types.Artifact, error)` (each `Artifact` carries its `TypeMeta.Kind`) and `Update(ctx, Artifact)` (routes by `artifact.Kind`, upserts) |
| `internal/infrastructure/repository/storage/fx.go` | provide `NewProvidersStore`, `NewTrackStore` |
| `internal/config/configuration.go` | add `UserHomeDirectory` (the real `~`, via `os.UserHomeDir()`), distinct from the `$SAURON_HOME` `HomeDirectory` |
| `internal/infrastructure/repository/agent/fx.go` | provide a second `afero.Fs` tagged `name:"provider"` = `BasePathFs(OsFs, UserHomeDirectory)` — the provider-directory filesystem, separate from storage's unnamed `$SAURON_HOME` fs |
| `internal/usecase/usecase_migrate.go` | **new** — `MigrateUseCase` struct (no bespoke interface — it is injected as the generic `UseCase[MigrateInput, MigrateResult]`); moves each artifact `<.claude\|.zencoder>/<path>` on the provider fs, bumps `spec.updatedAt`, `Update`s track; per-artifact failure recorded, migration continues (FR-005) |
| `internal/usecase/usecase_set_provider.go` | **new** — `SetProviderUseCase` composes `UseCase[MigrateInput, MigrateResult]`; validates name, idempotent no-op, runs migration on a real switch, then persists the `Provider` (even on partial failure) |
| `internal/usecase/fx.go` | provide `NewMigrateUseCase` as `fx.As(new(UseCase[MigrateInput, MigrateResult]))` and `NewSetProviderUseCase` |
| `internal/cmd/cmd_set_provider.go` | **new** — `SetProvider()` builder + cobra-free `setProvider()` handler |
| `internal/cmd/view_set_provider.go` | **new** — `renderSetProvider()`: `skills:`/`agents:` groups + summary on a populated migration; bare confirmation otherwise |
| `internal/cmd/cmd_set.go` | `cmd.AddCommand(SetProvider())` |
| `test/e2e/testdata/set_provider.feature` + `.../gherkin/set_provider_controller.go` | **new** — first set, idempotent re-set, switch (zero migrated, nothing installed), unknown name (exit 2), missing arg (exit 2) |

### Provider directory layout

claude → `~/.claude`, zencoder → `~/.zencoder`; each artifact lives under its
kind's subdir (`skills/` or `agents/`) as `sauron-<name>`. `spec.path` is recorded
relative to the provider home, so migration moves `<oldHome>/<path>` →
`<newHome>/<path>` over the injected provider fs.

### `SetProviderUseCase.Execute` flow

1. Validate the name is `claude`/`zencoder`; else `NewUsageError` (FR-004).
   (Empty/extra args rejected earlier by `usageArgs(cobra.ExactArgs(1))`, FR-006.)
2. `ProvidersStore.Get` the current provider.
3. If the current name equals the requested → `{Unchanged:true}`, persist nothing
   (FR-003).
4. Otherwise, when a current provider exists and differs, run
   `migrate.Execute(From:current, To:requested)` (FR-002/FR-005), then
   `ProvidersStore.Set` the new `Provider` — persisting even if some migration
   steps failed. Build the result's `Skills`/`Agents` groups from the migration
   result, grouped by `artifact.Kind`.

## Key decisions

- **Single `types.Artifact`, discriminated by `kind`.** Skill and Agent are one Go
  type; skill-vs-agent is read from the embedded `TypeMeta.Kind`, so the track store
  returns `[]Artifact` and newer artifact kinds extend without changing the return
  shape. The on-disk format is **unchanged** — documents stay `kind: Skill|Agent`
  and the two schemas are untouched. No redundant discriminator field.
- **Migration is its own use case.** `MigrateUseCase` owns the file moves and the
  track rewrite; `SetProviderUseCase` composes it. This isolates the move logic and
  lets install (0007) reuse the same primitive.
- **Provider-scoped filesystem, separately tagged.** Migration moves files under
  the user's real home (`~/.claude`, `~/.zencoder`), which is a different root from
  storage's `$SAURON_HOME`. A second `afero.Fs` tagged `name:"provider"` is
  injected, never storage's fs — both are testable with an in-memory fs.
- **Mirror `set registry`** for the command/use-case/view/store shape.

## Notes

- **FR-002 path-invariance.** `spec.path` is relative to the provider home, so it is
  invariant across a provider switch — only the home changes. Migration moves the
  files and bumps `spec.updatedAt` but does not change the `path` string, which
  contradicts FR-002's literal "update each artifact's recorded path". Treated as a
  spec imprecision under the relative-path model; schemas untouched. Install (0007),
  which owns `spec.path`, should reconcile FR-002's wording.

## Checkpoints

| # | Milestone | Verify |
|---|---|---|
| C1 | types unified; storage routes + `Upsert` + stores compile and are unit-tested | `go test ./pkg/... ./internal/infrastructure/repository/storage/...` |
| C2 | migration + set-provider use cases unit-tested (move, partial failure, empty, idempotent, unknown name) | `go test ./internal/usecase/...` |
| C3 | command + view wired and unit-tested; whole tree builds | `go build ./... && go test ./internal/cmd/...` |
| C4 | full unit gate green and coverage floor held | `task test && task gate-coverage` |
| C5 | `set_provider.feature` scenarios pass end-to-end | `task gate-integration` |
| C6 | style + security gates green | `task gate-lint && task gate-security` |

## Execution flow

Sequential chain on the shared packages: **T1 → T2 → T3 → T4 → T5 → T6 → T7**
(see [TASKS.md](TASKS.md)). T6 (the `.feature` + controller) touches only
`test/e2e/**` and may be **drafted in a parallel git worktree**
(`branch feat/set-provider-e2e`, merged back into the working tree uncommitted),
but it only *passes* once T5 lands.

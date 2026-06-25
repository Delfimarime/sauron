# Tasks — Unset Registry

The executable breakdown of [plan.md](plan.md). Each task is **independently
verifiable**: it owns a set of files and states the single command whose success
confirms it. A task may start once its dependencies have verified.

> Authoring rule (see [AUTHORING.md](../AUTHORING.md)): every task carries a
> verification — a task without a pass/fail check is not a task.

## Dependency order

- T1 → T2 → T3 → T4

## Tasks

### T1 — Unset use case
- **Delivers:** `UnsetRegistryUseCase` — the `Get` → (nil → `Nothing`) →
  (`DryRun` → `Preview`) → `Remove` → `Removed` pipeline; `UnsetRegistryInput`
  and the presentation-agnostic `UnsetRegistryResult`; the `UnsetOutcome` enum
  (`UnsetNothing`/`UnsetPreview`/`UnsetRemoved`); the `io` classification on a
  read or remove failure; the ECS-logged outcome; the fx wiring. The registry is
  removed through the existing singleton `RegistriesStore.Remove`, which preserves
  every installed artifact and every other kind in `settings.yaml`. Realizes
  FR-001, FR-002, FR-004, FR-005.
- **Files:** `internal/usecase/{usecase_unset_registry.go, fx.go}` (+ test).
- **Verify:** `go test ./internal/usecase/...`
- **Depends on:** —

### T2 — cmd surface + view rendering
- **Delivers:** the `Unset()` group, the `UnsetRegistry()` builder and the
  `unsetRegistry()` handler (`cobra.NoArgs` via the shared usage-args wrapper, the
  `--dry-run` flag, `runUseCase`, render); the cobra-free `renderUnsetRegistry`
  and the `unsetMessages` outcome → report-line map (`view_unset_registry.go`);
  the `root.AddCommand(Unset())` wiring. Realizes FR-004, FR-005, FR-006.
- **Files:** `internal/cmd/{cmd_unset.go, cmd_unset_registry.go,
  view_unset_registry.go, cmd_root.go}` (+ tests).
- **Verify:** `go test ./internal/cmd/...`
- **Depends on:** T1

### T3 — e2e suite
- **Delivers:** the `unset registry` feature file and controller — a configured
  registry is removed from `settings.yaml` and the report confirms artifacts are
  preserved; `--dry-run` leaves the registry in place; no registry configured
  exits 0 and reports nothing was unset; a seeded installed artifact remains
  after the unset. Covers FR-001, FR-002, FR-004, FR-005.
- **Files:** `test/e2e/**`.
- **Verify:** `task build && task gate-integration`
- **Depends on:** T2

### T4 — Full verification gate
- **Delivers:** the complete gate over the feature.
- **Verify:** `task all`
- **Depends on:** T3

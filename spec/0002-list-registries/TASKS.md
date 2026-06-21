# Tasks — List Registries

The executable breakdown of [plan.md](plan.md). Each task is **independently
verifiable**: it owns a set of files and states the single command or criterion
whose success confirms it. The suite is authored **TDD-first** — the e2e tests
(T2) are written before the product (T3–T6) and stay red until the command lands
(T6), per the [Constitution](../../CONSTITUTION.md) Chapter III, Article 7.

> Authoring rule (see [AUTHORING.md](../AUTHORING.md)): every task carries a
> verification — a task without a pass/fail check is not a task.

## Dependency order

- T1 → T2  (the e2e scenarios encode the corrected decisions)
- T1 → T3, T4  (T3 and T4 are independent; either may be worktree-isolated)
- T3, T4 → T5 → T6  (T6 turns the e2e suite from red to green)
- T6 → T7

## Tasks

### T1 — Specification, contract & constitution corrections
- **Delivers:** the [spec](spec.md) FR-002 default columns set to
  `name, transport, uri`; the [state](data/state.md) readable set and
  field→requirement table extended with `spec.uri` and `spec.ref`; the
  [architecture contract](../contracts/architecture.md) registering
  `internal/presentation`; the [harness reference](../../test/e2e/HARNESS.md)
  amended (the seeding exception admitted; the uniform-exercise section
  de-scoped from naming a single command).
- **Files:** `spec/0002-list-registries/spec.md`,
  `spec/0002-list-registries/data/state.md`, `spec/contracts/architecture.md`,
  `test/e2e/HARNESS.md`.
- **Verify:** FR-002 names `uri`; `data/state.md` realizes `spec.uri`;
  `architecture.md` lists `internal/presentation`; the harness reference's
  uniform-exercise section names no single command (inspection).
- **Depends on:** —

### T2 — e2e suite (authored TDD-first; red until T6)
- **Delivers:** the `list registries` feature file and a controller, authored so
  every step resolves and the only failure is the not-yet-built command.
- **New steps (`list_controller.go`):**
  | Step | Role |
  |---|---|
  | `Given the following registries are configured:` (table) | seed a schema-valid `types.Registry` stream into `$SAURON_HOME` via the runtime's `CopyTo` |
  | `Given the stored registries file is corrupt` | seed malformed bytes (drives FR-006) |
  | `Then the registries are listed in order: (.+)` | read the name column down the rows and assert the sequence |
  | `Then the output is empty` | assert stdout carries no bytes |
- **Reused steps:** `the user runs (.+)`, `the command succeeds`,
  `the command exits with status (\d+)`, `the output contains (.+)`.
- **Scenarios (`list_registries.feature`, host runtime via `@no-sandbox`):**
  | # | Requirement | Scenario |
  |---|---|---|
  | 1 | FR-001 / FR-002 | seed two registries → the default `name, transport, uri` table lists both |
  | 2 | FR-002 | `--fields name,uri` → only the name and uri columns appear |
  | 3 | FR-003 | `--search` with a mixed-case term → only the matching registry |
  | 4 | FR-004 | `--sort transport --order desc` → the asserted row order |
  | 5 | FR-005 | no registry configured → empty output, exit 0 |
  | 6 | FR-006 | a corrupt `registries.yaml` → exit 1 |
  | 7 | FR-007 | an invalid `--sort` value → exit 2 |
  | 8 | write-then-read | black-box: `add registry` (filesystem) → `list registries` shows it |
- **Files:** `test/e2e/testdata/list_registries.feature`,
  `test/e2e/internal/gherkin/list_controller.go` (+ its registration in
  `init.go`).
- **Verify:** `task build && task gate-integration` — green after T6; before T6,
  the suite resolves every step and fails only on the missing command (no
  undefined, pending, or ambiguous steps).
- **Depends on:** T1 (a green result also requires T6)

### T3 — Shared table renderer (`internal/presentation`)
- **Delivers:** `Table{Headers, Rows}` and `Render(w)` producing the
  [CLI contract](../contracts/cli.md) rendering — aligned columns, uppercase
  headers, `—` for an empty cell, and no output for zero rows; `doc.go`.
- **Files:** `internal/presentation/{table.go, doc.go, table_test.go}`.
- **Verify:** `go test ./internal/presentation/...`
- **Depends on:** T1

### T4 — Store listing read path (`FindAll` + `List`)
- **Delivers:** `Store.FindAll(ctx, kind)` (validate-on-read, all-or-nothing);
  `RegistriesStore.List(ctx)`; the regenerated
  `mock_based_registries_store.go`.
- **Files:** `internal/infrastructure/repository/storage/{store.go,
  registries_store.go, mock_based_registries_store.go}` (+ tests).
- **Verify:** `go test ./internal/infrastructure/repository/storage/...`
- **Depends on:** T1

### T5 — List use case
- **Delivers:** `ListRegistriesUseCase` and `ListRegistriesRequest` — the read →
  filter → sort → project → render pipeline over `presentation.Table`; the
  `usage`/`io` classification; the ECS-logged outcome; the fx wiring.
- **Files:** `internal/usecase/{usecase_list_registries.go, fx.go}` (+ test).
- **Verify:** `go test ./internal/usecase/...`
- **Depends on:** T3, T4

### T6 — cmd surface
- **Delivers:** the `List()` group, the `ListRegistries()` builder and the
  `listRegistries()` handler (`fx.Invoke`), the `--search`/`--fields`/`--sort`/
  `--order` flag groups, and the `root.go` wiring. Turns the T2 suite green.
- **Files:** `internal/cmd/{list.go, list_registries.go, helper_flags.go,
  root.go}` (+ tests).
- **Verify:** `go test ./internal/cmd/...`
- **Depends on:** T5

### T7 — Full verification gate
- **Delivers:** the complete gate over the feature.
- **Verify:** `task all`
- **Depends on:** T6

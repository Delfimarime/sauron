# Tasks — Delete Registry

The executable breakdown of [plan.md](plan.md). Each task is **independently
verifiable**: it owns a set of files and states the single command or criterion
whose success confirms it. The suite is authored **TDD-first** — the e2e tests
(T2) are written before the product (T3–T6) and stay red until the command lands
(T6), per the [integration constitution](../../test/e2e/CONSTITUTION.md)
Chapter I, Article 3.

> Authoring rule (see [AUTHORING.md](../AUTHORING.md)): every task carries a
> verification — a task without a pass/fail check is not a task.

> Deferral (see [plan.md](plan.md) §1, §8): the artifact cascade
> (FR-002/FR-003/FR-007) is **owned by
> [0007](../0007-uninstall-artifacts/spec.md)**. This feature ships the cascade as
> a **no-op shared Action** (T4). The e2e suite runs in full; only the single
> scenario that asserts installed artifacts were actually removed is **commented
> out** in the feature file (T2), pointing at 0007 — it cannot be arranged until
> install exists. No `track.yaml`, provider port, or track store is built here, and
> no `@deferred` tag or gate filter is introduced.

## Dependency order

- T1 → T2  (the e2e scenarios encode the spec, incl. the one commented-out scenario)
- T1 → T3, T4  (T3 and T4 are independent; either may be worktree-isolated)
- T3, T4 → T5 → T6  (T6 turns the e2e scenarios from red to green)
- T6 → T7

## Tasks

### T1 — Spec deferral note & contract reconciliation
- **Delivers:** the [spec](spec.md) `## Notes` recording that FR-002/FR-003/FR-007
  are realized only once the shared Action's body lands in
  [0007](../0007-uninstall-artifacts/spec.md). Reconcile genuine spec/contract
  drift only — do not invent corrections. No constitution amendment and no
  test-harness change belong to this feature.
- **Files:** `spec/0004-delete-registry/spec.md` (+ `data/state.md`,
  `contracts/delete-registry.md` only if a genuine drift is found).
- **Verify:** the spec `## Notes` names 0007 for the cascade (inspection).
- **Depends on:** —

### T2 — e2e suite (authored TDD-first; red until T6)
- **Delivers:** the `delete registry` feature file and a controller, authored so
  every step resolves and the only failure is the not-yet-built command.
- **New steps (`delete_controller.go`), only where an existing step does not fit:**
  | Step | Role |
  |---|---|
  | `Then the output reports nothing was deleted` | assert the FR-005 idempotent-delete message |
  | `Then the removal summary reads (.+)` | assert the `registry "X" removed; N artifacts removed` line |
- **Reused steps:** `Given the following registries are configured:` (table seed
  into `$SAURON_HOME` via the runtime's `CopyTo`), `the user runs (.+)`,
  `the command succeeds`, `the command exits with status (\d+)`,
  `the output contains (.+)`, and the `stateController` file assertions
  (`a registry named (.+) exists`, `there are (\d+) registries`,
  `the stored state does not contain (.+)`).
- **Scenarios — live (`delete_registry.feature`, `@no-sandbox`; all go green at T6):**
  | # | Requirement | Scenario |
  |---|---|---|
  | 1 | FR-001 | seed two registries → `delete registry acme` → `acme` is gone from `registries.yaml`, the other remains, summary reads `registry "acme" removed; 0 artifacts removed` |
  | 2 | FR-005 | `delete registry ghost` (absent) → exit 0, reports nothing was deleted, the state is unchanged |
  | 3 | FR-004 | `delete registry acme --dry-run` → `registries.yaml` still contains `acme`, the plan is printed, exit 0 |
  | 4 | FR-006 | `delete registry` with no `<name>` (or an unknown flag) → exit 2 |
  | 5 | write-then-read | black-box: `add registry` (filesystem) → `delete registry` → the added registry is gone |
- **Commented-out scenario (the one effect this feature defers):** a single
  scenario asserting that a registry's **installed artifacts were actually
  removed** by the cascade (FR-002/FR-003/FR-007) is written but **commented out**
  in `delete_registry.feature`, prefixed
  `# Deferred to 0007: requires install + the Action body; uncomment when artifact
  removal lands`. It cannot be arranged until install exists. No tag, no gate
  filter — it is simply not active Gherkin.
- **Files:** `test/e2e/testdata/delete_registry.feature`,
  `test/e2e/internal/gherkin/delete_controller.go` (+ its registration in
  `init.go`).
- **Verify:** `task build && task gate-integration` — the live scenarios are green
  after T6; before T6 they resolve every step and fail only on the missing command
  (no undefined, pending, or ambiguous steps). The commented-out scenario is inert.
- **Depends on:** T1 (a green result also requires T6)

### T3 — Registry-removal write path (`Store.Remove` + `RegistriesStore.Remove`)
- **Delivers:** `Store.Remove(ctx, kind, name)` — read the document stream, drop
  the matched document, rewrite the file through `writeAtomic` under the write
  lock; removing an absent document is a no-op success (FR-005). `RegistriesStore.Remove(ctx, name)`;
  the regenerated `mock_based_registries_store.go`.
- **Files:** `internal/infrastructure/repository/storage/{store.go,
  registries_store.go, mock_based_registries_store.go}` (+ tests).
- **Verify:** `go test ./internal/infrastructure/repository/storage/...`
- **Worktree isolation:** independent of T4; if executed in parallel, run on
  branch `feat/0004-registry-remove` in its own git worktree and merge the branch
  back into the working tree without committing or pushing.
- **Depends on:** T1

### T4 — Shared no-op cascade Action (`UninstallByRegistryAction`)
- **Delivers:** the `UninstallByRegistryAction` ([`Action[R,P]`](../contracts/architecture.md))
  and the `RemovalPlan` value type (`Skills`/`Agents`/`Personas` + `Total()`); the
  **no-op** body returning `&RemovalPlan{}, nil`, with a `// 0007 owns the real
  body` note; the test asserting the empty-plan/`nil` contract.
- **Files:** `internal/usecase/{action_uninstall_by_registry.go,
  action_uninstall_by_registry_test.go}`.
- **Verify:** `go test ./internal/usecase/...`
- **Worktree isolation:** independent of T3; if executed in parallel, run on
  branch `feat/0004-cascade-action` in its own git worktree and merge the branch
  back into the working tree without committing or pushing.
- **Depends on:** T1

### T5 — Delete use case
- **Delivers:** `DeleteRegistryUseCase` and `DeleteRegistryRequest` — the
  `FindByName` → (absent → exit 0) → cascade (no-op Action) → `--dry-run` →
  `Remove` → report pipeline; the grouped-report helper (non-empty groups +
  summary count); the `usage`/`io`/not-found-as-success classification; the
  ECS-logged outcome; the fx wiring through `NewFxOptions`.
- **Files:** `internal/usecase/{usecase_delete_registry.go, fx.go}` (+ test).
- **Verify:** `go test ./internal/usecase/...`
- **Depends on:** T3, T4

### T6 — cmd surface
- **Delivers:** the `Delete()` group (mirroring `List()`), the `DeleteRegistry()`
  builder and the `deleteRegistry()` handler (`fx.Invoke`), the `<name>`
  positional arg (via the shared usage-args wrapper), the `--dry-run` flag, and
  the `root.AddCommand(Delete())` wiring. Turns the implemented T2 scenarios green.
- **Files:** `internal/cmd/{delete.go, delete_registry.go, root.go}` (+ tests).
- **Verify:** `go test ./internal/cmd/...`
- **Depends on:** T5

### T7 — Full verification gate
- **Delivers:** the complete gate over the feature.
- **Verify:** `task all`
- **Depends on:** T6

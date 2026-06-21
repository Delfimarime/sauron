# Tasks — Describe Registry

The executable breakdown of [plan.md](plan.md). Each task is **independently
verifiable**: it owns a set of files and states the single command or criterion
whose success confirms it. The suite is authored **TDD-first** — the e2e tests
(T2) are written before the product (T3–T5) and stay red until the command lands
(T5), per the [integration constitution](../../test/e2e/CONSTITUTION.md)
Chapter I, Article 3.

> Authoring rule (see [AUTHORING.md](../AUTHORING.md)): every task carries a
> verification — a task without a pass/fail check is not a task.

## Dependency order

- T1 → T2  (the e2e scenarios encode the reconciled decisions)
- T1 → T3  (the descriptor renderer is independent; it may be worktree-isolated)
- T3 → T4 → T5  (T5 turns the e2e suite from red to green)
- T5 → T6

## Tasks

### T1 — Specification & contract reconciliation
- **Delivers:** confirmation that the [spec](spec.md), the
  [state](data/state.md) field→requirement table, and the
  [`describe registry` command contract](contracts/describe-registry.md) agree on
  the `--fields` valid set `{name, transport, uri, ref, auth, tls, sshKey,
  timeout}`; the **not-found error class** is settled (resolved,
  [plan.md](plan.md) §8) — a `TypeNotFound` mapping to exit 1, recorded in the
  spec `## Notes`. No re-registration of `internal/presentation` is needed: the
  package already ships from [0002](../0002-list-registries/plan.md).
- **Files:** `spec/0003-describe-registry/spec.md` (record the `TypeNotFound`/exit-1
  resolution; otherwise reconcile only genuine drift).
- **Verify:** the `--fields` set in `data/state.md` matches the command contract;
  the spec `## Notes` records not-found → `TypeNotFound` → exit 1 (inspection). Do
  not invent corrections beyond that.
- **Depends on:** —

### T2 — e2e suite (authored TDD-first; red until T5)
- **Delivers:** the `describe registry` feature file and a controller, authored
  so every step resolves and the only failure is the not-yet-built command.
- **New steps (`describe_controller.go`), only where an existing step does not fit:**
  | Step | Role |
  |---|---|
  | `Then the descriptor shows (.+) as (.+)` | read a `label: value` line from the descriptor and assert the pair |
  | `Then the output does not contain (.+)` | assert a resolved-secret value never appears (drives FR-002) |
- **Reused steps:** `Given the following registries are configured:` (table seed
  into `$SAURON_HOME` via the runtime's `CopyTo`), `the user runs (.+)`,
  `the command succeeds`, `the command exits with status (\d+)`,
  `the output contains (.+)`.
- **Scenarios (`describe_registry.feature`, host runtime via `@no-sandbox`):**
  | # | Requirement | Scenario |
  |---|---|---|
  | 1 | FR-001 | seed a git registry → `describe registry <name>` shows every field |
  | 2 | FR-003 | `--fields name,transport,uri` → only those fields, in that order, name first |
  | 3 | FR-002 | a registry with `auth` → the `auth` block shows the `${env:…}` references and never a resolved secret |
  | 4 | FR-004 | an unknown name → exit 1, reports the registry does not exist |
  | 5 | FR-005 | an invalid `--fields` value → exit 2 |
- **Files:** `test/e2e/testdata/describe_registry.feature`,
  `test/e2e/internal/gherkin/describe_controller.go` (+ its registration in
  `init.go`).
- **Verify:** `task build && task gate-integration` — green after T5; before T5,
  the suite resolves every step and fails only on the missing command (no
  undefined, pending, or ambiguous steps).
- **Depends on:** T1 (a green result also requires T5)

### T3 — Shared descriptor renderer (`internal/presentation`)
- **Delivers:** `Descriptor{Fields}` / `Field{Label, Value, Children}` and
  `Render(w)` producing the [CLI contract](../contracts/cli.md) detail rendering
  — a `kubectl describe`-style vertical view: left-aligned labels with their
  values and an indented nested block for a section field (e.g. `auth`);
  explicitly distinct from the column-aligned `Table`. `descriptor.go` beside the
  existing `table.go`.
- **Files:** `internal/presentation/{descriptor.go, descriptor_test.go}`.
- **Verify:** `go test ./internal/presentation/...`
- **Worktree isolation:** independent of T4/T5; if executed in parallel with T2,
  run on branch `feat/0003-descriptor-renderer` in its own git worktree and merge
  the branch back into the working tree without committing or pushing.
- **Depends on:** T1

### T4 — Describe use case
- **Delivers:** the `TypeNotFound` constant added to the `usecase.Error` model
  with `cmd/main.go` mapping it to exit 1 (per the resolved
  [plan.md](plan.md) §8 decision); `DescribeRegistryUseCase` and
  `DescribeRegistryRequest` — the `FindByName` → not-found (`TypeNotFound`) →
  field projection → render pipeline over `presentation.Descriptor`; the
  `usage`/`io`/not-found classification; the ECS-logged outcome; the fx wiring.
  The `auth` block is projected as a nested `Field`, its values left as the
  stored env references (FR-002).
- **Files:** `internal/usecase/{usecase_describe_registry.go, api.go, fx.go}`,
  `cmd/main.go` (+ tests).
- **Verify:** `go test ./internal/usecase/... ./cmd/...`
- **Depends on:** T3 (and the T1-recorded not-found resolution)

### T5 — cmd surface
- **Delivers:** the `Describe()` group (mirroring `List()`), the
  `DescribeRegistry()` builder and the `describeRegistry()` handler (`fx.Invoke`),
  the `<name>` positional arg (`cobra.ExactArgs(1)` via the shared usage-args
  wrapper), the `--fields` flag, and the `root.AddCommand(Describe())` wiring.
  Turns the T2 suite green.
- **Files:** `internal/cmd/{describe.go, describe_registry.go, root.go}`
  (+ tests).
- **Verify:** `go test ./internal/cmd/...`
- **Depends on:** T4

### T6 — Full verification gate
- **Delivers:** the complete gate over the feature.
- **Verify:** `task all`
- **Depends on:** T5

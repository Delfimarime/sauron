# Tasks — Describe Registry

**Status:** Built

The executable breakdown of [plan.md](plan.md). Each task is **independently
verifiable**: it owns a set of files and states the single command whose success
confirms it. A task may start once its dependencies have verified.

> Authoring rule (see [AUTHORING.md](../AUTHORING.md)): every task carries a
> verification — a task without a pass/fail check is not a task.

## Dependency order

- T1 → T3  (the projection renders onto the shared descriptor)
- T2 → T3  (the handler renders the use-case result)
- T3 → T4 → T5

## Tasks

### T1 — Shared descriptor renderer
- **Delivers:** `descriptor{Fields}` / `descriptorField{Label, Value, Children}`
  and `render(w)` producing the [CLI contract](../contracts/cli.md) detail
  rendering — a `kubectl describe`-style vertical view: left-aligned labels with
  their values aligned into one column and an indented nested block for a section
  field (e.g. `auth`, `tls`); explicitly distinct from the column-aligned table.
  A pure, cobra-free value type in `package cmd`, standard library only.
- **Files:** `internal/cmd/{view_descriptor.go, view_descriptor_test.go}`.
- **Verify:** `go test ./internal/cmd/...`
- **Depends on:** —

### T2 — Describe use case
- **Delivers:** `DescribeRegistryUseCase` — the `Get` → not-found
  (`NewNotFoundError`, `TypeNotFound`) pipeline; the **empty**
  `DescribeRegistryInput`; `DescribeRegistryUseCaseParams` (fx-injected
  `RegistriesStore` + `*zap.Logger`); the `io`/not-found classification; the
  ECS-logged outcome; the fx registration of `NewDescribeRegistryUseCase` in
  `NewFxOptions`. The use case is presentation-ignorant: it takes no field
  selection and returns the bare `*types.Registry` (no `DescribeRegistryResult`,
  no field-selection helper in `internal/usecase`). Realizes FR-001, FR-004,
  FR-006.
- **Files:** `internal/usecase/{usecase_describe_registry.go, fx.go}` (+ test).
- **Verify:** `go test ./internal/usecase/...`
- **Depends on:** —

### T3 — cmd surface + `--fields` validation + registry projection
- **Delivers:** the `Describe()` group, the `DescribeRegistry()` builder and the
  `describeRegistry()` handler (`cobra.NoArgs` via the shared `usageArgs` wrapper,
  the `--fields` flag, `runUseCase`, render); the **`--fields` vocabulary and
  boundary validation** in `view_describe_registry.go` — the ordered field set
  `{uri, transport, ref, auth, tls, sshKey, timeout, creationTimestamp,
  lastUpdatedTimestamp}` and `selectDescribeFields` (forces `uri` present and
  first, dedup; an unknown value → `errInvalidFlag`, usage, exit 2), run **before**
  the use case; the `renderDescribeRegistry` / `projectRegistry` projection mapping
  the validated fields onto descriptor fields, skipping empty values, nesting
  `auth`/`tls` as their stored env references (FR-002); the
  `root.AddCommand(Describe())` wiring. Realizes FR-002, FR-003, FR-005, FR-006.
- **Files:** `internal/cmd/{cmd_describe.go, cmd_describe_registry.go,
  view_describe_registry.go, cmd_root.go}` (+ tests).
- **Verify:** `go test ./internal/cmd/...`
- **Depends on:** T1, T2

### T4 — e2e suite
- **Delivers:** the `describe registry` feature file and controller — a seeded
  registry shows every populated field including the audit timestamps; `--fields`
  shows only the requested fields, in order, `uri` first; an `auth` block shows
  the `${env:…}` references and never a resolved secret; no registry configured
  exits 1; an invalid `--fields` value exits 2. Covers FR-001, FR-002, FR-003,
  FR-004, FR-005, FR-006.
- **Files:** `test/e2e/**`.
- **Verify:** `task build && task gate-integration`
- **Depends on:** T3

### T5 — Full verification gate
- **Delivers:** the complete gate over the feature.
- **Verify:** `task all`
- **Depends on:** T4

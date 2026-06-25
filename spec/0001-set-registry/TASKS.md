# Tasks — Set Registry

The executable breakdown of [plan.md](plan.md). Each task is **independently
verifiable**: it owns a set of files and states the single command whose success
confirms it. A task may start once its dependencies have verified.

> Authoring rule (see [AUTHORING.md](../AUTHORING.md)): every task carries a
> verification — a task without a pass/fail check is not a task.

## Dependency order

- T1 → T2  (the open action composes the transport ports; storage is independent of it)
- T1, T2 → T3 → T4 → T5 → T6

## Tasks

### T1 — Storage engine + singleton `RegistriesStore`
- **Delivers:** the kind-scoped `Store` engine (`First` reads the first document
  of a kind validate-on-read, `Replace` drops that kind's documents and appends
  the new one while preserving other kinds, lock + atomic temp+rename) and the
  singleton `RegistriesStore` (`Get` → nil when none set, `Set` stamps the
  envelope and replaces, `Remove` purges) over `settings.yaml`; the embedded
  `Registry` JSON schema; the regenerated `mock_based_registries_store.go`; the
  fx wiring. Realizes FR-001, FR-007.
- **Files:** `internal/infrastructure/repository/storage/{store.go,
  registries_store.go, schema.go, lock.go, filesystem.go, fx.go,
  mock_based_registries_store.go}` (+ tests).
- **Verify:** `go test ./internal/infrastructure/repository/storage/...`
- **Depends on:** —

### T2 — Shared `OpenRegistry` Action
- **Delivers:** `OpenRegistryAction`/`OpenRegistry` — selects the transport
  adapter, resolves `${env:VAR}` credential references, builds the
  `extension.Option` set from the registry spec, and opens the source, returning
  a `source.FileSystem`; an unknown transport is usage, an unset reference or a
  failed open is unreachable; the `mock_based_open_registry.go`; the fx
  `fx.As(new(OpenRegistry))` provider. Realizes FR-003 (reference resolution),
  FR-011, FR-012, FR-013 (option build).
- **Files:** `internal/usecase/{action_open_registry.go,
  mock_based_open_registry.go, fx.go}` (+ test).
- **Verify:** `go test ./internal/usecase/...`
- **Depends on:** T1

### T3 — Set use case
- **Delivers:** `SetRegistryUseCase` — the credential-format → `selectAdapter` →
  `Validate(opts)` → presence probe (`OpenRegistry` + `List(.skills/.agents, 1)`)
  → persist pipeline; `SetRegistryInput` and the presentation-agnostic
  `SetRegistryResult`; `classifyAdapterErr` (`api.ErrUsage` → usage, else
  unreachable); the audit-timestamp stamp at persist; the `usecase.Error{Type,
  Reason}` model (with `TypeUsage`/`TypeUnreachable`/`TypeIO`); the ECS-logged
  outcome; the fx wiring. Realizes FR-003, FR-004, FR-005, FR-006, FR-007,
  FR-010, FR-011, FR-012, FR-013, FR-014.
- **Files:** `internal/usecase/{usecase_set_registry.go, api.go, fx.go,
  helper.go}`, `internal/telemetry/fields.go` (+ tests).
- **Verify:** `go test ./internal/usecase/...`
- **Depends on:** T1, T2

### T4 — cmd surface + view rendering
- **Delivers:** the `Set()` group, the `SetRegistry()` builder and the
  `setRegistry()` handler (validates flags, maps flags+`<uri>` onto
  `SetRegistryInput`, runs `runUseCase`, renders the result); the
  `kindFlags`/`timeoutFlags` binders with the `http`/`30s` defaults (FR-002,
  FR-012) and `kind` validation; the cobra-free `renderSetRegistry`
  (`view_set_registry.go`); `runUseCase[U,P]`, `usageArgs`, and the `exitCode`
  mapping (`usage → 2`, else → 1); the `root.AddCommand(Set())` wiring. Realizes
  FR-002, FR-005, FR-009.
- **Files:** `internal/cmd/{cmd_set.go, cmd_set_registry.go, view_set_registry.go,
  helper.go, helper_flags.go, helper_fx.go, cmd_root.go}` (+ tests).
- **Verify:** `go test ./internal/cmd/...`
- **Depends on:** T3

### T5 — e2e suite
- **Delivers:** the `set registry` feature file and controller covering the three
  transports — filesystem from authored content, http behind basic auth with a
  `${env:VAR}` secret persisted as a reference, git over SSH pinned with `--ref`
  asserting `spec.transport`/`spec.ref` and the audit timestamps in
  `settings.yaml`; the empty-source and missing-`<uri>` failure scenarios. Covers
  FR-001, FR-003, FR-004, FR-005, FR-009, FR-010, FR-011, FR-013, FR-014.
- **Files:** `test/e2e/**`.
- **Verify:** `task build && task gate-integration`
- **Depends on:** T4

### T6 — Full verification gate
- **Delivers:** the complete gate over the feature.
- **Verify:** `task all`
- **Depends on:** T5

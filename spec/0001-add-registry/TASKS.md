# Tasks — Add Registry

The executable breakdown of [plan.md](plan.md). Each task is **independently
verifiable**: it owns a set of files and states the single command whose success
confirms it is done. A task may start once its dependencies have verified.

> Authoring rule (see [AUTHORING.md](../AUTHORING.md)): every task carries a
> verification — a task without a pass/fail check is not a task.

## Dependency order

- T1 → T2 → T3
- T3 → T5 → T4  (the rest adapter wraps the marketplace client)
- T1, T2 → T6  (storage; runs alongside T3–T4)
- T4 + T6 → T7 → T8 → T9 → T10

## Tasks

### T1 — Dependencies & build scaffolding
- **Delivers:** `go-git/v5`, `jsonschema-go`, `resty/v2` on `go.mod`; the
  `generate` Taskfile target (stages `spec/contracts/schemas/*.json` →
  `internal/.../storage/schemas/`, which `test`/`build` depend on); the
  `.gitignore` entry for the generated dir.
- **Files:** `go.mod`, `go.sum`, `Taskfile.yml`, `.gitignore`.
- **Verify:** `task generate && go build ./...`
- **Depends on:** —

### T2 — Domain type: `RegistrySpec.Ref`
- **Delivers:** `Ref string` on `pkg/sauron/types.RegistrySpec` (the git ref,
  persisted, `omitempty`).
- **Files:** `pkg/sauron/types/registry.go` (+ round-trip test).
- **Verify:** `go test ./pkg/sauron/types/...`
- **Depends on:** T1

### T3 — Ports: `source.FileSystem` + `extension.Registry`
- **Delivers:** `pkg/sauron/source` (`FileSystem`/`File`/`Stat`/`Options`/
  `ErrNotImplemented` + mock); `extension.Registry` evolved to `Validate` and
  `Open → source.FileSystem`, with `Options` (incl. `Ref`) + mock.
- **Files:** `pkg/sauron/source/*`, `pkg/sauron/extension/registry.go`
  (+ mocks/tests).
- **Verify:** `go build ./pkg/... && go test ./pkg/sauron/source/... ./pkg/sauron/extension/...`
- **Depends on:** T2

### T4 — Registry adapters + `api`
- **Delivers:** the single `registry` package — `os_filesystem.go`,
  `git_filesystem.go` (shallow clone @ ref), `rest_filesystem.go` (thin adapter
  over the marketplace client); `registry/api` (`Directory`, `ErrUsage`/
  `ErrRuntime`, the `Resolve`/`HasAuth`/`HasTLS` helpers); `fx.go` providing the
  three named `extension.Registry`.
- **Files:** `internal/.../registry/{os,git,rest}_filesystem.go`,
  `registry/api/*`, `registry/fx.go` (+ tests).
- **Verify:** `go test ./internal/infrastructure/repository/registry/...`
- **Depends on:** T3, T5

### T5 — Marketplace client
- **Delivers:** `pkg/sauron/marketplace` — `Client` / `ArtifactClient.List`, the
  typed results, `APIError` + predicates, plain resty (sends `q`/`limit`/`offset`/
  `sort`).
- **Files:** `pkg/sauron/marketplace/*`.
- **Verify:** `go test ./pkg/sauron/marketplace/...`
- **Depends on:** T1

### T6 — Storage engine + `RegistriesStore`
- **Delivers:** the `Store` engine (`FindOne` validate-on-read, `Append`
  lock+atomic, kind→file), `schema.go` (`go:embed`), `lock.go`,
  `registries_store.go` (+ mock), `fx.go`.
- **Files:** `internal/.../storage/*`.
- **Verify:** `go test ./internal/infrastructure/repository/storage/...`
- **Depends on:** T1, T2

### T7 — Use case
- **Delivers:** `AddRegistryUseCase` (the ordered pipeline) + `AddRegistryRequest`;
  `usecase.Error{Type,Reason}`; the ECS-logged outcome (`sauron.registry.*`); fx
  wiring.
- **Files:** `internal/usecase/{usecase_add_registry,api,fx}.go`,
  `internal/telemetry/fields.go` (+ tests).
- **Verify:** `go test ./internal/usecase/...`
- **Depends on:** T3, T4, T6

### T8 — cmd surface
- **Delivers:** the `Add()` group, the `AddRegistry()` builder, the
  `addRegistry()` handler (`fx.Invoke`), `usageArgs`, `kindFlags`, `exitCode`, the
  `cmd/main.go` single error site, the `root.go` wiring.
- **Files:** `internal/cmd/{add,add_registry,helper,helper_flags,root}.go`,
  `cmd/main.go` (+ tests).
- **Verify:** `go test ./internal/cmd/...`
- **Depends on:** T7

### T9 — e2e suite
- **Delivers:** the positional `addRegistryArgs`; the git-SSH and http-API
  testcontainers fixtures; the `--ref` scenario; the `@git` Linux gate.
- **Files:** `test/e2e/**`.
- **Verify:** `task build && task gate-integration`
- **Depends on:** T8

### T10 — Full verification gate
- **Delivers:** the complete gate over the feature.
- **Verify:** `task all`
- **Depends on:** T9

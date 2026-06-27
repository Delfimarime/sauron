# Describe Provider — tasks

The executable breakdown for [plan.md](plan.md). Each task owns its files, states
a single pass/fail verification, and lists its dependencies.

> **TDD-first, non-negotiable.** T1 (the `test/e2e` feature) is written **before
> any production code** and must run **red**. The code tasks (T2–T5) are each
> **test-first**: the unit test precedes its production code. T6 (`task all`) is
> the closing gate — the whole contract must end green.

## Order

**T1 (e2e, red) → T2 → T3 → T4 → T5 → T6 (`task all`)** — sequential. T1 touches
only `test/e2e/**` and may be drafted in worktree `feat/describe-provider-e2e`,
merged back uncommitted; it passes only once T5 lands.

## Tasks

| # | Task | Owns | Depends | Verify (pass/fail) |
|---|---|---|---|---|
| **T1** | **TDD red.** Author `describe_provider.feature` (FR-001 first describe · FR-002 `--fields` · FR-003 none-set exit 0 · FR-005 invalid field exit 2 · a synced-provider doc-string asserting the full descriptor `reads:` block) + `describe_provider_controller.go` (one new step: `the output reports no provider is set`; reuse descriptor/command steps) + suite wiring. | `test/e2e/**` | — | `task gate-integration` **fails** only on the missing `describe provider` command |
| **T2** | **Data model (test-first).** Add `ProviderSpec{LastSyncedAt, LastSyncAttemptAt}` + `Spec` to `Provider`; add the optional `spec` block to `Provider.schema.json`; update the Provider note in `state.md`. Test: schema validates a synced Provider doc and rejects an unknown `spec` key. | `pkg/sauron/types/provider.go`, `spec/contracts/schemas/Provider.schema.json`, `spec/contracts/state.md` | T1 | `go test ./pkg/...` |
| **T3** | **Use case (test-first).** `DescribeProviderUseCase.Execute`: `Get` → none → `(nil, nil)` → read-error → `NewIOError`; wire `NewDescribeProviderUseCase` in `fx.go`. | `internal/usecase/usecase_describe_provider.go` (+`_test.go`), `internal/usecase/fx.go` | T2 | `go test ./internal/usecase/...` |
| **T4** | **View (test-first).** `view_describe_provider.go`: field set + selector (`name` forced first, dedup, unknown → usage error), `directory` derivation (`claude`→`~/.claude`, `zencoder`→`~/.zencoder`), `labels` → key-sorted section, the `no provider is set` line. | `internal/cmd/view_describe_provider.go` (+`_test.go`) | T3 | `go test -run DescribeProvider ./internal/cmd/...` |
| **T5** | **Command + spec docs (test-first).** `cmd_describe_provider.go` builder/handler (nil result → none-set line, exit 0); `AddCommand(DescribeProvider())` in `cmd_describe.go`; update 0006 `spec.md` FR-002, `contracts/describe-provider.md`, and `data/state.md` to the full field set. | `internal/cmd/cmd_describe_provider.go` (+`_test.go`), `internal/cmd/cmd_describe.go`, `spec/0006-describe-provider/{spec.md,contracts/describe-provider.md,data/state.md}` | T4 | `go build ./... && go test ./internal/cmd/...` |
| **T6** | **Closing gate.** Run the whole contract — the T1 e2e feature now passes. | — | T5 | **`task all`** green (test, lint, build, coverage ≥80%, security, integration) |

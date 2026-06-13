---
name: sauron-implementing-architecture
description: Use when writing or modifying Go code in this repository — the Use Case/Action orchestration pattern (UseCase[R Request], Action, the Request context object, Serve()/serve() + fx.Populate), the ports-and-adapters layout under internal/infrastructure with pkg/ ports, the storage internal capability (fx-injected afero.Fs), the root-command and package.json/git ldflags versioning, the no-rogue-goroutines (pond) rule, and the local Taskfile gates. Normative rules live in spec/contracts/architecture.md and CONSTITUTION.md.
---

# Implementing Sauron's Architecture

When writing Go for sauron, follow the normative
[architecture contract](../../../spec/contracts/architecture.md) and the
[Constitution](../../../CONSTITUTION.md).

**This skill overrides `golang-personal-architecture` for this repo.** Sauron uses
the **Use Case/Action** pattern, **not** services-as-interfaces, and the
infrastructure layout below — when they conflict, this skill wins.

## Procedural reminders

1. **Use Case = command entrypoint.** A command's business logic is a
   `UseCase[R Request]` (`Execute(R) error`), not a service. Reusable steps are
   `Action[R, P any]` (`Execute(context.Context, R) (*P, error)`). Both live in
   `internal/usecase` as `<Name>UseCase` / `<Name>Action`, in files
   `usecase_<name>.go` / `action_<name>.go`.
2. **The `Request` is a context object** (gin-style): it *extends*
   `context.Context` and exposes `Out() io.Writer`. A use case is **stateless** —
   its collaborators (the `pkg/` ports, the `storage` stores, the logger) are
   injected by uberfx; everything call-scoped arrives through the `Request`,
   which is built per invocation and never retained.
3. **`serve()` wires it.** A command's private `serve()` maps its flag struct +
   args into a concrete `Request`, resolves the use case with `fx.Populate`, and
   calls `Execute` — no business logic in the cobra layer.
4. **Ports & adapters.** Public interfaces live in `pkg/{registry,provider,backend}`;
   adapters live in `internal/infrastructure/<name>/<kind>` with a
   `NewFxOptions()`. `storage` is an **internal capability** (no `pkg/` port): its
   `afero.Fs` is fx-injected, and it locates files under
   `Configuration.HomeDirectory`. **`fx.go` is wiring-only** — just `NewFxOptions`
   and its supporting (unexported) provider helpers; interfaces, structs, and
   construction logic go in sibling files (`api.go`, `configuration.go`,
   `logger.go`, `<store>.go`), never in `fx.go`.
5. **No rogue goroutines.** No bare `go`. All concurrency runs on the fx-injected
   `pond` pool, whose lifecycle is bound to the `fx.App`.
6. **Versioning.** `AppName`/`AppVersion` come from `package.json`, `AppHash` from
   the git worktree hash, injected by `task build` via `-ldflags -X main.<var>`.
7. **Gates before done.** Verify with the local Taskfile targets — `task gate-lint`,
   `task test`, `task gate-coverage` (≥ 80%), `task gate-security`,
   `task gate-integration` (Linux + Docker) — or `task all`. See the
   [verification gate](../../../CONSTITUTION.md).
8. **Style.** Uber Go Style Guide, gocognit ≤ 15, parameter structs over >7 args,
   testify table-driven tests, `MockBased<Iface>` in `mock_based_<iface>.go`.
   **Doc comments are minimal** — one concise line per exported symbol, one
   package comment per package, none on trivial unexported helpers; never
   paraphrase the contract or narrate obvious code.
9. **Integration tests are out of scope here.** The black-box BDD suite lives in
   its own `test/e2e` module (godog + testcontainers, `replace` → root, imports
   only `pkg/`, `depguard`-banned from `internal/`); it does **not** follow Use
   Case/Action or ports-and-adapters. See the
   [`sauron-implementing-integration-tests`](../sauron-implementing-integration-tests/SKILL.md)
   skill and the architecture contract's
   [Integration tests](../../../spec/contracts/architecture.md) section.

# Architecture Contract

The normative technical contract for sauron's implementation: module identity,
toolchain, project layout, package responsibilities, dependency wiring, and the
approved dependency set. Every feature's implementation conforms to it.

## Module & toolchain

- Module path: `github.com/delfimarime/sauron`
- Go: `1.26`
- License: Apache-2.0 — every dependency must carry a permissive,
  Apache-2.0-compatible license.

## Project layout

Follows [golang-standards/project-layout](https://github.com/golang-standards/project-layout):

```
cmd/
  main.go                  binary entrypoint; defines AppName, AppVersion, AppHash
internal/
  cmd/
    root.go                root cobra command
    helper.go              NewApp() builder and common command helpers
    helper_flags.go        shared flag structs and their bind functions
    <sub-command>.go       one file per command
    <sub-command>_capability_<name>.go   a capability of a command
  config/
    fx.go                  NewFxOptions() fx.Option; exposes the Configuration struct
  telemetry/
    fx.go                  NewFxOptions() fx.Option; provides the zap+ECS logger
    constants.go           shared ECS log/trace field keys
  infrastructure/
    registry/
      fx.go                exposes NewFxOptions() fx.Option
      fs/                  filesystem registry adapter
      git/                 git registry adapter
      http/                HTTP registry adapter
    provider/
      fx.go                exposes NewFxOptions() fx.Option
      claude/              Claude provider adapter
      zencoder/            Zencoder provider adapter
    backend/
      fx.go                exposes NewFxOptions() fx.Option
      fs/                  filesystem backend adapter
      git/                 git backend adapter
      http/                HTTP backend adapter
    storage/
      fx.go                NewFxOptions() fx.Option; provides the afero.Fs and stores
      <store>.go           per-type store over the ~/.sauron state files
  usecase/
    fx.go                  NewFxOptions() fx.Option; provides use cases and actions
    usecase_<name>.go      a command's UseCase entrypoint
    action_<name>.go       a reusable Action a use case composes
pkg/
  registry/              public interfaces implemented by internal/infrastructure/registry/<kind>
  provider/              public interfaces implemented by internal/infrastructure/provider/<kind>
  backend/               public interfaces implemented by internal/infrastructure/backend/<kind>
```

The behavioral interfaces under `pkg/` are a public surface: external code may
implement new registries, providers, or backends against them. Their adapters
live under `internal/infrastructure/` — the driven-adapter layer reaching
external systems — and are never imported across adapter boundaries: callers
depend on the `pkg/` interfaces, not on a concrete adapter. `internal/infrastructure/`
also houses **internal capabilities** that are not public extension points —
[`storage`](#state-storage), which manipulates the `~/.sauron/` state — whose
types stay wholly within their package with no `pkg/` port. The transversal
framework modules (`internal/config`, `internal/telemetry`, `internal/cmd`) are
not adapters and stay at the `internal/` root.

## Dependency wiring (uberfx)

- Module packages own an `fx.go` exposing `NewFxOptions() fx.Option`
  (`internal/config/fx.go`, `internal/telemetry/fx.go`,
  `internal/infrastructure/registry/fx.go`,
  `internal/infrastructure/provider/fx.go`,
  `internal/infrastructure/backend/fx.go`,
  `internal/infrastructure/storage/fx.go`). Configuration is
  loaded with viper, but only the resulting `Configuration` struct is provided
  into the container — `*viper.Viper` is never placed in the fx graph.
  `Configuration` carries the resolved home as `HomeDirectory string`
  (see [Root command](#root-command)), the single value `storage` uses to locate
  the `~/.sauron/` state files.
- `internal/cmd/helper.go` provides
  `NewApp(ctx context.Context, opts ...fx.Option) *fx.App`. It **builds but does
  not start** a minimal app wired with the modules transversal to every command
  (telemetry, configuration, and the like), supplies the command's context
  (`cmd.Context()`) into the container, sets `fx.WithLogger` from the
  DI-provided zap logger (constructed by `internal/telemetry`), and appends the
  caller's `opts`.
- Each command owns its `fx.Option`s — constructing them directly or via
  `<package>.NewFxOptions()` — and passes them to `NewApp` from its `Serve()`.
  `cmd/main.go` bootstraps the binary.

## Root command

`internal/cmd/root.go` exposes
`func New(appName, appVersion, appHash string) (*cobra.Command, error)`, which
builds the root cobra command. `cmd/main.go` owns the ldflags build variables
(`AppName`, `AppVersion`, `AppHash`) and passes them in, so the root command is
constructible with arbitrary identity strings in a test.

`New` sets the command's version template — the output of `--version` and of the
bare root command — to this banner:

```
<AppName> v<AppVersion>
Hash <AppHash>
Home: <home>
```

`<home>` is the resolved home directory: `$SAURON_HOME` when set, the platform
default `~/.sauron` otherwise, exactly as fixed by the
[configuration data contract](configuration.md). `internal/config` owns that
resolution as a package-level function — callable eagerly by `New` (which runs at
bootstrap, before any `fx.App` exists) and also used to populate
`Configuration.HomeDirectory` for the fx graph, so the banner and `storage`
never diverge. `New` returns an error when the home cannot be resolved.

The root command is the **one exception** to the spec-and-contract rules: it has
no feature spec and no `contracts/command-line.md`; its behavior is fixed here in
the architecture contract.

## Command flags

Flags are bound into structs in package `internal/cmd`; command logic never
reads flags off the `*cobra.Command`.

- Flags shared across commands are defined once as small, concern-grouped
  structs (e.g. listing, dry-run, timeout) in
  `internal/cmd/helper_flags.go`, each paired with a `bind<Group>Flags` function
  that registers the flags and binds them to the struct. These are the shared
  flags defined by the CLI conventions.
- Each command declares its own `<command>Flags` struct that embeds the common
  group structs for the flags it shares and adds any command-specific fields.
- A command's public `Serve()` registers the flags into its `<command>Flags`
  value via the bind functions; the private `serve()` receives that populated
  struct — alongside the `context.Context` and the command's positional
  arguments — so the logic is tested without cobra.
- Flag values are not bound to viper; there is no flag→viper binding. Flags pass
  directly to `serve()`, independent of the persisted configuration in
  `internal/config`.

## Use Case orchestration

Business logic is organized as **use cases** (interactors), not services. A
command's entrypoint is a `UseCase`; the reusable capabilities a use case
composes are `Action`s. Both live in `internal/usecase`.

```go
type Request interface {
	context.Context
	Out() io.Writer
}

type UseCase[R Request] interface {
	Execute(R) error
}

type Action[R, P any] interface {
	Execute(context.Context, R) (*P, error)
}
```

- **`Request` is a context object**, in the spirit of `gin.Context`: it
  *extends* `context.Context` and adds `Out()`, the writer the command's output
  goes to, so a use case stays ignorant of where its bytes are printed.
  Persisted state is reached through the [`storage`](#state-storage) collaborator,
  not the `Request` — the filesystem is not part of the invocation environment. A
  concrete `Request` holds the per-invocation context; `golangci-lint`'s
  `containedctx` is scope-disabled for these types, because the context object is
  the sanctioned exception to the no-context-in-a-struct rule under
  [Coding standards](#coding-standards) — it *is* the context rather than storing
  one as data.
- **`UseCase` is the command entrypoint and is stateless.** Its collaborators —
  the `pkg/` ports (`pkg/registry`, `pkg/provider`, `pkg/backend`), the
  [`storage`](#state-storage) stores, and the zap logger — are injected by
  uberfx; everything call-scoped arrives through the `Request`, so a single
  instance is safe to reuse across invocations. `Execute` takes the `Request`
  alone (the context rides inside it) and returns only `error`; user-facing
  results are written to `Out()`.
- **`Action` is a reusable step** a use case composes. Unlike `UseCase`, it takes
  an explicit `context.Context` first parameter and a plain input `R` — left
  unconstrained on purpose: a pure action uses a value type, while an action that
  needs the IO environment declares its `R` as a `Request`. It returns
  `(*P, error)`.
- **Lifecycle.** A `Request` is constructed per invocation and passed directly to
  `Execute`; a use case or action never retains it beyond the call, so the
  embedded context cannot go stale.

### Layout & naming

`internal/usecase` owns both kinds and exposes `NewFxOptions() fx.Option`,
through which use cases and actions are provided with their `pkg/` ports, the
[`storage`](#state-storage) stores, and the logger. Types are named
`<Name>UseCase` / `<Name>Action`; their files are `usecase_<name>.go` /
`action_<name>.go`.

A command's private `serve()` maps its `<command>Flags` struct and positional
arguments into a concrete `Request`, resolves the use case from the container
with `fx.Populate`, and calls `Execute` — exercising the orchestration without
the cobra API, consistent with the [`Serve()`/`serve()` split](#testing).

## State storage

`internal/infrastructure/storage` owns all manipulation of Sauron's persisted
state — the files under `~/.sauron/` whose schema is fixed by the
[configuration data contract](configuration.md) (`registries.yaml`,
`backend.yaml`, `personas.yaml`, `track.yaml`, `settings.yaml`). It is the single
package that reads and writes those files; no use case or adapter touches them
directly.

- **It is an internal capability, not a public port.** Unlike registry, provider,
  and backend, storage has no `pkg/` interface — there is one way to persist
  state and no external implementation plugs in. Its types live entirely in
  `internal/infrastructure/storage` and are consumed by use cases.
- **The `afero.Fs` is injected by uberfx**, not carried on the `Request`:
  `storage`'s `fx.go` provides the filesystem (`afero.NewOsFs()` in production,
  an `afero.NewMemMapFs()` override in tests) into the container, and the stores
  depend on it. Centralizing the filesystem here is why the `Request` no longer
  exposes it.
- **The base path is `Configuration.HomeDirectory`.** The stores resolve every
  file under the home the fx-provided `Configuration` carries (`$SAURON_HOME` or
  `~/.sauron`, see [Root command](#root-command)); they never re-derive the home
  themselves.
- **Per-type stores.** State access is expected to split into one store per
  persisted concern (e.g. a registries store, a track store, a settings store)
  rather than a single catch-all, each provided through `storage`'s
  `NewFxOptions()`.

## Coding standards

Go code follows the [Uber Go Style Guide](https://github.com/uber-go/guide). In
addition:

- **Design principles.** DRY, SOLID, and YAGNI are first-class concerns in both
  production and test code.
- **No global mutable state.** Dependencies are supplied through uberfx
  (`fx.Option`s), not package-level variables.
- **Parameter structs.** A function that would take more than seven parameters
  takes a single parameter struct instead. `context.Context` is never moved into
  that struct — it stays an explicit parameter (conventionally first). In tests,
  ambient values such as `*testing.T` may likewise remain in the signature
  rather than the struct. The sole exception is the use-case `Request` context
  object (see [Use Case orchestration](#use-case-orchestration)): it deliberately
  *extends* `context.Context` rather than storing one as data, is built per
  invocation, and is passed live to `Execute`.
- **Cognitive complexity.** Each function stays at or below 15 (measured with
  `gocognit`); a higher value may be suppressed only with a comment that
  justifies it. A function below 3 is questionable — a trivial helper is not
  extracted unless it is reused by more than three callers, to avoid
  fragmentation.

## Telemetry & logging

Logging is structured: `go.uber.org/zap` encoded for Elastic Common Schema via
`go.elastic.co/ecszap`, conforming to the
[ECS field reference](https://www.elastic.co/docs/reference/ecs). Shared field
keys are defined once as constants in `internal/telemetry/constants.go` and
referenced from there — never written as scattered string literals. The
`internal/telemetry` package owns logger construction and its fx wiring.

## Testing

- Tests use the **Arrange / Act / Assert** structure.
- **Table-driven** tests are the encouraged default: one `Test<Subject>`
  function whose cases are parametrized, covering both successful and negative
  paths in the same table. Each case carries a comment stating its intent.
- Assertions use `testify` `assert`/`require`; collaborators are substituted
  with `testify` `mock` where applicable.
- **Mocks** live in the same package that defines the interface they implement,
  are named `MockBased<Interface>`, and reside in a file named
  `mock_based_<interface>.go`.
- **Coverage** is 90% per package as the ideal; a lower figure is permitted only
  when justified, and never falls below 75%.
- **Command testability.** Each command is split into a public `Serve()` that
  builds the `*cobra.Command` (the cobra wiring) and a private `serve()` — the
  same name, unexported — that holds the command's logic, decoupled from the
  cobra API, so the logic is unit tested without constructing a command.

## Approved dependencies

Only these are used. Adding a dependency requires a license review and a new row
here; nothing outside this list is imported without amending it. Licenses are
recorded as verified at vetting time.

| Dependency | Purpose | License |
|---|---|---|
| `github.com/spf13/cobra` | CLI command framework | Apache-2.0 |
| `github.com/spf13/viper` | Configuration management | MIT |
| `github.com/spf13/afero` | Filesystem abstraction; injected into `internal/infrastructure/storage` | Apache-2.0 |
| `net/http` (stdlib) | HTTP client | BSD-3-Clause |
| `os/exec` (stdlib) | Invoking external provider CLIs and the OS scheduler (`crontab`) | BSD-3-Clause |
| `github.com/go-git/go-git/v5` | Git operations | Apache-2.0 |
| `gopkg.in/yaml.v3` | YAML read/write | MIT and Apache-2.0 |
| `github.com/google/jsonschema-go` | JSON Schema validation | MIT |
| `go.uber.org/fx` | Dependency injection & lifecycle | MIT |
| `go.uber.org/zap` | Structured logging | MIT |
| `go.elastic.co/ecszap` | ECS-formatted zap encoder | Apache-2.0 |
| `github.com/stretchr/testify` | Test assertions and mocks (test scope) | MIT |

## Notes

- `os/exec` is scoped to invoking external provider CLIs and the OS scheduler
  (`crontab`) — the scheduler is a use-case concern, not an infrastructure
  adapter, so it introduces no new dependency; `github.com/go-git/go-git/v5`
  owns all git interaction.
- `github.com/google/jsonschema-go` is a young library — track its maturity
  before deeper reliance, per the dependency-scrutiny rule.

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
    helper.go              builds the uberfx application the commands use
    helper_flags.go        shared flag structs and their bind functions
    <sub-command>.go       one file per command
    <sub-command>_capability_<name>.go   a capability of a command
  config/                  spf13/viper init and configuration management
  telemetry/
    constants.go           shared ECS log/trace field keys
  repository/
    fx.go                  exposes NewFxOptions() fx.Option
    fs/                    filesystem repository adapter
    git/                   git repository adapter
    http/                  HTTP repository adapter
  provider/
    fx.go                  exposes NewFxOptions() fx.Option
    claude/                Claude provider adapter
    zencoder/              Zencoder provider adapter
pkg/
  repository/              public interfaces implemented by internal/repository/<type>
  provider/                public interfaces implemented by internal/provider/<type>
```

The behavioral interfaces under `pkg/` are a public surface: external code may
implement new repositories or providers against them. Implementations live
under `internal/` and are never imported across adapter boundaries — callers
depend on the `pkg/` interfaces, not on a concrete adapter.

## Dependency wiring (uberfx)

- Each adapter family package owns an `fx.go` exposing
  `NewFxOptions() fx.Option` (`internal/repository/fx.go`,
  `internal/provider/fx.go`).
- `cmd/main.go` bootstraps the binary; `internal/cmd/helper.go` assembles the
  `fx.Option`s into the application that the commands execute.

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
  rather than the struct.
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
| `github.com/spf13/afero` | Filesystem abstraction | Apache-2.0 |
| `net/http` (stdlib) | HTTP client | BSD-3-Clause |
| `os/exec` (stdlib) | Invoking external provider/target CLIs | BSD-3-Clause |
| `github.com/go-git/go-git/v5` | Git operations | Apache-2.0 |
| `gopkg.in/yaml.v3` | YAML read/write | MIT and Apache-2.0 |
| `github.com/google/jsonschema-go` | JSON Schema validation | MIT |
| `go.uber.org/fx` | Dependency injection & lifecycle | MIT |
| `go.uber.org/zap` | Structured logging | MIT |
| `go.elastic.co/ecszap` | ECS-formatted zap encoder | Apache-2.0 |
| `github.com/stretchr/testify` | Test assertions and mocks (test scope) | MIT |

## Notes

- `os/exec` is scoped to invoking external provider/target CLIs;
  `github.com/go-git/go-git/v5` owns all git interaction.
- `github.com/google/jsonschema-go` is a young library — track its maturity
  before deeper reliance, per the dependency-scrutiny rule.

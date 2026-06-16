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
Taskfile.yml               build & verification gate tasks (run with `task`)
package.json               AppName + AppVersion source (its name, version)
cmd/
  main.go                  binary entrypoint; AppName/AppVersion (from package.json) and AppHash (git worktree hash) set via ldflags
internal/
  cmd/
    root.go                root cobra command
    helper.go              NewApp() builder and common command helpers
    helper_flags.go        shared flag structs and their bind functions
    <sub-command>.go       one file per command
    <sub-command>_capability_<name>.go   a capability of a command
  config/
    fx.go                  NewFxOptions() fx.Option; wiring only (Configuration lives in configuration.go)
  telemetry/
    fx.go                  NewFxOptions() fx.Option; provides the zap+ECS logger
    logger.go              logger construction; references pkg/telemetry for shared ECS keys
  infrastructure/
    repository/            umbrella module; its fx.go aggregates the adapters + storage below
      fx.go                NewFxOptions() fx.Option; composes storage + registry + agent
      registry/            extension.Registry adapters
        fx.go              exposes NewFxOptions() fx.Option
        fs/                filesystem registry adapter
        git/               git registry adapter
        http/              HTTP registry adapter
      agent/               extension.Provider adapters (the destination agent environments)
        fx.go              exposes NewFxOptions() fx.Option
        claude/            Claude provider adapter
        zencoder/          Zencoder provider adapter
      storage/             manifest state over ~/.sauron (internal capability)
        fx.go              NewFxOptions() fx.Option; provides the afero.Fs and stores
        <store>.go         per-type store over the ~/.sauron state files
  usecase/
    fx.go                  NewFxOptions() fx.Option; provides use cases and actions
    usecase_<name>.go      a command's UseCase entrypoint
    action_<name>.go       a reusable Action a use case composes
pkg/
  http/                  public HTTP client: functional-options New() + composable round trippers (basic-auth, zap logger)
  telemetry/             shared ECS field-key vocabulary, referenced by public packages and internal/telemetry
  sauron/
    extension/           public ports (SPI): Registry, Provider — implemented under internal/infrastructure/repository
    types/               public domain & manifest types (Skill, Agent, Persona, Registry, Provider, Schedule, provenance)
test/
  e2e/                   external black-box integration tests — own go.mod (replace → root); godog + testcontainers; excluded from `go test ./...`
    testdata/            Gherkin .feature files
    integration_test.go  godog TestSuite entrypoint (no main)
    internal/            step definitions, scenario world, binary runner
dist/                    build output (git-ignored): the per-OS sauron binaries (sauron-<os>-<arch>) and coverage report
```

The public surface lives under `pkg/`. The **ports** are in `pkg/sauron/extension`
(`Registry`, `Provider`) — external code may implement new registries or providers
against them — and the shared **domain and manifest types** are in
`pkg/sauron/types` (data, not ports), spoken by the ports, `storage`, the use
cases, and the CLI output the `test/e2e` harness decodes. `pkg/` also carries two
public toolkits: `pkg/telemetry` (the ECS field-key vocabulary, see
[Telemetry & logging](#telemetry--logging)) and `pkg/http` (a composable HTTP
client). The port adapters live under `internal/infrastructure/repository/` — the
driven-adapter layer reaching external systems, grouped under a single
`repository` module whose `fx.go` aggregates them: `registry/{fs,git,http}`
implements `extension.Registry`, and `agent/{claude,zencoder}` implements
`extension.Provider` (a provider destination is modeled as an agent environment).
Adapters are never imported across boundaries — callers depend on the
`pkg/sauron/extension` ports, not a concrete adapter. The same `repository` module
also houses the **internal capability** [`storage`](#state-storage), which
manipulates the `~/.sauron/` state and has no `pkg/` port. The transversal
framework modules (`internal/config`, `internal/telemetry`, `internal/cmd`) are
not adapters and stay at the `internal/` root.

## Dependency wiring (uberfx)

- Module packages own an `fx.go` exposing `NewFxOptions() fx.Option`
  (`internal/config/fx.go`, `internal/telemetry/fx.go`, and
  `internal/infrastructure/repository/fx.go`, which aggregates its
  `registry/`, `agent/`, and `storage/` sub-modules — each of which owns its own
  `fx.go`). An `fx.go` holds only `NewFxOptions`
  and its supporting (unexported) provider helpers — it carries no business
  interfaces, structs, or construction logic; those live in sibling files
  (`api.go`, `configuration.go`, `logger.go`, `<store>.go`). Configuration is
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
- A `pond` (`github.com/alitto/pond/v2`) pool is among the transversal modules
  `NewApp` wires: it is constructed once, provided into the container as the sole
  sanctioned source of goroutines, and registered with the `fx.Lifecycle` so
  `OnStop` stops it and waits for in-flight tasks. This is what guarantees no
  goroutine outlives the process (see
  [No rogue goroutines](#coding-standards)). Components inject the pool rather
  than calling `go`.
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
no feature spec and no command contract; its behavior is fixed here in
the architecture contract.

## Build, versioning & gates

A `Taskfile` at the repo root — run with `task` — is the project's canonical
build and verification entrypoint. Its task names match the CI job names
one-to-one, so each CI job runs the identically-named target. It exposes at
least:

- `test` — runs the unit tests with the race detector and writes a coverage
  report under `dist/`.
- `gate-lint` — lints with `golangci-lint` (the enforcement point for the Uber
  style guide and the gocognit ≤15 ceiling).
- `build` — cross-compiles the binary `CGO_ENABLED=0` from a single toolchain to
  each supported target (`linux/amd64`, `darwin/arm64`, `darwin/amd64`), writing
  `dist/sauron-<os>-<arch>` (made executable) and injecting the version variables
  via the Go linker (`-ldflags`).
- `gate-coverage` — reads the `test` report and enforces the coverage gate: 90%
  ideal, failing below the project-level floor of 80%.
- `gate-security` — depends on `build` and runs `trivy` over the built
  `dist/sauron-linux-amd64` binary, enforcing the [verification gate](../../CONSTITUTION.md)
  (no CRITICAL, at most two HIGH), with accepted exceptions carried by a
  project-level ADR under `spec/architecture/`.
- `gate-integration` — depends on `build`; runs the black-box BDD suite in the
  `test/e2e` module against the **host's** binary
  (`SAURON_BIN=$ROOT/dist/sauron-$(go env GOOS)-$(go env GOARCH)`), so it runs on
  any platform with a Docker daemon (the suite provisions its dependencies via
  Testcontainers). The task carries **no OS guard** — the Linux restriction is a
  CI concern (see below), not a property of the task; on a Linux CI runner the
  host binary resolves to `sauron-linux-amd64`.
- `all` — builds and runs every gate.

### Continuous integration & delivery

The CI pipeline is provider-agnostic — GitHub Actions and GitLab CI are the
reference targets — and runs the Taskfile gates in dependency-gated stages:

1. `test` and `gate-lint` in parallel.
2. On their success, `build` and `gate-coverage` in parallel — `build`
   cross-compiles all targets (`CGO_ENABLED=0`) in one Linux job, and
   `gate-coverage` consumes the coverage report `test` published as an artifact.
3. On their success, `gate-security` — scanning the `dist/sauron-linux-amd64`
   binary `build` published as an artifact.
4. On its success, `gate-integration` — runs alone, **pinned to a Linux runner**
   (this is where the Linux-only constraint is enforced; the suite needs a Docker
   daemon), against the `dist/sauron-linux-amd64` artifact.
5. On its success and **only on the `main`/`master` branch**, `publish` —
   generates each binary's SHA-256 checksum and publishes every
   `dist/sauron-<os>-<arch>` binary and its `.sha256` as **release assets**
   (a GitHub Release / a GitLab Release).

### Versioning

Go exposes no built-in version mechanism, so `AppName` and `AppVersion` are kept
in a root `package.json` (its `name` and `version`), and `AppHash` is the short
git hash of the worktree. The `build` task injects all three into the
`cmd/main.go` variables with `-ldflags -X main.<var>=<value>`; they are not
sourced any other way. `package.json`'s `version` is the strict-SemVer source of
truth, bumped by hand to match the change type (see
[CONTRIBUTING.md](../../CONTRIBUTING.md)); CI only *decorates* it into the
artifact label, via the build task's overridable `AppVersion`, by build context:

| Context | `AppVersion` | Published |
|---|---|---|
| local `task build` | `<version>` | no |
| `main`/`master` branch | `<version>-RELEASE` | yes |
| PR/MR into `main`/`master` | `<version>-PRE-RELEASE.<ci-number>` | no |
| any other build | `<version>-SNAPSHOT.<ci-number>` | no |

All three decorations are valid SemVer pre-release identifiers (hyphen-prefixed),
so the artifact label remains SemVer-parseable.

`task`, `golangci-lint`, `trivy`, and `jq` are development/CI tooling, not module
dependencies, so they are absent from the
[approved-dependency table](#approved-dependencies).

## Command flags

Flags are bound into structs in package `internal/cmd`; command logic never
reads flags off the `*cobra.Command`.

- Flags shared across commands are defined once as small, concern-grouped
  structs (e.g. listing, paging, dry-run, timeout) in
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
  the `pkg/` ports (`pkg/registry`, `pkg/provider`), the
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
`track.yaml`, `settings.yaml`). It is the single
package that reads and writes those files; no use case or adapter touches them
directly.

- **It is an internal capability, not a public port.** Unlike registry and
  provider, storage has no `pkg/` interface — there is one way to persist
  state and no external implementation plugs in. Its types live entirely in
  `internal/infrastructure/storage` and are consumed by use cases.
- **Files are multi-document manifest streams.** Each file holds Kubernetes-style
  documents (`apiVersion: sauron.raitonbl.com/v1`, `kind`, `metadata`, `spec`);
  storage decodes and encodes the stream and validates every document against its
  per-kind JSON Schema (under `spec/contracts/schemas/`) with
  `github.com/google/jsonschema-go`. Writes are atomic (write-temp + rename) and
  serialized by a lockfile under the home, so a scheduled run and a manual command
  never corrupt a file.
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
- **No rogue goroutines.** No component starts a bare goroutine. All concurrency
  runs on a [pond](https://github.com/alitto/pond) (`github.com/alitto/pond/v2`)
  pool injected via uberfx; the raw `go` keyword is disallowed in production code.
  The pool's lifecycle is bound to the `fx.App` (see
  [Dependency wiring](#dependency-wiring-uberfx)), so application shutdown stops
  the pool and waits for every in-flight task — no goroutine outlives the
  process. Enforced in review, and by a linter ban on `go` statements outside the
  pool wiring where the toolchain supports it.
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
- **Doc comments are minimal.** A single concise doc line on each exported
  symbol (and one package comment per package); no comment on a trivial
  unexported helper. Comments clarify what code cannot — they never paraphrase
  this contract or narrate the obvious.

## Telemetry & logging

Logging is structured: `go.uber.org/zap` encoded for Elastic Common Schema via
`go.elastic.co/ecszap`, conforming to the
[ECS field reference](https://www.elastic.co/docs/reference/ecs). Shared ECS field
keys are defined once as constants in **`pkg/telemetry`** — the public home — so
public packages (e.g. `pkg/http`) and `internal/telemetry` reference the same
vocabulary without a public→internal dependency, and are never written as
scattered string literals. A key lives in exactly one place: keys emitted by
public packages live in `pkg/telemetry`; any internal-only key stays in
`internal/telemetry`, which references `pkg/telemetry` for the shared set and never
redefines it. `internal/telemetry` owns logger construction and its fx wiring.

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
- **Coverage** ideal is 90%; the [verification gate](../../CONSTITUTION.md)
  enforces a project-level floor of 80%, a lower figure permitted only when
  justified.
- **Command testability.** Each command is split into a public `Serve()` that
  builds the `*cobra.Command` (the cobra wiring) and a private `serve()` — the
  same name, unexported — that holds the command's logic, decoupled from the
  cobra API, so the logic is unit tested without constructing a command.

## Integration tests

The black-box BDD suite lives in its own module, `test/e2e`
(`github.com/delfimarime/sauron/test/e2e`), under the project-layout `/test`
directory. It is **not** bound by the Use Case/Action or ports-and-adapters
rules — it is a test harness.

- **Graybox.** Steps `exec` the built `dist/sauron-<os>-<arch>` binary (located
  via the `SAURON_BIN` environment variable) and decode its output into the
  public `pkg/` types for type-safe assertions — an external consumer of the
  `pkg/` surface.
- **Own module, `replace` to root.** `test/e2e/go.mod` requires the root module
  and resolves it with `replace github.com/delfimarime/sauron => ../..`, so it
  needs no version tag. Its dependencies — `github.com/cucumber/godog`,
  `github.com/testcontainers/testcontainers-go`, `github.com/stretchr/testify` —
  live in this module's `go.mod` only and are **absent from the
  approved-dependency table** below, which governs the production module alone.
- **`pkg/`-only.** The harness imports `pkg/` and never `internal/`. Go's
  `internal/` rule does not enforce this (the harness import paths share the root
  module prefix), so it is enforced by a `depguard` rule in the module's
  `golangci-lint` config.
- **Gherkin.** Feature files are `test/e2e/testdata/*.feature`; the runner is
  `test/e2e/integration_test.go` (`godog.TestSuite`, no `main`), invoked by the
  `gate-integration` task.
- **Platform.** The suite runs on **any host with a Docker daemon** (Testcontainers
  needs one); `gate-integration` exercises the host's own `sauron-<os>-<arch>`
  binary, so a developer on macOS runs it against the `darwin` build. **CI** pins
  the gate to a Linux runner — that is the only Linux-only constraint, and it is a
  CI policy, not a property of the suite or the task.
- **Hermeticity.** Per-scenario git and HTTP dependencies are provisioned in-test
  via Testcontainers; the concrete fixture strategy is still being settled.

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
| `github.com/alitto/pond/v2` | Worker-pool / goroutine management (the only sanctioned source of goroutines) | MIT |
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
- The dependency set is scanned for known vulnerabilities with `trivy` (or an
  equivalent scanner) as part of the verification gate (CONSTITUTION, Chapter IV,
  Article 2): zero CRITICAL and at most two HIGH findings per scan, any exception
  carried by a project-level ADR under `spec/architecture/`.

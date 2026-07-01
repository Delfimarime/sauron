# Architecture Contract

The normative technical contract for sauron's implementation. Every feature's
implementation conforms to it. It covers:

- **[Module & toolchain](#module--toolchain)** and the **[approved
  dependencies](#approved-dependencies)** — what the code is built on.
- **[Project layout](#project-layout)** and **[dependency
  wiring](#dependency-wiring-uberfx)** — where code lives and how uberfx composes
  it.
- **[Root command](#root-command)** and **[command structure](#command-structure)** —
  the cobra surface, flag binding, the builder/handler split, and error reporting.
- **[Use Case orchestration](#use-case-orchestration)** and **[state
  storage](#state-storage)** — the business-logic and persistence layers.
- **[Coding standards](#coding-standards)**, **[telemetry &
  logging](#telemetry--logging)**, **[testing](#testing)**, and **[integration
  tests](#integration-tests)** — how the code is written, observed, and verified.
- **[Build, versioning & delivery](#build-versioning--delivery)** — a pointer to
  the [delivery contract](delivery.md).

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
    cmd_root.go            root cobra command
    helper.go              NewApp() builder and shared command helpers
    helper_flags.go        shared flag-group structs and their bind functions
    cmd_<verb>.go          a command group (e.g. cmd_set.go, cmd_list.go)
    cmd_<verb>_<noun>.go   a command in that group (e.g. cmd_set_registry.go)
    view_<name>.go         rendering for the paired cmd_<name>.go: turns a use case's domain result + view options (selected fields, sort, search) into the table/descriptor and its bytes (stdlib text/tabwriter); cobra-free, pure value types, no fx — the command layer's view, not a separate package
  config/
    fx.go                  NewFxOptions() fx.Option; wiring only (Configuration lives in configuration.go)
  telemetry/
    fx.go                  NewFxOptions() fx.Option; provides the zap+ECS logger
    logger.go              logger construction; references pkg/telemetry for shared ECS keys
  infrastructure/
    repository/            umbrella module; its fx.go aggregates the adapters + storage below
      fx.go                NewFxOptions() fx.Option; composes storage + registry + agent
      registry/            extension.Registry adapters (one package; one file per transport)
        fx.go              exposes NewFxOptions() fx.Option
        api/               shared adapter primitives: error classes, option helpers, the directory-backed source.FileSystem
        git_filesystem.go  git adapter (shallow clone)
        rest_filesystem.go HTTP adapter (REST client)
      agent/               extension.Provider adapters (the destination agent environments)
        fx.go              exposes NewFxOptions() fx.Option
        claude/            Claude provider adapter
        zencoder/          Zencoder provider adapter
      storage/             manifest state over ~/.sauron (internal capability)
        fx.go              NewFxOptions() fx.Option; provides the afero.Fs and stores
        <store>.go         per-type store over the ~/.sauron state files
  usecase/
    fx.go                  NewFxOptions() fx.Option; provides use cases and actions
    usecase_<name>.go      a UseCase — a command entrypoint or a composed step
pkg/
  http/                  public HTTP client: functional-options New() + composable round trippers (basic-auth, zap logger)
  telemetry/             shared ECS field-key vocabulary, referenced by public packages and internal/telemetry
  sauron/
    extension/           public ports (SPI): Registry, Provider — implemented under internal/infrastructure/repository
    source/              public port: FileSystem/File/Stat — the content view a Registry.Open() returns
    marketplace/         public client SDK for the Registry HTTP API (resty-backed); used by the http registry adapter
    types/               public domain & manifest types (Skill, Agent, Registry, Provider)
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
`repository` module whose `fx.go` aggregates them: `registry/` — one package with
a file per transport (git, http) over a shared `registry/api` of
common primitives — implements `extension.Registry`, and `agent/{claude,zencoder}`
implements `extension.Provider` (a provider destination is modeled as an agent
environment).
Adapters are never imported across boundaries — callers depend on the
`pkg/sauron/extension` ports, not a concrete adapter. The same `repository` module
also houses the **internal capability** [`storage`](#state-storage), which
manipulates the `~/.sauron/` state and has no `pkg/` port. The transversal
framework modules (`internal/config`, `internal/telemetry`, `internal/cmd`) are
not adapters and stay at the `internal/` root. Rendering is **not** a separate
module — it lives inside `internal/cmd` as cobra-free `view_<name>.go` files,
since the command layer is its only consumer (see
[Command structure](#command-structure)).

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
  `<package>.NewFxOptions()` — and passes them to `NewApp` from its command
  builder. `cmd/main.go` bootstraps the binary.

## Root command

`internal/cmd/cmd_root.go` exposes
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
[state data contract](state.md). `internal/config` owns that
resolution as a package-level function — callable eagerly by `New` (which runs at
bootstrap, before any `fx.App` exists) and also used to populate
`Configuration.HomeDirectory` for the fx graph, so the banner and `storage`
never diverge. `New` returns an error when the home cannot be resolved.

The root command is the **one exception** to the spec-and-contract rules: it has
no feature spec and no command contract; its behavior is fixed here in
the architecture contract.

## Build, versioning & delivery

How sauron is built, gated, versioned, and shipped — the `Taskfile` gates, the
CI/CD pipeline, and the version-decoration scheme — is the
[delivery contract](delivery.md). The gate names there (`test`, `gate-lint`,
`build`, `gate-coverage`, `gate-security`, `gate-integration`, `all`) are the
enforcement points for the standards this contract defines (coverage target,
gocognit ceiling, the approved-dependency set, the CGO-free build).

## Command structure

The cobra layer is thin: a command is a **builder** that wires cobra and a
**handler** that holds the logic, split so the logic is tested without cobra. This
shape is canonical — the [use-case](#use-case-orchestration) and
[testing](#testing) sections refer here rather than restating it.

- **Builder / handler split.** A command's public **builder is named for the
  command verb** (`Add()` for `add`; a subcommand follows suit, e.g.
  `AddRegistry()`, `ListRegistries()`); it builds the `*cobra.Command` and binds
  its flags. The private **handler is named `<verb><Noun>()`** (e.g.
  `addRegistry()`, `listRegistries()`); it receives the populated flag struct —
  alongside the `context.Context` and positional arguments — builds the use-case
  input, calls `Execute`, and renders the returned result to stdout through the
  command layer's `view_<name>.go` rendering, so the logic is tested without
  cobra. View flags
  (`--fields`, `--sort`) are validated at this boundary, yielding a usage error
  before the use case runs. `Serve()`/`serve()` names apply only to a server's
  `serve` command, never to an action command.
- **One file per command, named `cmd_<name>.go`.** A command's builder and handler
  live together in `cmd_<name>.go`, where `<name>` is the command path — `cmd_set.go`
  for the `set` group, `cmd_set_registry.go` for `set registry`, `cmd_root.go` for the
  root command. The `cmd_` prefix pairs the file with — and visually separates it from
  — the `view_<name>.go` that renders that command's result.
- **Flags are bound into structs** in `internal/cmd`; command logic never reads
  flags off the `*cobra.Command`. Flags shared across commands are defined once as
  small, concern-grouped structs (listing, paging, dry-run, timeout) in
  `internal/cmd/helper_flags.go`, each paired with a `bind<Group>Flags` function;
  a command's own `<command>Flags` struct embeds the groups it shares and adds its
  command-specific fields. Flag values are not bound to viper — they pass directly
  to the handler, independent of the persisted `internal/config`.
- **One error-reporting site.** A handler returns a **classified error** and never
  prints it; `cmd/main.go` is the single place that maps the error to an exit code
  and writes exactly one `error: <message>` line to stderr — no per-handler print
  paired with an `IsReported` flag.
- **Classified errors.** A use case returns a `usecase.Error{Type, Reason}`: the
  `Type` buckets the failure (usage, conflict, unreachable, validation, io) and
  `Reason` is the message. `cmd/main.go` maps the `Type` to the process exit code —
  a usage class to `2`, every other class to `1`, success to `0` — per the
  [CLI contract](cli.md) exit-status semantics. The handler never chooses an exit
  code itself.

## Use Case orchestration

Business logic is organized as **use cases** (interactors), not services. A
command's entrypoint is a `UseCase`; a use case may compose other use cases as
reusable steps. They all live in `internal/usecase`.

```go
type UseCase[I, P any] interface {
	Execute(ctx context.Context, in I) (*P, error)
}
```

- **A use case returns a result, never bytes.** `Execute` answers the question
  and returns a *presentation-agnostic* `*P` — domain objects from
  `pkg/sauron/types`, or a small struct composed of them — alongside a classified
  `*Error`. It never renders: no `Table`/`Descriptor`, no `io.Writer`, no field
  projection. How the result is displayed is the client's decision, performed by
  the command layer's [`view_<name>.go`](#project-layout) rendering after
  `Execute` returns (see [Command structure](#command-structure)). A use case is
  thus ignorant of presentation *entirely* — not merely of the output
  destination — which is the separation an `Out()` writer could not provide.
- **One shape, two roles.** A `UseCase` is either a command's entrypoint or a
  reusable step another use case composes — there is no separate `Action` type.
  Every use case takes an explicit `context.Context` first and a typed input `in`
  (a value type, or an empty struct for a parameterless query), and returns
  `(*P, error)`. There is no `Request` context object and no `Out()`: the
  per-invocation context is the explicit first parameter, and call-scoped
  *business* input is `in`. *View* options (field selection, sort, search) are
  **not** use-case input — they belong to the client.
- **Composition is acyclic by discipline.** A use case may call another use case.
  Nothing structurally enforces a direction, so keeping the call graph acyclic is
  the developer's responsibility. **Boundary concerns** — the top-level telemetry
  span, error classification at the edge, and the 1:1 command↔entrypoint mapping —
  belong to the command-invoked use case; a composed use case assumes it already
  runs inside that boundary and does not re-establish it.
- **Resource-acquiring use cases may return a port.** The `(*P, error)` shape
  has one sanctioned exception: an internal use case whose role is to *open* a
  resource returns the live `pkg/` port it acquires rather than a `*P` product —
  e.g. `OpenRegistry` returns a `source.FileSystem`. Such a seam is composed only by
  other use cases, never resolved by a handler and never rendered, so wrapping the
  handle in a `*P` would be ceremony with no benefit. It remains stateless and
  takes `context` first.
- **`UseCase` is stateless.** Its collaborators — the `pkg/` ports
  (`pkg/sauron/extension`), the [`storage`](#state-storage) stores, and the zap
  logger — are injected by uberfx; everything call-scoped arrives as `ctx`/`in`,
  so a single instance is safe to reuse across invocations. Persisted state is
  reached through the [`storage`](#state-storage) collaborator, not the input —
  the filesystem is not part of the invocation environment.
- **Lifecycle.** Inputs are plain values constructed per call and passed to
  `Execute`; there is no context object to retain, so nothing can go stale.

### Layout & naming

`internal/usecase` exposes `NewFxOptions() fx.Option`, through which use cases are
provided with their `pkg/` ports, the [`storage`](#state-storage) stores, and the
logger. Types are named `<Name>UseCase`; each lives in `usecase_<name>.go`.

The cobra **handler** that drives a use case — its naming and the builder/handler
split — is fixed under [Command structure](#command-structure). It maps the flag
struct and positional arguments into the use-case input, resolves and runs the use
case from the container with `fx.Invoke` — which calls `Execute` inside the started
fx lifecycle on the run context (resolving with `fx.Populate` and then calling
`Execute` is equally acceptable) — and renders the returned `*P` result to the
command's stdout through the `view_<name>.go` rendering.

## State storage

`internal/infrastructure/repository/storage` owns all manipulation of Sauron's
persisted state — the files under `~/.sauron/` whose schema is fixed by the
[state data contract](state.md) (`track.yaml`, `settings.yaml`). It is the single
package that reads and writes those files; no use case or adapter touches them
directly.

- **It is an internal capability, not a public port.** Unlike registry and
  provider, storage has no `pkg/` interface — there is one way to persist
  state and no external implementation plugs in. Its types live entirely in
  `internal/infrastructure/repository/storage` and are consumed by use cases.
- **Files are multi-document manifest streams.** Each file holds Kubernetes-style
  documents (`apiVersion: sauron.raitonbl.com/v1`, `kind`, `metadata`, `spec`);
  storage decodes and encodes the stream and validates every document **it reads**
  against its per-kind JSON Schema (under `spec/contracts/schemas/`) with
  `github.com/google/jsonschema-go`. Validation guards externally-modifiable
  input: the home files are hand-editable, so every document is validated on load
  before any action uses it, and an invalid document is a runtime error. Documents
  the app itself authors are written **without** re-validation — they are
  constructed from typed values, so schema validation is a concern for input, not
  for app output. Writes are atomic (write-temp + rename) and serialized by a
  lockfile under the home, so a periodic run and a manual command never corrupt a
  file.
- **The `afero.Fs` is injected by uberfx**, not carried on the `Request`:
  `storage`'s `fx.go` provides the filesystem (`afero.NewOsFs()` in production,
  an `afero.NewMemMapFs()` override in tests) into the container, and the stores
  depend on it. Centralizing the filesystem here is why the `Request` no longer
  exposes it.
- **The base path is `Configuration.HomeDirectory`.** The stores resolve every
  file under the home the fx-provided `Configuration` carries (see
  [Root command](#root-command) for how it is resolved); they never re-derive the
  home themselves.
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
  runs on the fx-injected [pond](https://github.com/alitto/pond)
  (`github.com/alitto/pond/v2`) pool; the raw `go` keyword is disallowed in
  production code, enforced in review and by a linter ban on `go` statements
  outside the pool wiring where the toolchain supports it. The pool's lifecycle
  binding — so no goroutine outlives the process — is fixed under
  [Dependency wiring](#dependency-wiring-uberfx).
- **Parameter structs.** A function that would take more than seven parameters
  takes a single parameter struct instead. `context.Context` is never moved into
  that struct — it stays an explicit parameter (conventionally first); in tests,
  ambient values such as `*testing.T` may likewise stay in the signature.
- **Cognitive complexity.** Each function stays at or below 15 (measured with
  `gocognit`); a higher value may be suppressed only with a comment that
  justifies it. A function below 3 is questionable — a trivial helper is not
  extracted unless it is reused by more than three callers, to avoid
  fragmentation.
- **Doc comments are minimal.** A single concise doc line on each exported
  symbol; no comment on a trivial unexported helper. Comments clarify what code
  cannot — they never paraphrase this contract or narrate the obvious.
- **Package comments live in `doc.go`.** Every package carries exactly one
  package comment, placed in a dedicated `doc.go` that holds only that comment
  and the `package` clause — never on an arbitrary source file.
- **A file leads with its primary type.** The type a file is named for comes
  first; its constructor and methods follow. The primary type is never buried
  below the helpers.
- **Behavior belongs to its type.** A function used by exactly one struct and
  nowhere else is **bound to that struct** as a method — never a free function that
  takes the struct (or its data) as a parameter — so the behavior lives with the
  type that owns it. A function used by **two or more** structs is **shared** and
  stands alone — preferentially in a `helper.go` (or a topical `helper_<area>.go`
  variant), otherwise in a dedicated package — and is then type-agnostic,
  frequently generic. Reuse, not size, is the deciding test: a stateless or tiny
  function still binds to its sole user, and only genuine reuse across types lifts
  it into a helper.
- **No test-only seams.** Production code does not grow an injectable indirection
  (a function-type field, a package-level var) solely so a test can replace it.
  Use the standard library directly and exercise it through the real graph or
  `t.Setenv`; reserve interfaces and injection for genuine production
  collaborators.

## Telemetry & logging

Logging is structured: `go.uber.org/zap` encoded for Elastic Common Schema via
`go.elastic.co/ecszap`, conforming to the
[ECS field reference](https://www.elastic.co/docs/reference/ecs). Every log is
ECS-compatible: standard fields use their canonical ECS keys (`event.action`,
`service.name`, …), and any field not in standard ECS is namespaced under the
single custom top-level key `sauron` (e.g. `sauron.registry.name`,
`sauron.registry.transport`) — never a bare key like `name`. Shared ECS field
keys are defined once as constants in **`pkg/telemetry`** — the public home — so
public packages (e.g. `pkg/http`) and `internal/telemetry` reference the same
vocabulary without a public→internal dependency, and are never written as
scattered string literals. A key lives in exactly one place: keys emitted by
public packages live in `pkg/telemetry`; any internal-only key stays in
`internal/telemetry`, which references `pkg/telemetry` for the shared set and never
redefines it. `internal/telemetry` owns logger construction and its fx wiring.

## Testing

- **Test-first (TDD).** A change starts from a failing test: the unit test —
  and, for user-observable behavior, the `test/e2e` Gherkin scenario
  ([integration tests](#integration-tests)) — is written or updated to fail
  *before* the production code that satisfies it. Plans and task breakdowns
  sequence this explicitly, test before implementation
  ([AUTHORING.md](../AUTHORING.md)).
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
- **Command testability.** The builder/handler split fixed under
  [Command structure](#command-structure) exists so a command's logic is unit
  tested through its handler, without constructing a `*cobra.Command`.

## Integration tests

The black-box BDD suite lives in its **own module**, `test/e2e`
(`github.com/delfimarime/sauron/test/e2e`), under the project-layout `/test`
directory; its harness reference is [`test/e2e/HARNESS.md`](../../test/e2e/HARNESS.md)
(runtime/Source architecture, controllers, fixtures, tags, the gate). The
two facts that bind on *this* contract:

- **Module boundary & dependency isolation.** `test/e2e/go.mod` resolves the root
  with `replace github.com/delfimarime/sauron => ../..`; its test-only
  dependencies live in that module's `go.mod` and are **absent from the
  [approved-dependency table](#approved-dependencies)** below, which governs the
  production module alone.
- **`pkg/`-only graybox.** The harness `exec`s the built binary and decodes its
  output into the public `pkg/` types — never importing `internal/` (enforced by a
  `depguard` rule, since Go's `internal/` rule does not apply across the shared
  module prefix). It is **not** bound by the Use Case or ports-and-adapters
  rules.

The gate that runs it (`gate-integration`) is defined in the
[delivery contract](delivery.md).

## Approved dependencies

Only these are used. Adding a dependency requires a license review and a new row
here; nothing outside this list is imported without amending it. Licenses are
recorded as verified at vetting time.

| Dependency | Purpose | License |
|---|---|---|
| `github.com/spf13/cobra` | CLI command framework | Apache-2.0 |
| `github.com/spf13/viper` | Configuration management | MIT |
| `github.com/spf13/afero` | Filesystem abstraction; injected into `internal/infrastructure/repository/storage` | Apache-2.0 |
| `net/http` (stdlib) | HTTP client (`pkg/http` toolkit) | BSD-3-Clause |
| `github.com/go-resty/resty/v2` | REST client for the `http` registry adapter | MIT |
| `os/exec` (stdlib) | Invoking external provider CLIs | BSD-3-Clause |
| `github.com/go-git/go-git/v5` | Git operations | Apache-2.0 |
| `gopkg.in/yaml.v3` | YAML read/write | MIT and Apache-2.0 |
| `github.com/google/jsonschema-go` | JSON Schema validation | MIT |
| `go.uber.org/fx` | Dependency injection & lifecycle | MIT |
| `github.com/alitto/pond/v2` | Worker-pool / goroutine management (the only sanctioned source of goroutines) | MIT |
| `go.uber.org/zap` | Structured logging | MIT |
| `go.elastic.co/ecszap` | ECS-formatted zap encoder | Apache-2.0 |
| `github.com/stretchr/testify` | Test assertions and mocks (test scope) | MIT |

## Notes

- `os/exec` is scoped to invoking external provider CLIs;
  `github.com/go-git/go-git/v5` owns all git interaction.
- `github.com/google/jsonschema-go` is a young library — track its maturity
  before deeper reliance, per the dependency-scrutiny rule.
- The dependency set is scanned for known vulnerabilities with `trivy` (or an
  equivalent scanner) as part of the verification gate (CONSTITUTION, Chapter IV,
  Article 2): zero CRITICAL and at most two HIGH findings per scan, any exception
  carried by a project-level ADR under `spec/architecture/`.

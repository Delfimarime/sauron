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
    <sub-command>.go       one file per command
  config/                  spf13/viper init and configuration management
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

## Notes

- `os/exec` is scoped to invoking external provider/target CLIs;
  `github.com/go-git/go-git/v5` owns all git interaction.
- `github.com/google/jsonschema-go` is a young library — track its maturity
  before deeper reliance, per the dependency-scrutiny rule.

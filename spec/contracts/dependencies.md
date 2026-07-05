# Dependencies Contract

The normative contract for sauron's approved third-party dependency set. It
specializes the [architecture contract](architecture.md), which owns code
structure and wiring.

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

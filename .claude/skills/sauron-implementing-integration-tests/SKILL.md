---
name: sauron-implementing-integration-tests
description: Use when writing or modifying sauron's black-box integration tests — the test/e2e Go module (godog Gherkin under go test, testcontainers, replace → root), where features and step definitions live, the graybox pattern (exec the built binary, assert via pkg/ types), the SAURON_BIN contract, the depguard internal/ ban, and the gate-integration Taskfile target. Specializes the test skills for the e2e surface.
---

# Implementing Sauron's Integration Tests

Sauron's integration tests are a **black-box BDD suite** in its own module,
`test/e2e`. They drive the built binary and verify behaviour through Gherkin.
Normative rules: the architecture contract's
[Integration tests](../../../spec/contracts/architecture.md) section and the
[Constitution](../../../CONSTITUTION.md) verification gate.

This builds on `writing-unit-tests` → `golang-personal-tests` (testify,
table-driven, AAA) and `developing-code` → `golang-personal-code` (Uber style,
gocognit ≤ 15). It does **not** apply `sauron-implementing-architecture` — the
harness is exempt from the Use Case pattern and ports-and-adapters.

## Where the tests live

    test/e2e/
      go.mod                module .../sauron/test/e2e; require root + replace → ../..
      integration_test.go   godog.TestSuite entrypoint — func TestFeatures(t); NO main
      testdata/*.feature     Gherkin features (one per command / behaviour)
      internal/             step definitions, scenario "world", binary runner
      .golangci.yml         depguard: ban github.com/delfimarime/sauron/internal/...

## Procedural reminders

1. **No `main`; run under `go test`.** godog (v0.10+) runs inside `go test`:
   `integration_test.go` builds a `godog.TestSuite` with
   `Options{ Paths: []string{"testdata"}, TestingT: t }` and runs it. There is no
   `main`, so no duplicate-`main` problem, and the suite is invisible to the root
   module's `go test ./...` (separate module boundary).
2. **Graybox, not white-box.** Steps `exec` the built binary at `$SAURON_BIN` (an
   absolute path the `gate-integration` task injects), capture stdout/stderr and
   exit code, and decode output into the public `pkg/` types for assertions.
   Never call use cases or `internal/` code in-process — that is unit-test
   territory and is `depguard`-banned.
3. **`pkg/` only.** Import `pkg/{registry,provider,backend}` for typed
   assertions; never `internal/`. Go will not stop an `internal/` import here
   (shared module path prefix), so the `depguard` rule is the real guard — keep
   it.
4. **Dependencies stay local.** `github.com/cucumber/godog`,
   `github.com/testcontainers/testcontainers-go`, and `github.com/stretchr/testify`
   live in `test/e2e/go.mod` only. Never add them to the root module or its
   approved-dependency table.
5. **Hermetic via Testcontainers.** Provision each scenario's dependencies (git
   endpoints, HTTP) from an ephemeral container started by the harness — no
   public-internet dependence in a blocking gate. CI notes: GitHub `ubuntu-*`
   runners have Docker out of the box; GitLab needs docker-in-docker
   (`TESTCONTAINERS_HOST_OVERRIDE`, often `TESTCONTAINERS_RYUK_DISABLED=true`).
   The concrete fixture/ssh strategy (ADR-0002 remotes are ssh-only) is still
   being settled — confirm before relying on it.
6. **Linux only.** The suite runs on Linux (Testcontainers needs a Docker
   daemon). macOS binaries are built and published but not exercised here.
7. **AAA & reuse.** Inside step definitions and helpers, keep testify assertions
   and Arrange/Act/Assert structure; factor repeated setup into the scenario
   "world" rather than copy-paste across steps.
8. **Run it.** `cd test/e2e && SAURON_BIN=$ROOT/dist/sauron-linux-amd64 go test ./...`
   — the `gate-integration` Taskfile target. Build the binary first.

# Delivery Contract

The normative contract for how sauron is **built, versioned, gated, and
shipped** — the Taskfile gates, the CI/CD pipeline, and the version-decoration
scheme. It is the operational companion to the
[architecture contract](architecture.md) (which owns code structure, wiring, and
the approved dependency set) and realizes the
[verification gate](../../CONSTITUTION.md) of the Constitution. The
`sauron-operating-ci` skill and the `sauron-gatekeeper` agent are governed by
this file.

## Build & verification gates

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
  `dist/sauron-linux-amd64` binary, enforcing the
  [verification gate](../../CONSTITUTION.md) (no CRITICAL, at most two HIGH),
  with accepted exceptions carried by a project-level ADR under
  `spec/architecture/`.
- `gate-integration` — depends on `build`; runs the black-box BDD suite in the
  `test/e2e` module against the **host's** binary
  (`SAURON_BIN=$ROOT/dist/sauron-$(go env GOOS)-$(go env GOARCH)`), so it runs on
  any platform with a Docker daemon (the suite provisions its dependencies via
  Testcontainers). The task carries **no OS guard** — the Linux restriction is a
  CI concern (see below), not a property of the task; on a Linux CI runner the
  host binary resolves to `sauron-linux-amd64`. The harness and its conventions
  are owned by [`test/e2e/CONSTITUTION.md`](../../test/e2e/CONSTITUTION.md).
- `all` — builds and runs every gate.

## Continuous integration & delivery

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

## Versioning

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
dependencies, so they are absent from the architecture contract's
[approved-dependency table](architecture.md#approved-dependencies).

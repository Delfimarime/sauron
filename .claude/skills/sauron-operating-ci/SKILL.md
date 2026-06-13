---
name: sauron-operating-ci
description: Use when writing or modifying CI/CD pipeline files for this repository (.github/workflows/ for GitHub Actions, .gitlab-ci.yml for GitLab CI) — the dependency-gated stages, parity with the Taskfile gate names, coverage/binary artifact hand-off, context-based version decoration, and publishing to Releases on the default branch. Normative rules live in spec/contracts/architecture.md.
---

# Operating Sauron's CI/CD

When writing or maintaining CI pipeline files, follow the
[Continuous integration & delivery](../../../spec/contracts/architecture.md#continuous-integration--delivery)
and [Versioning](../../../spec/contracts/architecture.md#versioning) sections of
the architecture contract. The pipeline is provider-agnostic; GitHub Actions and
GitLab CI are the reference targets.

## Procedural reminders

1. **Parity, not duplication.** Each CI job runs the **identically-named Taskfile
   target** (`test`, `gate-lint`, `build`, `gate-coverage`, `gate-security`,
   `gate-integration`). Never reimplement gate logic in the pipeline — call the task.
2. **Stage gating.**
   - Stage 1 (parallel): `test`, `gate-lint`.
   - Stage 2 (parallel, on stage 1 success): `build`, `gate-coverage`. `build`
     cross-compiles every target (`CGO_ENABLED=0`) — `linux/amd64`,
     `darwin/arm64`, `darwin/amd64` — in one Linux job; no per-OS runners.
   - Stage 3 (on stage 2 success): `gate-security`.
   - Stage 4 (on stage 3 success): `gate-integration` — alone, on a Linux runner
     (the suite needs a Docker daemon), against `dist/sauron-linux-amd64`.
   - Stage 5 (on stage 4 success, **default branch only**): `publish`.
3. **Artifact hand-off.** `test` publishes the coverage report consumed by
   `gate-coverage`; `build` publishes the per-OS binaries — `gate-security` and
   `gate-integration` consume `dist/sauron-linux-amd64`, and `publish` consumes
   all of them.
4. **Version decoration.** Pass the computed version to the build via the
   overridable `AppVersion`, by context: `main`/`master` →
   `<version>-RELEASE`; PR/MR into it → `<version>-PRE-RELEASE.<run-number>`;
   any other build → `<version>-SNAPSHOT.<run-number>`. `package.json` `version`
   stays the strict-SemVer source; CI only decorates it.
5. **Publish.** Only on the default branch: for every `dist/sauron-<os>-<arch>`
   binary, generate its SHA-256 checksum and publish the binary + `.sha256` as
   **Release assets** (GitHub Release / GitLab Release). No OCI artifacts, no
   package registry.
6. **Files only.** This is the operating surface for CI *files* under `.github/`
   and `.gitlab/` — no live-pipeline API access.

# Constitution

Project-governing principles for sauron. These constrain how features are
specified, planned, and implemented. Authoring mechanics live in
[spec/AUTHORING.md](spec/AUTHORING.md); the CLI conventions in the
[CLI contract](spec/contracts/cli.md).

> Status: initial draft — refine as implementation begins.

The principles are organized into four chapters that follow a feature's
lifecycle: how it is **specified**, the **contracts** it must honor, how it is
**implemented**, and how it stays **traceable**.

## Chapter I — Specification

How features are described, and the decisions behind them recorded, before any
code is written.

**Compliance with [spec/AUTHORING.md](spec/AUTHORING.md) is mandatory.** It is
the single normative source for spec structure, the
[feature/capability taxonomy](spec/AUTHORING.md#spec-types),
[EARS phrasing](spec/AUTHORING.md#ears-templates-normative), the
[glossary](spec/GLOSSARY.md),
[ADR format](spec/AUTHORING.md#adr-structure), and
[cross-reference style](spec/AUTHORING.md#cross-references). The articles below
state the principles; AUTHORING.md states how to satisfy them, and a spec that
violates it is not approvable.

### Article 1 — Spec-driven

Every feature begins as an approved spec under `spec/`. No implementation is
written without a spec, and each change traces back to the requirements (FR ids)
it realizes.

### Article 2 — EARS requirements

Requirements are expressed in EARS, following the
[EARS templates](spec/AUTHORING.md#ears-templates-normative) in AUTHORING.md.
They describe observable behavior, not implementation.

### Article 3 — Feature/capability separation

User-observable features are specified separately from the technical
capabilities that enable them, per the
[spec types](spec/AUTHORING.md#spec-types) in AUTHORING.md. A capability has no
CLI surface of its own.

### Article 4 — Decisions recorded as ADRs

Every significant technical decision — one not dictated by a requirement — is
captured as an Architecture Decision Record under the feature's `architecture/`
directory and linked from the spec's `## Decision Records` section; a decision
that is cross-cutting and owned by no single feature is recorded instead under
the project-level `spec/architecture/` directory. A decision that shapes
behavior or constrains implementation is written down before it is coded, never
made implicitly in source. An ADR is authored only with the maintainer's
explicit intent; it is never generated automatically. Authoring mechanics (the
`ADR-NNNN-<slug>.md` naming and layout) live in the
[ADR structure](spec/AUTHORING.md#adr-structure) of AUTHORING.md.

An ADR records a **decision, not a dependency**. Choosing a library or tool is
not an architectural decision — it is implementation detail, vetted under
Chapter III, Article 5 and enumerated in the architecture contract's
approved-dependency table, never in an ADR. ADRs capture the durable shape:
boundaries, contracts, conventions, and the trade-offs behind them.

### Article 5 — ADR structure and lifecycle

Each ADR carries a **Status**, **Date**, and **Feature** header (a project-level
ADR carries **Scope** in place of **Feature**), then states its **Context**, the
**Decision**, its **Consequences**, and a **Revisit when** condition naming what
would reopen it, as defined in the
[ADR structure](spec/AUTHORING.md#adr-structure) of AUTHORING.md. An accepted
ADR is not rewritten; a decision is changed by recording a new ADR that
supersedes the old one, so the history of why the system is shaped as it is
remains intact.

### Article 6 — No implementation with open questions

A spec is not ready for implementation while it carries an open or pending
question. Every ambiguity or deferred decision is clarified first — and where
the answer constrains the design, recorded as an ADR (Article 4) — so that
implementation proceeds only from a spec with nothing left unresolved. When a
question surfaces mid-implementation, work pauses until it is answered rather
than resolved by guesswork in code.

### Article 7 — Locality and single source

Each feature owns its own spec; a similar capability in another feature is
specified where it lives, not merged to avoid repetition. Semantics shared across
features — the glossary, the CLI / state / architecture / delivery contracts, the
canonical requirement boilerplate, and ADRs — are authored **once** in their
owning file and **linked**, never copied into a feature. There is one place to
change each rule, and therefore no drift between copies.

## Chapter II — Contracts

The command surface and the observable behavior every command must honor.

### Article 1 — Contract-first CLI

Every command's behavior is defined in its `contracts/<verb>-<noun>.md` and
conforms to the [CLI contract](spec/contracts/cli.md) — command grammar, shared
flags, exit-status semantics, and output discipline — before it is implemented. A
feature owns one or more such command contracts.

The sole exception is the **root command** (version and help plumbing): it has no
feature spec and no command contract, and its construction and
`--version` banner are fixed by the
[architecture contract](spec/contracts/architecture.md) instead.

### Article 2 — Safety and idempotency

Commands are idempotent where reasonable. Unregistering or deleting a source
preserves already-installed artifacts. Destructive operations support
`--dry-run`.

## Chapter III — Implementation

How the code that satisfies the contracts is written.

### Article 1 — Implementation standards

Implementation follows the project's Go conventions — uberfx wiring, cobra CLI,
the Uber Go Style Guide, structured zap + ECS logging, cognitive complexity ≤15,
and no rogue goroutines (all concurrency runs on a managed pool, never a bare
`go`) — and is test-first to a 90% coverage target. The full coding, telemetry,
and testing practices are fixed by the
[architecture contract](spec/contracts/architecture.md). The binary is built
CGO-free (`CGO_ENABLED=0`) so a single Go toolchain cross-compiles every
supported target without per-OS build hosts; the target matrix and build
mechanics are fixed by that contract.

### Article 2 — Standard project layout

Sauron is a Go 1.26 project following
[golang-standards/project-layout](https://github.com/golang-standards/project-layout),
with module path `github.com/delfimarime/sauron`. The directory structure,
package responsibilities, and build variables (`AppName`, `AppVersion`,
`AppHash`) are fixed by the
[architecture contract](spec/contracts/architecture.md). The external integration
test module under `/test` is governed by Articles 6–7 below.

### Article 3 — Ports and adapters

Public behavioral interfaces (ports) live under `pkg/sauron/extension`
(`Registry`, `Provider`), with shared data types in `pkg/sauron/types`, and are
implemented by adapters under `internal/infrastructure/repository/`
— the driven-adapter layer reaching external systems, grouped under one
`repository` module: `registry/{fs,git,http}` and `agent/{claude,zencoder}`. Each
adapter family exposes its wiring through an uberfx `NewFxOptions() fx.Option`. The
ports are a public surface: external code may implement new registries or providers
against them. The `repository` module also houses internal capabilities that
are not public extension points — `internal/infrastructure/repository/storage`,
which owns all manipulation of the `~/.sauron/` state — kept wholly inside their
package with no `pkg/` port. The transversal framework modules (`internal/config`,
`internal/telemetry`, `internal/cmd`) are not adapters and stay at the
`internal/` root. Exact paths are fixed by the
[architecture contract](spec/contracts/architecture.md).

### Article 4 — Use Case orchestration

A command's business logic is an interactor, not a service. Each command's
entrypoint is a **use case** that orchestrates the work; the reusable steps a use
case composes are **actions**, and the two share one shape. A use case is
stateless: its collaborators — the `pkg/` ports, the state stores, and the
logger — are supplied by uberfx, while the per-invocation `context.Context` and
the call's business input are passed to its `Execute`. A use case **returns a
presentation-agnostic result** — domain objects, not rendered output — and never
decides how that result is displayed: rendering is the client's responsibility,
performed by the command layer (`internal/cmd`) after `Execute` returns. The
use case therefore stays ignorant of presentation entirely, and of where state is
persisted — persisted state flows through the
`internal/infrastructure/repository/storage` package (which owns the `afero.Fs`),
never through direct OS calls or a hard-coded destination. The `UseCase`
interface shape, the `internal/usecase` layout, and the naming convention are
fixed by the [architecture contract](spec/contracts/architecture.md).

### Article 5 — Dependency, license & security discipline

The dependency set is deliberately small and vetted. Adding a dependency
requires scrutiny of its maturity, maintenance, license — which must be
permissive and compatible with the project's Apache-2.0 license — and security:
the dependency set is scanned for known vulnerabilities (Chapter IV, Article 2),
and a dependency carrying an unresolved advisory is not introduced. The approved
dependencies and their licenses are enumerated in the
[architecture contract](spec/contracts/architecture.md); nothing outside that
list is used without amending it. No dependency may require CGO; the binary stays
statically linked and cross-compilable (Article 1).

### Article 6 — Integration verification

The product is verified end-to-end by a black-box suite that drives the **built
binary** as an external operator would — process, arguments, exit code,
stdout/stderr, and the state files it persists — and asserts only through the
public `pkg/` surface, never `internal/` in-process. Go's `internal/` rule cannot
enforce this across the shared module prefix, so a `depguard` rule does. A
scenario may *arrange* by seeding a public, schema-valid state document only for a
command that merely reads it; a feature using that shortcut keeps at least one
black-box produce-then-read scenario so the write path stays exercised.

The suite lives as a **separate Go module** under `/test` (`test/e2e`) that
resolves the root through a `replace`; its heavy test dependencies stay there and
out of the approved-dependency table. It is **exempt from Articles 3–4** (no
ports-and-adapters, no Use Case orchestration): it consumes the binary rather than being
part of it. The module layout is fixed by the
[architecture contract](spec/contracts/architecture.md); the harness's own
architecture and conventions are governed locally by
[test/e2e/HARNESS.md](test/e2e/HARNESS.md).

### Article 7 — Integration test discipline

The suite is hermetic and authored spec-first. Each scenario's dependencies are
provisioned from ephemeral containers — no public-internet dependence in a
blocking gate — and `$SAURON_HOME` is pinned to a temporary directory so the real
`~/.sauron` is never touched and no process environment is mutated; it runs on
Linux. The harness is written test-first: it is correct when every step resolves
and the *only* failure is the not-yet-built command, and it turns green when the
production work lands with no harness change.

## Chapter IV — Governance & Traceability

The cross-cutting rules that keep spec, plan, and code in agreement, and that
gate what may ship.

### Article 1 — Traceability

Plans and implementation reference the spec and FRs they fulfill, so every unit
of behavior maps back to an approved requirement.

### Article 2 — Verification gate

A feature is not complete until it passes the project's verification gate:

- **Tests pass and coverage meets target.** The feature's tests — covering both
  success and failure paths — pass, and project coverage meets the target fixed
  by the [architecture contract](spec/contracts/architecture.md): 90% ideal,
  never below 80%.
- **Dependencies are scanned for vulnerabilities** with [Trivy](https://trivy.dev)
  (or an equivalent scanner) on every feature, against these per-scan thresholds:
  - **CRITICAL — none.** A feature does not ship with a CRITICAL finding unless a
    project-level ADR under `spec/architecture/` records a clear reason it is not
    fixed.
  - **HIGH — at most two.** More than two HIGH findings across the dependency set
    are not allowed unless a project-level ADR under `spec/architecture/` records
    the reason.
- **Integration tests pass.** The black-box BDD suite under `test/e2e` — driving
  the built binary end-to-end (Chapter III, Articles 6–7) — passes on Linux before
  a feature ships. Every transport — git and HTTP — is first-class: its scenarios
  run in the gate like any other.

Each such exception is an ADR that names the advisory and a **Revisit when**
condition, authored only with explicit user intent (Chapter I, Article 4) — never
generated automatically. A feature that fails either condition does not merge.
The gate is enforced by the project's Taskfile tasks (`task gate-lint`,
`task test`, `task gate-coverage`, `task gate-security`, `task gate-integration`),
run as the CI pipeline.

### Article 3 — Development workflow

Development follows a Trunk-based flow: `main`/`master` is the only long-lived
branch, and work lands through short-lived feature branches merged via
pull/merge request — never Gitflow. Commits and PR/MR titles follow
[Conventional Commits](https://www.conventionalcommits.org), and the
`package.json` SemVer `version` is bumped by hand to match the change type
(a feature → minor, a fix → patch, a breaking change → major). The workflow, the
feature-proposal (`PROPOSAL:` issues carrying EARS and intended ADRs) and
bug-report (Context / Problem / Expected Outcome) processes, and the issue and
PR/MR templates are defined in [CONTRIBUTING.md](CONTRIBUTING.md).

### Article 4 — Single-source documentation

Every fact is documented in exactly one place and linked from elsewhere. The
product and domain model are authored once in [the spec README](spec/README.md);
the repository [README](README.md) and other documents link to it rather than
restating it. A normative rule lives in its owning contract; narrative docs point
to that contract and never paraphrase it. Documentation therefore cannot drift
out of agreement with what it describes.

### Article 5 — Distill

Delivery and use feed back into the spec; the lifecycle is a loop, not a line. A
command leaves an observable record of what it did — its terminal outcome is
emitted through the structured zap + ECS logger (`event.action`, the outcome, and
on failure the error's type and reason) — so the system's behavior can be
inspected, not guessed. An insight that delivery or use reveals — a spec that
mispredicted behavior, an edge case the requirements missed, a usage pattern that
contradicts the design — is recorded in the owning feature's `plan.md` `## Distill`
section and closed out either as a spec amendment carrying its `FR-NNN` trace or as
a `PROPOSAL:` issue (Article 3) that reopens discovery. An insight is never left
only in someone's head, and an insight that changes a requirement is reflected in
the spec before the code.

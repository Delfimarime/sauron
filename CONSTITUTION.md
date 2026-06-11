# Constitution

Project-governing principles for sauron. These constrain how features are
specified, planned, and implemented. Authoring mechanics and CLI conventions
live in [spec/AUTHORING.md](spec/AUTHORING.md); the compiled command reference
in [spec/contracts/cli.md](spec/contracts/cli.md).

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
[glossary](spec/AUTHORING.md#glossary),
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
directory and linked from the spec's `## Decision Records` section. A decision
that shapes behavior or constrains implementation is written down before it is
coded, never made implicitly in source. Authoring mechanics (the
`ADR-NNNN-<slug>.md` naming and layout) live in the
[ADR structure](spec/AUTHORING.md#adr-structure) of AUTHORING.md.

### Article 5 — ADR structure and lifecycle

Each ADR carries a **Status**, **Date**, and **Feature** header, then states its
**Context**, the **Decision**, its **Consequences**, and a **Revisit when**
condition naming what would reopen it, as defined in the
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

## Chapter II — Contracts

The command surface and the observable behavior every command must honor.

### Article 1 — Contract-first CLI

Every command's behavior is defined in its `contracts/command-line.md` and
conforms to the [CLI conventions](spec/AUTHORING.md#cli-conventions) — command
grammar, shared flags, exit-status semantics, and output discipline — before it
is implemented. The compiled [command reference](spec/contracts/cli.md)
summarizes every command in one place.

### Article 2 — Safety and idempotency

Commands are idempotent where reasonable. Unregistering or deleting a source
preserves already-installed artifacts. Destructive operations support
`--dry-run`.

## Chapter III — Implementation

How the code that satisfies the contracts is written.

### Article 1 — Implementation standards

Implementation follows the project's Go conventions (uberfx architecture, cobra
CLI, Uber style, cognitive complexity ≤15) and is test-first with a 90%
coverage target.

### Article 2 — Standard project layout

Sauron is a Go 1.26 project following
[golang-standards/project-layout](https://github.com/golang-standards/project-layout),
with module path `github.com/delfimarime/sauron`. The directory structure,
package responsibilities, and build variables (`AppName`, `AppVersion`,
`AppHash`) are fixed by the
[architecture contract](spec/contracts/architecture.md).

### Article 3 — Ports and adapters

Public behavioral interfaces live under `pkg/` (`pkg/repository`,
`pkg/provider`) and are implemented by adapters under `internal/`
(`internal/repository/{fs,git,http}`, `internal/provider/{claude,zencoder}`).
Each adapter family exposes its wiring through an uberfx
`NewFxOptions() fx.Option`. The interfaces are a public surface: external code
may implement new repositories or providers against them.

### Article 4 — Dependency & license discipline

The dependency set is deliberately small and vetted. Adding a dependency
requires scrutiny of its maturity, maintenance, and license — which must be
permissive and compatible with the project's Apache-2.0 license. The approved
dependencies and their licenses are enumerated in the
[architecture contract](spec/contracts/architecture.md); nothing outside that
list is used without amending it.

## Chapter IV — Governance & Traceability

The cross-cutting rule that keeps spec, plan, and code in agreement.

### Article 1 — Traceability

Plans and implementation reference the spec and FRs they fulfill, so every unit
of behavior maps back to an approved requirement.

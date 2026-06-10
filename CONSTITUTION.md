# Constitution

Project-governing principles for sauron. These constrain how features are
specified, planned, and implemented. Authoring mechanics and CLI conventions
live in [spec/AUTHORING.md](spec/AUTHORING.md); the compiled command reference
in [spec/contracts/cli.md](spec/contracts/cli.md).

> Status: initial draft — refine as implementation begins.

## Article I — Spec-driven

Every feature begins as an approved spec under `spec/`. No implementation is
written without a spec, and each change traces back to the requirements (FR ids)
it realizes.

## Article II — EARS requirements

Requirements are expressed in EARS, following
[spec/AUTHORING.md](spec/AUTHORING.md). They describe observable behavior, not
implementation.

## Article III — Feature/capability separation

User-observable features are specified separately from the technical
capabilities that enable them. A capability has no CLI surface of its own.

## Article IV — Contract-first CLI

Every command's behavior is defined in its `contracts/command-line.md` and
conforms to the [CLI conventions](spec/AUTHORING.md#cli-conventions) — command
grammar, shared flags, exit-status semantics, and output discipline — before it
is implemented. The compiled [command reference](spec/contracts/cli.md)
summarizes every command in one place.

## Article V — Implementation standards

Implementation follows the project's Go conventions (uberfx architecture, cobra
CLI, Uber style, cognitive complexity ≤15) and is test-first with a 90%
coverage target.

## Article VI — Safety and idempotency

Commands are idempotent where reasonable. Unregistering or deleting a source
preserves already-installed artifacts. Destructive operations support
`--dry-run`.

## Article VII — Traceability

Plans and implementation reference the spec and FRs they fulfill, so every unit
of behavior maps back to an approved requirement.

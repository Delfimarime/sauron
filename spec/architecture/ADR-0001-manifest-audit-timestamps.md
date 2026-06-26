# ADR-0001: Audit timestamps live on the shared manifest metadata envelope

**Status**: Accepted
**Date**: 2026-06-21
**Scope**: Project-wide

## Context

Every persisted manifest document — Registry, Skill, Agent, Persona, Schedule —
shares one `metadata` envelope. Until now that envelope recorded no temporal
audit information: a document carried no record of when it was first written or
when it last changed.

This information is wanted in two places. Read-side surfaces (describe and list)
want to show when a source or artifact was created and last modified. Audit and
troubleshooting want the same facts to reconstruct what changed and when.

Because every kind already shares the `metadata` envelope, the audit facts
belong there once, uniformly, rather than each kind inventing and re-deriving its
own scheme. A per-kind approach would drift: different field names, different
formats, different stamping rules across kinds that are otherwise identical in
their envelope.

## Decision

The shared `metadata` block gains two audit attributes:

- `metadata.createdAt` — when the document was first written.
- `metadata.lastUpdatedAt` — when the document last changed.

The decision applies to the **shared envelope**, not to any individual kind, so
all manifest kinds carry the fields identically.

Rules:

- **Format** — both values are RFC3339 timestamps expressed in UTC.
- **Who stamps** — the values are stamped by the writing use case, not by the
  persistence layer, and are never hand-edited in source documents.
- **Determinism** — the writer obtains the current instant from an injected
  clock, so writes are deterministic and assertable under test rather than
  reading wall-clock time directly.
- **Create vs. update** — on create both timestamps are set to the same instant.
  A subsequent update advances `lastUpdatedAt` only and leaves
  `createdAt` untouched.
- **Compatibility** — both fields are optional in the schema. Documents written
  before this decision lack the fields, and readers tolerate their absence, so
  the change is backward compatible.

## Consequences

**Positive**

- All manifest kinds gain audit timestamps consistently, defined once on the
  shared envelope with no per-kind divergence.
- Describe and list can surface creation and last-change times for any kind.
- Stamping in the writing use case against an injected clock keeps writes
  deterministic and testable.
- Reading older documents that lack the fields remains valid; no migration is
  forced.
- No secret or personally identifiable information is introduced — only UTC
  instants.

**Negative**

- Every writing use case must stamp the timestamps; a writer that forgets leaves
  the fields absent.
- The manifest schema and the state contract must be updated to define the two
  optional fields and the stamping rules.
- Testing writers now requires an injectable clock rather than reading the
  system time directly.

## Revisit when

Audit needs grow beyond two instants — for example, a need to record *who*
changed a document, a change history, or per-field provenance — such that a
single creation/last-update pair on the envelope is no longer sufficient.

# ADR-0001: Persona priority is zero-anchored, optional, and undefined ranks last

**Status**: Accepted

**Date**: 2026-06-10

**Feature**: Import Persona

## Context

Personas need a precedence order — listing wants a deterministic sort, and sync needs a deterministic way to attribute artifacts when several personas bring the same one. Repositories solve ordering with a required, always-unique positive priority, but forcing the same on personas would make every import pick a free number even when ordering does not matter to the user yet.

## Decision

Persona priority is **zero-anchored and optional**:

- The **first persona** imported is assigned priority `0`; the user cannot choose it (`--priority` on the first import is a usage error).
- **Subsequent imports** may pass `--priority <n>` with an integer ≥ 1 that is not in use, or omit it entirely — an omitted priority is stored as **undefined**.
- Uniqueness applies to **defined values only**; any number of personas may have undefined priority at the same time.
- **Ordering**: defined priorities rank first, ascending (`0` is the highest precedence); personas with undefined priority rank after all defined ones, ordered by name among themselves.
- Priorities are adjusted later with [set priority persona](../../0010-set-persona-priority/spec.md), which is blocked while only one persona exists — that persona keeps `0`. `0` can be reassigned via `set priority` only when no persona holds it (e.g. after the original first persona was deleted); there is no invariant that some persona must hold `0`.

## Consequences

**Positive**

- Low-friction import: ordering is opt-in, and the common single-persona setup needs no priority decisions at all.
- Deterministic order everywhere (list, sync) even when priorities are left undefined.
- The first persona is the natural default and highest precedence.

**Negative**

- Two-tier semantics (defined vs undefined) must be implemented consistently by every consumer of the order.
- A persona's effective rank can shift implicitly as other personas gain or lose defined priorities.

## Revisit when

Ordering needs become richer (weights, groups, per-target precedence) or the zero-anchor proves limiting.

# ADR-0002: Priority is optional and unified across repositories and personas

**Status**: Accepted

**Date**: 2026-06-12

**Feature**: Import Persona

## Context

Sauron orders both repositories and personas by an integer priority, where a
lower value wins. Two divergent models had grown up: repositories required a
unique **positive** priority on every [add repository](../../0001-add-repository/spec.md),
while personas were zero-anchored and optional with an **undefined** tier that
ranked last (the superseded
[ADR-0001](ADR-0001-persona-priority-model.md)). Maintaining two models forces
every consumer of the order — list, sync, conflict resolution — to special-case
which kind it is reading, and forces repository registration to invent a free
number even when ordering does not yet matter. One model for both is simpler to
specify, implement, and explain.

## Decision

Repositories and personas share **one** priority model. `--priority` is
**optional** on registration ([add repository](../../0001-add-repository/spec.md),
[import persona](../spec.md)):

- **First resource of its kind** (registry empty): priority is `0`. Omitting
  `--priority` defaults it to `0`; passing `--priority` is accepted only when it
  is `0`, and any other value is a usage error.
- **Subsequent resource** (at least one already registered):
  - Omitting `--priority` assigns the value at the end of the priority-ordered
    list — one greater than the current highest priority (`max + 1`) — which
    never collides.
  - Passing `--priority <n>` uses that value, rejected when another resource of
    the same kind already holds it.
- Priority is **always defined** — the undefined tier is removed — and
  **unique** within its kind. Lower value wins; `0` is the highest precedence.
- Priorities change after registration only through
  [set priority repository](../../0011-set-repository-priority/spec.md) and
  [set priority persona](../../0010-set-persona-priority/spec.md), which is
  **blocked while a single resource of that kind exists** — that lone resource
  keeps `0`.

This supersedes both the zero-anchored, undefined-tolerant persona model in
[ADR-0001](ADR-0001-persona-priority-model.md) and the always-required positive
repository priority, replacing them with the rules above.

## Consequences

**Positive**

- One model for both kinds: list ordering, sync attribution, and conflict
  resolution never special-case an undefined or missing value.
- Low-friction registration: the common case needs no `--priority` decision —
  the first resource is `0`, every later one appends at the end.
- Deterministic, collision-free defaults; every resource carries a concrete,
  comparable priority.

**Negative**

- Append-at-end defaults (`max + 1`) leave gaps after deletions or custom
  values; priorities are not guaranteed contiguous.
- Reusing `0` requires the current zero-holder to be removed first (or
  priorities re-set), since the single-resource guard fixes a lone resource at
  `0`.
- Repository registration loses the guarantee that a priority is explicitly
  chosen at add time.

## Revisit when

Ordering needs become richer (weights, groups, per-target precedence), or
contiguous/compacting priorities become a requirement.

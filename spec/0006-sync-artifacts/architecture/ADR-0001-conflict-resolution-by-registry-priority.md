# ADR-0001: Conflict resolution — pin, then registry priority

**Status**: Accepted

**Date**: 2026-06-12

**Feature**: Sync Artifacts

## Context

More than one registered registry can host an artifact with the same name — for
example, two registries each providing a `code-review` skill. Sync artifacts must
deliver exactly one of them, deterministically, or the environment would depend on
iteration order.

Every registry already carries a priority that is unique across all registries
regardless of kind ([add registry](../../0001-add-registry/spec.md)), with a lower
value meaning higher precedence. That priority is the right *default* tiebreaker.
But a team sometimes needs a specific artifact to come from a specific registry
regardless of priority — for example, to shadow an upstream `code-review` skill
with an internal one hosted on a lower-precedence registry. A pin
([pin artifact](../../0020-pin-artifact/spec.md)) lets a user declare that binding,
and sync must honour it above priority.

## Decision

When the desired set contains an artifact name that more than one registry
provides, sync artifacts resolves it in two layers:

1. **Pin** — if the artifact's track entry has `pinned: true`, its recorded
   `registry` (the user's pin) is the source; priority is not consulted.
2. **Priority** — otherwise, the **registry with the lowest priority value**
   (highest precedence) wins, regardless of kind.

The chosen registry is recorded as the artifact's provenance in
`~/.sauron/track.yaml`, and `pinned` records whether that winner was a user pin
rather than the priority-resolved default — so the track file alone shows which
artifacts are pinned. A pin lives on the artifact's track entry, not in a separate
file; pinning records the binding ahead of the next sync, which reconciles the
on-disk artifact to the pinned registry (immediately when the caller passes
`--reconcile`).

Persona priority plays a different role: it does not select artifact contents (two
personas naming the same artifact get the same resolved artifact); it only
determines which persona is recorded as provenance when several bring the same
artifact into the desired set ([priority model](../../AUTHORING.md#priority-model)).

## Consequences

**Positive**

- Registry priority is operational and syncs are deterministic and reproducible.
- Teams can deliberately shadow an upstream artifact — by priority (a
  higher-precedence registry) or, deterministically, by pinning a specific one.
- The pin is visible in the one place that already records provenance.

**Negative**

- A lower-precedence or unpinned same-named artifact is silently never delivered;
  name hygiene across registries is the user's responsibility.
- A pinned name ignores higher-precedence registries until unpinned, and a
  record-only pin leaves a transient gap between the recorded `registry` and the
  on-disk source until the next sync.

## Revisit when

Pins need to target something finer than a registry (e.g. a version or ref), or a
pin should apply across providers as a single declaration.

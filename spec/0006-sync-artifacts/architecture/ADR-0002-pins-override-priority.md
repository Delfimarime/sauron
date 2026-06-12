# ADR-0002: A pinned artifact overrides priority-based conflict resolution

**Status**: Accepted

**Date**: 2026-06-12

**Feature**: Sync Artifacts

## Context

[ADR-0001](ADR-0001-conflict-resolution-by-registry-priority.md) resolves a
same-named artifact offered by several registries to the registry with the lowest
priority value. That is the right *default*, but a team sometimes needs a specific
artifact to come from a specific registry regardless of priority — for example,
to shadow an upstream `code-review` skill with an internal one hosted on a
lower-precedence registry. [Pin artifact](../../0020-pin-artifact/spec.md) lets a
user declare that binding; sync must honour it.

## Decision

Conflict resolution gains a layer **above** priority:

1. **Pin** — if the artifact's track entry has `pinned: true`, its recorded
   `registry` (the user's pin) is the source; priority is not consulted.
2. **Priority** — otherwise, the lowest-priority-value registry wins, per
   ADR-0001.

The pin lives on the artifact's `track.yaml` entry (`pinned` boolean), not in a
separate file, so the track file alone shows which artifacts are pinned and which
are priority-resolved. Pinning records the binding ahead of the next sync; sync
reconciles the on-disk artifact to the pinned registry (immediately when the
caller passes `--reconcile`). This extends ADR-0001; it does not supersede it —
priority still governs every unpinned artifact.

## Consequences

**Positive**

- Teams can deterministically source an artifact from a chosen registry.
- The pin is visible in the one place that already records provenance.

**Negative**

- A pinned name silently ignores higher-precedence registries until unpinned.
- A record-only pin leaves a transient gap between the recorded `registry` and
  the on-disk source until the next sync.

## Revisit when

Pins need to target something finer than a registry (e.g. a version or ref), or a
pin should apply across providers as a single declaration.

# Pin Artifact

**Type:** feature

**Depends on:** [sync artifacts](../0006-sync-artifacts/spec.md), [add registry](../0001-add-registry/spec.md)

## Overview

A person responsible for a team's agentic-AI setup needs a specific skill or agent
to come from a specific registry regardless of registry priority — for example to
shadow an upstream artifact with an internal one — so that conflict resolution
follows an explicit choice rather than priority alone. Pinning records the binding
on the artifact's track entry; [sync artifacts](../0006-sync-artifacts/spec.md)
then honours it (see
[ADR-0001](../0006-sync-artifacts/architecture/ADR-0001-conflict-resolution-by-registry-priority.md)).

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to pin a skill or agent to a
  registry and to unpin it, controlling which registry sources that artifact.

### Event-driven

- **FR-002**: When a user runs `pin skill <name> <registry>` or `pin agent
  <name> <registry>`, Sauron shall record the artifact as pinned to that registry
  on its track entry, setting the entry's `registry` and `pinned` to `true`.
- **FR-003**: When a user runs `unpin skill <name>` or `unpin agent <name>`,
  Sauron shall clear the pin on that artifact's track entry, so the next sync
  re-resolves its registry by priority.
- **FR-004**: When the request would not change the track entry — the artifact is
  already pinned to that registry, or already unpinned — Sauron shall make no
  change and report success.
- **FR-005**: When a pin or unpin completes, Sauron shall report the artifact and
  its source registry.

### State-driven

- **FR-006**: While recording a pin or unpin, Sauron shall change only the named
  artifact's track entry, leaving every other entry untouched.

### Unwanted behavior

- **FR-007**: If required arguments are missing — a name, or a registry for
  `pin` — then Sauron shall reject the request and report what is required.
- **FR-008**: If the named registry is not registered, then Sauron shall reject
  the request and report that the registry is unknown.
- **FR-009**: If the named registry does not offer the named artifact, then
  Sauron shall reject the request and report that the artifact is unavailable
  there.
- **FR-010**: If the artifact is not installed and `--reconcile` is not given,
  then Sauron shall reject the request and report that the artifact is not
  installed and that `--reconcile` would install and pin it.
- **FR-011**: If the track file cannot be read or written, then Sauron shall
  reject the request and report that it cannot be updated.

### Optional

- **FR-012**: Where `--reconcile` is provided, Sauron shall apply the change
  immediately by reconciling the affected artifact from its source registry,
  installing it when the pin targets a not-yet-installed artifact.
- **FR-013**: Where `--dry-run` is provided to `unpin`, Sauron shall print the
  current pinned registry and the registry priority would pick after unpinning,
  without changing the track file or the environment.

## Key Entities

- **Pin**: a binding of an artifact (`skill` or `agent`, by name) to a registry,
  recorded as `pinned: true` on the artifact's
  [track file](../0006-sync-artifacts/spec.md) entry; it overrides priority during
  [sync artifacts](../0006-sync-artifacts/spec.md).

## Decision Records

- Pin precedence and its on-the-track-entry storage are decided with sync
  artifacts in
  [ADR-0001](../0006-sync-artifacts/architecture/ADR-0001-conflict-resolution-by-registry-priority.md).

## Notes

- Pinning records the binding ahead of the next sync; without `--reconcile` the
  on-disk artifact is reconciled to the pinned registry on the next
  [sync artifacts](../0006-sync-artifacts/spec.md). The pin lives on the track
  entry — there is no separate pins file.

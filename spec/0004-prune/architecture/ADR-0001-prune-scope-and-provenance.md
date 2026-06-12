# ADR-0001: Prune scope is unregistered-registry orphans, by recorded provenance

**Status**: Accepted

**Date**: 2026-06-09

**Feature**: Prune Orphaned Skills and Agents

## Context

Deleting a registry unregisters the source but keeps its installed artifacts ([delete registry ADR-0001](../../0003-delete-registry/architecture/ADR-0001-unregister-keeps-installed-artifacts.md)), which can leave orphans on the team's providers. "Stale artifact" could mean two things: (a) an artifact whose source registry is no longer registered, or (b) an artifact that drifted from a still-registered registry (renamed or removed upstream). Identifying either requires knowing where each installed artifact came from.

## Decision

Prune removes only category (a): installed skills/agents whose **source registry is not among the currently registered registries**. Drift within a still-registered registry is out of scope and belongs to a future sync operation.

Provenance comes from `~/.sauron/track.yaml` — the installed-artifact record that associates each delivered artifact with its source registry. That file is produced and maintained by [sync artifacts](../../0006-sync-artifacts/spec.md); prune reads it and removes entries for the artifacts it prunes.

## Consequences

**Positive**

- Clear, predictable scope that matches "registries no longer part of Sauron".
- No guessing: an artifact is pruned only when its recorded source is gone.

**Negative**

- Artifacts with no entry in `track.yaml` cannot be pruned (they are not Sauron's to reason about).
- Upstream drift within a registered registry is not addressed here.

## Revisit when

A sync/reconcile operation is introduced to handle in-registry drift, or when artifacts need to be attributed to more than one source registry.

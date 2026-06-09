# ADR-0001: Deleting a repository unregisters the source only

**Status**: Accepted

**Date**: 2026-06-09

**Feature**: Delete Repository

## Context

Deleting a repository could mean one of two things: unregister the source so Sauron stops watching it, or additionally cascade-remove every skill and agent that source delivered to the team's targets. A cascade removal is destructive — installed artifacts may be in active use, may have been delivered by more than one source, or may have been modified locally.

## Decision

Deletion **unregisters the source only**. The repository entry is removed from `~/.sauron/settings.json`; skills and agents already installed from it are left untouched (FR-003). Cleaning up delivered artifacts is a separate concern — a future uninstall/sync operation — not part of deleting the source.

## Consequences

**Positive**

- Deletion is safe and non-destructive to running setups; predictable and simple.
- A repository can be removed and re-added without disturbing installed artifacts.

**Negative**

- Artifacts from a deleted repository remain until cleaned up by other means; orphaned artifacts are possible.

## Revisit when

A cascade-uninstall becomes a requirement. A future ADR would then define how delivered artifacts are tracked and removed.

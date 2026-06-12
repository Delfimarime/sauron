# ADR-0001: Deleting a registry unregisters the source only

**Status**: Accepted

**Date**: 2026-06-09

**Feature**: Delete Registry

## Context

Deleting a registry could mean one of two things: unregister the source so Sauron stops watching it, or additionally cascade-remove every skill and agent that source delivered to the team's providers. A cascade removal is destructive — installed artifacts may be in active use, may have been delivered by more than one source, or may have been modified locally.

## Decision

Deletion **unregisters the source only**. The registry entry is removed from `~/.sauron/settings.yaml`; skills and agents already installed from it are left untouched (FR-003). Cleaning up delivered artifacts is a separate concern — a future uninstall/sync operation — not part of deleting the source.

## Consequences

**Positive**

- Deletion is safe and non-destructive to running setups; predictable and simple.
- A registry can be removed and re-added without disturbing installed artifacts.

**Negative**

- Artifacts from a deleted registry remain until cleaned up by other means; orphaned artifacts are possible.

## Revisit when

A cascade-uninstall becomes a requirement. A future ADR would then define how delivered artifacts are tracked and removed.

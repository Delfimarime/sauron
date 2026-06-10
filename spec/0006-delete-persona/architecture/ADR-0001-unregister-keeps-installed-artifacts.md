# ADR-0001: Deleting a persona removes the definition only

**Status**: Accepted

**Date**: 2026-06-10

**Feature**: Delete Persona

## Context

Deleting a persona could mean one of two things: remove the definition so it stops shaping deliveries, or additionally cascade-remove every skill and agent the persona brought to the team's targets. A cascade removal is destructive — installed artifacts may be in active use, may also be claimed by other personas, or may have been modified locally. The same reasoning led [delete repository](../../0003-delete-repository/spec.md) to keep installed artifacts when a repository is deleted.

## Decision

Deletion **removes the persona definition only**. The entry is removed from `personas[]` in `~/.sauron/settings.yaml`; skills and agents already installed are left untouched (FR-003). Reconciling the environment with the new set of personas is the job of [sync](../../0009-sync/spec.md): its next run computes the new desired set and removes out-of-scope artifacts deliberately, as visible `-` lines in its plan.

## Consequences

**Positive**

- Deletion is safe and non-destructive to running setups; predictable and simple.
- A persona can be removed and re-imported without disturbing installed artifacts.
- Cleanup happens in one well-defined place (sync), where it is previewable with `--dry-run`.

**Negative**

- Between the deletion and the next sync, installed artifacts may no longer be claimed by any persona.

## Revisit when

A cascade-uninstall on persona deletion becomes a requirement.

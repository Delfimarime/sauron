# ADR-0001: Same-named artifacts resolve by repository priority

**Status**: Accepted

**Date**: 2026-06-10

**Feature**: Sync

## Context

More than one registered repository can host an artifact with the same name — for example, two repositories each providing a `code-review` skill. Sync must deliver exactly one of them, deterministically, or the environment would depend on iteration order.

Every repository already carries a priority that is unique across all repositories regardless of kind ([add repository](../../0001-add-repository/spec.md)), with a lower value meaning higher precedence — but until now that priority had no operational meaning.

## Decision

When the desired set contains an artifact name that more than one repository provides, sync takes it from the **repository with the lowest priority value** (highest precedence). The same rule applies regardless of repository kind. The chosen repository is recorded as the artifact's provenance in `~/.sauron/track.yaml`.

Persona priority plays a different role: it does not select artifact contents (two personas naming the same artifact get the same resolved artifact); it only determines which persona is recorded as provenance when several bring the same artifact into the desired set ([import persona ADR-0002](../../0005-import-persona/architecture/ADR-0002-unified-priority-model.md)).

## Consequences

**Positive**

- Repository priority becomes operational and syncs are deterministic and reproducible.
- Teams can deliberately shadow an upstream artifact by hosting a same-named one in a higher-precedence repository.

**Negative**

- A lower-precedence repository's same-named artifact is silently never delivered; name hygiene across repositories is the user's responsibility.

## Revisit when

Per-artifact or per-persona source pinning becomes a requirement.

# ADR-0001: Same-named artifacts resolve by registry priority

**Status**: Accepted

**Date**: 2026-06-10

**Feature**: Sync

## Context

More than one registered registry can host an artifact with the same name — for example, two registries each providing a `code-review` skill. Sync artifacts must deliver exactly one of them, deterministically, or the environment would depend on iteration order.

Every registry already carries a priority that is unique across all registries regardless of kind ([add registry](../../0001-add-registry/spec.md)), with a lower value meaning higher precedence — but until now that priority had no operational meaning.

## Decision

When the desired set contains an artifact name that more than one registry provides, sync artifacts takes it from the **registry with the lowest priority value** (highest precedence). The same rule applies regardless of registry kind. The chosen registry is recorded as the artifact's provenance in `~/.sauron/track.yaml`.

Persona priority plays a different role: it does not select artifact contents (two personas naming the same artifact get the same resolved artifact); it only determines which persona is recorded as provenance when several bring the same artifact into the desired set ([priority model](../../AUTHORING.md#priority-model)).

## Consequences

**Positive**

- Registry priority becomes operational and syncs are deterministic and reproducible.
- Teams can deliberately shadow an upstream artifact by hosting a same-named one in a higher-precedence registry.

**Negative**

- A lower-precedence registry's same-named artifact is silently never delivered; name hygiene across registries is the user's responsibility.

## Revisit when

Per-artifact or per-persona source pinning becomes a requirement.

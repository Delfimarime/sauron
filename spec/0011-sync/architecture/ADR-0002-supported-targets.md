# ADR-0002: Supported targets are claude and zencoder, defaulting to zencoder

**Status**: Accepted

**Date**: 2026-06-10

**Feature**: Sync

## Context

The Target concept (`spec/README.md`) is the provider destination where artifacts are persisted — each provider stores skills and agents differently. There is no Target registration feature yet, but sync must deliver somewhere concrete to be implementable.

## Decision

Sync takes a `--target` flag whose supported values are **`claude`** and **`zencoder`**, defaulting to **`zencoder`** when omitted. Each target maps artifact types to that provider's conventional locations in the current environment (for example, Claude's `~/.claude/skills/` and `~/.claude/agents/`). The delivered provider is recorded per artifact in `~/.sauron/track.yaml` (`target` field), so one environment can hold synced artifacts for multiple providers side by side.

`codex` is the first named candidate for a future addition.

## Consequences

**Positive**

- Sync is implementable now without waiting for a Target registration feature.
- Multiple providers can be synced independently in the same environment, tracked separately.

**Negative**

- Supporting a new provider requires a spec change (extend the enum and its location mapping) rather than configuration.
- Provider location conventions are baked into Sauron.

## Revisit when

A Target registration feature is specced (making targets configurable), or `codex` (or another provider) needs support.

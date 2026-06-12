# ADR-0001: The provider is a single global setting, defaulting to claude

**Status**: Accepted

**Date**: 2026-06-10

**Feature**: Set Provider

## Context

The Provider concept (`spec/README.md`) is the provider destination where artifacts are persisted — each provider stores skills and agents differently. An earlier draft of sync took a per-invocation `--provider` flag, but that left the "current" provider ambiguous across commands (sync, prune, delete artifacts, cron) and forced every caller to repeat it. A team realistically delivers to one provider at a time.

There is no Provider registration feature yet, but Sauron must deliver somewhere concrete to be implementable.

## Decision

The provider is a **single global setting** stored in `~/.sauron/settings.yaml` (`provider:`), with supported values **`claude`** and **`zencoder`** and a default of **`claude`** when never set. No command takes a `--provider` flag; sync and cron read this setting.

Changing the provider with `sauron set provider <value>` **migrates installed artifacts**: each artifact recorded on the previous provider is installed at the new provider's locations and, by default, removed from the previous one (a move). With `--copy-only`, the previous provider's artifacts are kept and the new provider's copies are tracked as additional entries, so artifacts can exist on more than one provider. Each artifact's provider is recorded per entry in `~/.sauron/track.yaml` (`provider` field).

`codex` is the first named candidate for a future addition.

## Consequences

**Positive**

- One unambiguous active provider; sync, prune, delete artifacts, and cron all agree without repeating a flag.
- Switching providers is a single, explicit operation with a clear migration story.
- `--copy-only` supports running two providers side by side during a transition.

**Negative**

- Supporting a new provider requires a spec change (extend the enum and its location mapping) rather than configuration.
- Provider location conventions are baked into Sauron.

## Revisit when

A Provider registration feature is specced (making providers configurable and possibly plural/active-set), or `codex` (or another provider) needs support.

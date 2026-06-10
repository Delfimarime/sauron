# ADR-0001: The target is a single global setting, defaulting to claude

**Status**: Accepted

**Date**: 2026-06-10

**Feature**: Set Target

## Context

The Target concept (`spec/README.md`) is the provider destination where artifacts are persisted — each provider stores skills and agents differently. An earlier draft of sync took a per-invocation `--target` flag, but that left the "current" target ambiguous across commands (sync, prune, clear, cron) and forced every caller to repeat it. A team realistically delivers to one provider at a time.

There is no Target registration feature yet, but Sauron must deliver somewhere concrete to be implementable.

## Decision

The target is a **single global setting** stored in `~/.sauron/settings.yaml` (`target:`), with supported values **`claude`** and **`zencoder`** and a default of **`claude`** when never set. No command takes a `--target` flag; sync and cron read this setting.

Changing the target with `sauron set target <value>` **migrates installed artifacts**: each artifact recorded on the previous target is installed at the new target's locations and, by default, removed from the previous one (a move). With `--copy-only`, the previous target's artifacts are kept and the new target's copies are tracked as additional entries, so artifacts can exist on more than one provider. Each artifact's provider is recorded per entry in `~/.sauron/track.yaml` (`target` field).

`codex` is the first named candidate for a future addition.

## Consequences

**Positive**

- One unambiguous active target; sync, prune, clear, and cron all agree without repeating a flag.
- Switching providers is a single, explicit operation with a clear migration story.
- `--copy-only` supports running two providers side by side during a transition.

**Negative**

- Supporting a new provider requires a spec change (extend the enum and its location mapping) rather than configuration.
- Provider location conventions are baked into Sauron.

## Revisit when

A Target registration feature is specced (making targets configurable and possibly plural/active-set), or `codex` (or another provider) needs support.

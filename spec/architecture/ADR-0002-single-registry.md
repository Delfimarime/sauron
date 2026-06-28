# ADR-0002: v1 supports exactly one registry; multi-registry is deferred

**Status**: Accepted
**Date**: 2026-06-25
**Scope**: Project-wide

## Context

Sauron installs skills, agents, and personas from a registry — a source of
artifacts reached over a transport (`git` or `http`). An early
design treated the registry as a **named, plural** concept: a developer could
register many sources side by side and Sauron would track, per artifact, which
source delivered it.

That plural shape threaded through the whole system:

- **State.** `registries.yaml` held a collection of `Registry` documents, each
  identified by a unique, path-safe `metadata.name`. The name was a namespacing
  segment (`sauron-<registry>-<name>`) and the value every tracked artifact
  referenced.
- **Provenance.** In `track.yaml`, the unique key for an artifact was the triple
  `(kind, registry, name)` — the same name could appear under different
  registries — and each tracked document carried `spec.registry` recording its
  source.
- **Commands.** `add registry <name>` registered a source; `list registries`
  reviewed them as a table (columns `NAME`, `TRANSPORT`, `URI`); `delete
  registry <name>` unregistered one and cascade-uninstalled every artifact whose
  `spec.registry` matched. `install`, `uninstall`, and the catalogue commands
  carried a `<registry>` positional argument to scope an operation to one named
  source, and list/describe output reserved a `REGISTRY` column.

In practice, the developer Sauron targets points at a single source — an
organization's registry — and works entirely within it. The plural machinery
imposed a constant cost for a capability no current user exercises: a name to
invent and repeat on every command, a `REGISTRY` column on every listing, a
`<registry>` argument threaded through install/uninstall/catalogue, and a
source-tagged identity on every tracked artifact. Carrying that surface into
v1.0.0 would lock the larger grammar in before there is a concrete need for more
than one source.

## Decision

Sauron v1.0.0 supports **exactly one registry**.

- **A single global setting.** The registry is one global value persisted in
  `settings.yaml`, alongside the provider — not a collection in its own file.
  There is no `registries.yaml` and no named `Registry` collection in v1.
- **Configured as a setting, not a roster.** The registry is set with `set
  registry` and cleared with `unset registry`, and its current value is shown by
  `describe registry`. There is no `add registry`, no `list registries`, and no
  `delete registry`.
- **No registry name on artifacts.** Because there is only one source, tracked
  artifacts in `track.yaml` carry **no per-registry provenance**: an artifact's
  identity is `(kind, name)` and there is no `spec.registry` field and no
  source-tagged namespacing segment.
- **No registry argument on commands.** `install`, `uninstall`, and the
  catalogue commands take **no `<registry>` positional argument**; they operate
  against the one configured registry. List and describe output carry no
  `REGISTRY` column.

The single configured registry is still reached over one of the existing
transports (`git` or `http`) and validated the same way; only its
multiplicity and its naming are removed.

## Consequences

**Positive**

- The command grammar is smaller and the everyday path is shorter: no name to
  invent or repeat, no `<registry>` argument to thread through
  install/uninstall/catalogue, and no `REGISTRY` column to read past.
- State is simpler: one setting in `settings.yaml` instead of a collection file,
  and tracked artifacts identified by `(kind, name)` with no source tag to keep
  consistent.
- The plural surface is not frozen into v1.0.0 before a real second-source need
  exists, so the eventual multi-registry design is unconstrained by an early
  guess.
- `set`/`unset`/`describe registry` reuse the existing settings idiom (the same
  shape the provider follows), so the registry needs no bespoke lifecycle
  commands.

**Negative**

- A developer cannot mix sources — e.g. an organization registry alongside a
  personal one — in v1. Changing sources means `unset registry` then `set
  registry`, which orphans artifacts installed from the previous source rather
  than tracking them per source.
- Restoring multiplicity later is a breaking change to both state and grammar:
  re-introducing a named `Registry` collection, a `spec.registry` provenance
  field on every tracked artifact, the `<registry>` argument, and the
  `add`/`list`/`delete registry` commands. Artifacts tracked under the single
  registry would need migrating to carry an explicit source.
- The `(kind, name)` identity assumes one source; the day a second source
  appears, name collisions across sources become possible and the identity must
  widen to include the registry.

## Revisit when

A concrete need for more than one registry source arises — for example, a
developer who must mix an organization's registry with a personal one in the same
environment. At that point the deferred multi-registry shape is reconstituted: a
named `Registry` collection persisted in `registries.yaml`; `add registry
<name>` / `list registries` / `delete registry <name>` for its lifecycle;
artifact identity widened to the `(kind, registry, name)` triple with a
`spec.registry` provenance field in `track.yaml`; and a `<registry>` argument
threading registry-scoped `install`, `uninstall`, and catalogue operations, with
a `REGISTRY` column on list and describe output.

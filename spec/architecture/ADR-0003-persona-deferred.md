# ADR-0003: The persona artifact kind is deferred until after v1.0.0

**Status**: Accepted
**Date**: 2026-06-25
**Scope**: Project-wide

## Context

The domain model defines three artifact kinds: **skill**, **agent**, and
**persona**. A persona is a named grouping that references a set of skills and
agents **within the same registry**. It is first-class — installed, listed, and
described like any other artifact — and its realized content is not text of its
own but its resolved **membership**: the concrete set of member skills and agents
the grouping names.

Persona was fully specified across the install, uninstall, list-catalogue,
list-artifacts, and describe features, plus a resolution capability and a
persisted track-file document. Yet every persona behavior is a composition over
behavior that skills and agents already provide on their own: resolving a
membership, installing each member, cascading an uninstall, and re-resolving on
sync. None of it is reachable until a registry actually offers persona
definitions and a user has a curated bundle to distribute.

Shipping persona in v1.0.0 would mean carrying membership resolution, the
provenance bookkeeping that separates a directly-installed artifact from one a
persona brought in, and the sync/upgrade re-resolution diff — all to serve a
grouping no early registry yet publishes. The skill and agent kinds deliver the
whole install/uninstall/list/describe/sync surface without it. The cost of
persona is real and its near-term value is zero, so v1.0.0 ships **only skill and
agent**, and persona is deferred.

This ADR is the durable home of the persona design. It exists so the persona
contracts and capability can be deleted from the feature specs without losing the
design that was worked out for them. The persisted shape is retained as the
[Persona schema](../contracts/schemas/Persona.schema.json); the behavioral design
is recorded below.

## Decision

The **persona** artifact kind is a defined domain concept that is **not
implemented in v1.0.0**. v1.0.0 ships the `skill` and `agent` kinds only. No CLI
command accepts `persona`, no track-file `Persona` document is written, and no
persona membership is resolved. The design below is recorded for the
implementation that reintroduces it.

**Concept.** A persona is a first-class artifact whose realized content is its
**membership** — the set of member skills and agents it references, all within
the persona's own registry. A persona never references across registries.

**Resolution** (the enabling capability for install, uninstall, sync, and
upgrade):

- Resolving a persona reads its definition and produces its membership: the
  concrete member skills and agents it references within its own registry.
- The resolved membership is recorded as the persona's `members` snapshot, and
  the persona is added to each member's provenance `personas` list.
- A member the registry does not offer is reported as unresolved and the rest of
  the membership still resolves — an unresolved member does not abort the
  operation.
- On reconcile, a member is removed only when its provenance has `direct: false`
  **and** no personas still claim it. A member that was also installed directly,
  or is brought in by another persona, is retained.

**Provenance** is what makes the cascade safe. Each installed artifact records,
in the track file, whether it was installed **directly** and **which personas**
brought it in. This distinction lets uninstall and sync remove exactly the
members a persona owned and nothing a user or another persona still wants.

**Install** — `install persona <registry> <name>...`: resolves each named
persona's membership and installs every member. The plan is grouped under a
`personas:` heading and the `skills:`/`agents:` headings for the members it
brings in (`+` additions, `~` updates), followed by a summary count. An
unresolved member is reported without stopping the run.

**Uninstall** — `uninstall persona <registry> <name>... [--dry-run]`: removes the
members a persona brought in, keeping any member also installed directly or
brought in by another persona (per the reconcile rule above). The plan is grouped
under `personas:`/`skills:`/`agents:` with `-` for removals and a summary count.
Uninstalling a persona that is not installed reports nothing was removed and exits
`0`. `--dry-run` prints the plan without touching the environment or track file.

**Sync and upgrade** re-resolve persona membership against the registry. Sync
applies both additions and removals (newly-added members are installed, members
that vanished from the membership are removed subject to provenance); upgrade is
non-destructive — it installs newly-added members but never removes.

**List catalogue** — `list catalogue persona <registry>`: browses the personas a
registry offers, live and paginated, with each entry able to summarize the
membership it would resolve to.

**List installed** — `list personas`: lists installed personas, one per row, with
a `members` field that surfaces the resolved skills and agents the persona brings
in (the field that distinguishes a persona row from a skill/agent row).

**Describe** — `describe persona <name>`: shows one installed persona's full
detail, including its resolved membership and its `digest`, `installed`, and
`updated` facts.

**Persisted shape.** An installed persona is a `Persona` track-file document, the
shape retained at [Persona.schema.json](../contracts/schemas/Persona.schema.json):
the shared `metadata` envelope plus a `spec` carrying `registry`, optional
`version`, `digest` (content identity of the persona definition, for change
detection), the `members` snapshot (`skills` and `agents` arrays, last-resolved,
for diffing on sync/upgrade), and `installedAt`/`updatedAt`. The
`registry` + `kind` + `name` triple is the unique key, consistent with the other
artifact kinds.

## Consequences

**Positive**

- v1.0.0 ships a smaller, fully exercised surface: every install, uninstall,
  list, describe, and sync path is reachable through the skill and agent kinds,
  with no code carried for a grouping no registry yet offers.
- The membership-resolution capability and the provenance cascade — the parts
  most likely to harbor subtle bugs — are not shipped before there is a real bundle
  to validate them against.
- The design is not lost: this ADR plus the retained
  [Persona schema](../contracts/schemas/Persona.schema.json) preserve the concept,
  resolution rules, provenance model, command surface, and persisted shape, so the
  persona contracts and capability can be deleted from the feature specs cleanly.

**Negative**

- A user wanting to distribute a curated bundle of skills and agents as one named
  unit must, in v1.0.0, install each member individually; there is no single name
  that pulls them in together.
- Reintroducing persona later means re-adding the per-artifact provenance
  (`direct` flag and `personas` list) and the sync/upgrade re-resolution diff,
  which the v1.0.0 track-file handling does not implement.
- The domain vocabulary names a kind the v1.0.0 product does not expose, so the
  glossary and README must be clear that persona is a defined-but-deferred concept
  rather than a shipped one.

## Revisit when

Teams need to distribute a curated bundle of skills and agents as one named unit
— i.e. a registry begins offering persona definitions and users want to install,
maintain, and remove the bundle by a single name rather than member by member.

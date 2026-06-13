# Spec Authoring Rules

Normative rules for authoring and organizing specifications in this registry.
Domain concepts live in [the spec README](README.md); the conventions every CLI
command obeys are defined in [CLI conventions](#cli-conventions) below, and the
compiled per-command reference lives in [contracts/cli.md](contracts/cli.md).
When authoring specs in this repo, the `authoring-specs` skill loads and points
here.

## Spec types

Every spec declares one of two types:

- **Feature** — user-observable behavior. Requirements are phrased around what
  the user can see: commands, inputs, outputs, exit codes, and observable
  outcomes. A feature that owns a CLI command also owns a
  `contracts/command-line.md`.
- **Capability** — platform/technical behavior that enables one or more
  features (transports, validation mechanics, scheduling internals). A
  capability has **no CLI surface of its own** and lives nested under the
  feature that introduces it, at `NNNN-<feature>/capabilities/<name>.md`.

**Litmus test:** if it owns a command the user invokes, it is a feature.

## Cross-references

- Vocabulary: `Realized by` (feature → capability), `Enables`
  (capability → feature), `Depends on` (feature → feature).
- **Link form (normative):** every cross-file or cross-feature reference — in
  cross-link declarations, requirement text, Decision Records, and Notes — is a
  relative markdown link to the provider file, resolved from the referencing
  file's own location. For example, a feature's `spec.md` declares
  `Depends on: [sync](../0006-sync-artifacts/spec.md)`, and a
  `data/configuration.md` links the schema as
  `[configuration data contract](../../contracts/configuration.md)`. Bare ids or
  unlinked feature names are not allowed.

## Numbering and layout

- Specs are numbered chronologically, not hierarchically: `NNNN-<kebab-name>`.
  Numbers are never reused.
- Feature directory layout:

  ```
  spec/NNNN-<name>/
  ├── spec.md                      required
  ├── contracts/command-line.md    required when the feature owns a command
  ├── data/configuration.md        required when the feature touches config
  ├── capabilities/<name>.md       optional, one file per capability
  └── architecture/ADR-NNNN-*.md   optional decision records
  ```

- Global, cross-feature contracts live in `spec/contracts/` — the compiled
  [CLI command reference](contracts/cli.md), the
  [configuration data contract](contracts/configuration.md) (the schema of every
  file Sauron persists), and the
  [architecture contract](contracts/architecture.md).
- Project-level ADRs — cross-cutting decisions owned by no single feature (e.g.
  an accepted dependency vulnerability) — live in
  `spec/architecture/ADR-NNNN-<slug>.md`, numbered sequentially **project-wide**
  (starting at `0001`, never reused), a separate sequence from the feature-scoped
  ADRs below.

Each file has a fixed purpose and a section that defines how its content is
written:

| Path | Present when | Holds | Content rules |
|---|---|---|---|
| `spec.md` | always | the feature or capability: overview, EARS requirements, key entities | [Required sections](#required-sections), [EARS templates](#ears-templates-normative) |
| `contracts/command-line.md` | the feature owns a command | the command's synopsis, arguments, flags, output, and exit codes | [CLI conventions](#cli-conventions) (and the `authoring-cli-contracts` skill) |
| `data/configuration.md` | the feature reads or writes config | which configuration file(s) and fields the feature owns or writes, the feature-specific read/write semantics, and the field→requirement (`FR-NNN`) realization for the fields it owns — **not** the schema, which is owned by [contracts/configuration.md](contracts/configuration.md) and linked from here | [Glossary](#glossary) terms; link the [configuration data contract](contracts/configuration.md). The contract never links back to feature requirements (one-directional, no cycle) |
| `capabilities/<name>.md` | the feature introduces a capability | one nested technical capability with no CLI surface | [Required sections](#required-sections) |
| `architecture/ADR-NNNN-<slug>.md` | a significant decision needs recording | one architectural decision and its rationale | [ADR structure](#adr-structure) |

## Required sections

`spec.md`, in this order:

1. `# <Title>`
2. Header block: `**Type:** feature` plus cross-links
   (`**Realized by:**` / `**Depends on:**`) as markdown links, each on its own
   line with a blank line between them.
3. `## Overview` — the user's need and intent, in problem/solution form.
4. `## Requirements` — subsections in this order, omitting empty ones:
   `### Ubiquitous`, `### Event-driven`, `### State-driven`,
   `### Unwanted behavior`, `### Optional`.
5. `## Key Entities`
6. `## Decision Records` — optional; links to `architecture/ADR-*.md`.
7. `## Notes` — optional.

Capability files (`capabilities/<name>.md`), in this order: title, header block
(`**Type:** capability`, `**Enables:**` link), `## Overview`, `## Requirements`
(same EARS subsections), optional `## Decision Records`.

Requirement ids are `FR-NNN`, three digits, sequential within one file.

## EARS templates (normative)

One sentence shape per pattern:

- Ubiquitous: `Sauron shall <behavior>.`
- Event-driven: `When <trigger>, Sauron shall <behavior>.`
- State-driven: `While <state>, Sauron shall <behavior>.`
- Unwanted behavior: `If <condition>, then Sauron shall <behavior>.`
- Optional: `Where <option applies>, Sauron shall <behavior>.`

## ADR structure

An Architecture Decision Record captures one significant technical decision for
its feature. Files are named `architecture/ADR-NNNN-<kebab-slug>.md`, where
`NNNN` is sequential **within the feature** (each feature's ADRs start at
`0001`) and never reused. Every ADR is linked from the spec's
`## Decision Records` section. A **project-level ADR** — a cross-cutting decision
owned by no single feature — instead lives at
`spec/architecture/ADR-NNNN-<kebab-slug>.md`, numbered project-wide, and is not
linked from any feature's `## Decision Records`.

Structure, in this order:

1. `# ADR-NNNN: <decision as a short declarative title>`
2. Header fields, one per line:
   - `**Status**:` — `Accepted`, or `Superseded by [ADR-NNNN](ADR-NNNN-<slug>.md)`
     once replaced.
   - `**Date**:` — `YYYY-MM-DD`.
   - `**Feature**:` — the feature's human-readable name. A project-level ADR
     carries `**Scope**:` in its place — `Project-wide`, or an area such as
     `Dependencies / Security`.
3. `## Context` — the forces, constraints, and problem that make a decision
   necessary.
4. `## Decision` — what was decided, in the present tense; reference the
   `FR-NNN` ids it satisfies where relevant.
5. `## Consequences` — the results of the decision, grouped under `**Positive**`
   and `**Negative**`.
6. `## Revisit when` — the condition that would reopen the decision.

An accepted ADR is not rewritten: a changed decision is recorded as a new ADR
that supersedes it, and the old one's `**Status**` is updated to point at the
replacement.

Every ADR — feature or project-level — is authored only with explicit user
intent and is never generated automatically.

## Glossary

One canonical term per concept; specs do not use synonyms for these:

| Term | Meaning |
|---|---|
| artifact | A skill or an agent distributed by Sauron |
| skill | An artifact under a registry's `.skills/` directory |
| agent | An artifact under a registry's `.agents/` directory |
| registry | A registered source of artifacts |
| kind | A registry's or backend's type: `http`, `filesystem`, or `git`; it selects how the source is validated and how artifacts are fetched |
| persona | A named set of artifacts shared by a group |
| backend | The singleton source that owns persona definitions; the persona analog of a `registry` (one per instance) |
| catalog | The live view of *available* personas, assembled at command time from the installed personas plus a live fetch from the backend; it is never persisted, and the backend portion is omitted when the backend is unreachable |
| installed persona | A persona activated locally via `set persona`, stored with its definition in `personas.yaml`; it participates in artifact sync and carries a priority |
| available persona | A persona the backend offers live but that is not installed; it appears in the catalog yet has no entry in `personas.yaml` |
| provider | The destination environment where artifacts are installed (e.g. `claude`, `zencoder`); a single global setting recorded in `settings` |
| priority | Integer precedence, always defined and unique within its kind; lower value wins (`0` is highest). See the [priority model](#priority-model) |
| pin | A user-declared binding of an artifact to a specific registry that overrides priority-based conflict resolution; recorded as `pinned` on the artifact's `track file` entry |
| sync | Either reconcile operation: `sync artifacts` reconciles the provider with the desired artifact set, while `sync personas` refreshes the installed personas' definitions from the backend |
| plan | The printed list of pending additions/removals (`+`/`-` lines) |
| track file | `track.yaml`, recording installed artifacts and provenance |
| provenance | The origin recorded for each installed artifact in the `track file`: its source registry, the installed persona that brought it in (when any), and the provider |
| configuration | The set of files Sauron persists under `~/.sauron/` — `registries.yaml`, `backend.yaml`, `personas.yaml`, `track.yaml`, and `settings.yaml` — whose schema is owned by the [configuration data contract](contracts/configuration.md) |
| settings | `settings.yaml`, the global settings file: the active `provider` and the sync `schedules` |

## Canonical boilerplate

Shared semantics use these sentences verbatim (entity substituted):

- Idempotent deletion: `When a user deletes a <entity> that does not exist,
  Sauron shall exit successfully and report that nothing was deleted.`
- Dry run: `Where --dry-run is provided, Sauron shall print the plan without
  changing the environment or the track file.`
- Validation transactionality: `While a <entity> is being validated, Sauron
  shall leave the existing configuration unchanged until validation succeeds.`
- Usage error: `If required arguments or flags are missing or invalid, then
  Sauron shall exit with code 2 without executing the command.`
- Failure output: `If a command fails, then Sauron shall write exactly one
  human-readable message to stderr.`

## CLI conventions

Command grammar, shared flags, exit-status semantics, and output discipline are
normative here. Every command's `contracts/command-line.md` conforms to them,
and the compiled [CLI command reference](contracts/cli.md) summarizes each
command. Per-feature contracts may refine which conditions map to which exit
code; they may not redefine the meanings.

### Command grammar

```
sauron <verb> [<noun> [<noun>]] [flags] <args...>
```

- Verb–noun hierarchy: `add registry`, `list personas`,
  `set priority registry`, `unset backend`, `describe persona`,
  `schedule sync artifacts`, `pin skill`. `unset` is the inverse of `set` (clears
  a setting or selection, as opposed to `delete`, which destroys an owned
  resource); `unschedule` is the inverse of `schedule` (removes a scheduled job);
  `unpin` is the inverse of `pin` (removes an artifact's registry binding);
  `describe` shows a single resource's detail.
- Flags are GNU-style long options: `--flag` for booleans, `--flag <value>`
  otherwise. Repeatable flags are marked `...` in synopses.
- Positional arguments follow flags in synopses and are written `<name>`.

### Shared flags

These flags mean the same thing in every command that accepts them. A feature
contract may narrow a shared flag (e.g. restrict `--sort` values) but may not
contradict it.

| Flag | Meaning |
|---|---|
| `--dry-run` | Print the plan without changing the environment or the track file |
| `--priority <n>` | Optional integer precedence (lower value wins), unique within its kind; the first resource is `0`, an omitted value appends at the end (`max + 1`). See the [priority model](#priority-model) |
| `--kind <kind>` | Registry kind: `http` (default), `filesystem`, or `git` |
| `--search <term>` | Case-insensitive substring filter |
| `--sort <field>` | Sort field for list output |
| `--order <asc\|desc>` | Sort direction, default `asc` |
| `--persona <name>` | Scope the operation to one persona's artifacts |
| `--fields <list>` | Comma-separated columns to display, in order, for `list` and `describe`; the identity field is always present and first. Each contract defines its valid field set |
| `--force` | Re-pull authoritatively, reconciling away local entries no longer present upstream |
| `--reconcile` | Apply the change immediately by reconciling the affected artifacts (a scoped sync), instead of only recording it |
| `--timeout <duration>` | Bound on network operations, default `30s` |

### Exit status

| Code | Meaning |
|---|---|
| `0` | Success — including idempotent no-ops: deleting an absent resource, an empty list, an already-up-to-date sync, an already-set value, and any `--dry-run` run |
| `2` | Usage error — invalid or missing arguments/flags; nothing was executed |
| `1` | Runtime error — validation failure, unreadable configuration or track file, unreachable external resource, or a failed artifact operation |

Feature contracts may only refine *which conditions* map to each code; they
never redefine these meanings.

### Output discipline

- Results (confirmations, tables, plans) go to stdout.
- A failing command writes exactly one human-readable message to stderr and
  produces no partial output.
- Commands that apply changes in bulk (`sync artifacts`, `sync personas`, `prune artifacts`, `delete artifacts`, `set provider`)
  print a shared plan/report format: artifacts grouped under `skills:` and
  `agents:` headings, one artifact per line, prefixed `+` for additions/updates
  and `-` for removals, followed by a summary count line when changes are
  applied. Per-artifact failures are reported without stopping the run.

## Priority model

Registries and personas share one priority model; they differ only in *when and
how* a priority is assigned:

- A **registry** takes its priority at add time, via an optional `--priority` on
  [add registry](0001-add-registry/spec.md).
- A **persona** takes its priority at install time, from its position in
  [set personas](0014-select-personas/spec.md)' `set persona` argument order
  (the first listed persona is highest precedence).

Both then obey the same rules:

- **First resource of its kind** is priority `0` — for a registry, an omitted
  `--priority` defaults to `0` and an explicit `--priority` is accepted only when
  it is `0`; for personas, the first persona in the `set persona` list.
- **Each subsequent resource** appends at the end of the priority-ordered list —
  one greater than the current highest priority (`max + 1`) — which never
  collides. For a registry, an explicit `--priority <n>` overrides this and is
  rejected when another registry already holds it.
- Priority is **always defined** and **unique** within its kind. Lower value
  wins; `0` is the highest precedence.
- Priorities change afterward only through
  [set registry priority](0008-set-registry-priority/spec.md) and
  [set persona priority](0007-set-persona-priority/spec.md), each **blocked while
  a single resource of that kind exists** — that lone resource keeps `0`.
  Re-running `set persona` redeclares the installed set and resets positional
  priorities.

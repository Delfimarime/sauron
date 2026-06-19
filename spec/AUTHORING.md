# Spec Authoring Rules

Normative rules for authoring and organizing specifications in this registry.
Domain concepts live in [the spec README](README.md); the conventions every CLI
command obeys — command grammar, shared flags, exit status, and output
discipline — are the [CLI contract](contracts/cli.md). When authoring specs in
this repo, the `sauron-authoring-specs` skill loads and points here.

## Spec types

Every spec declares one of two types:

- **Feature** — user-observable behavior. Requirements are phrased around what
  the user can see: commands, inputs, outputs, exit codes, and observable
  outcomes. A feature **owns one or more commands**; for each command it owns it
  has a `contracts/<verb>-<noun>.md` contract file.
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
  `Depends on: [install](../0006-install-artifacts/spec.md)`, and a
  `data/state.md` links the schema as
  `[state data contract](../../contracts/state.md)`. Bare ids or
  unlinked feature names are not allowed.

## Numbering and layout

- Specs are numbered chronologically, not hierarchically: `NNNN-<kebab-name>`.
  Numbers are never reused.
- Feature directory layout:

  ```
  spec/NNNN-<name>/
  ├── spec.md                      required
  ├── contracts/<verb>-<noun>.md   one per command the feature owns
  ├── data/state.md                required when the feature touches persisted state
  ├── capabilities/<name>.md       optional, one file per capability
  └── architecture/ADR-NNNN-*.md   optional decision records
  ```

  A feature owning a command family (e.g. `list skills` / `list agents` /
  `list personas`) has one contract file per command
  (`contracts/list-skills.md`, …), each specifying that command's own output.

- Global, cross-feature contracts live in `spec/contracts/` — the
  [CLI contract](contracts/cli.md) (the command conventions), the
  [state data contract](contracts/state.md) (the schema of every document Sauron
  persists), the [architecture contract](contracts/architecture.md) (code
  structure and wiring), and the [delivery contract](contracts/delivery.md)
  (build, gates, CI/CD, versioning).
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
| `contracts/<verb>-<noun>.md` | per command the feature owns | the command's synopsis, arguments, flags, output, and exit codes | [CLI contract](contracts/cli.md) (and the `sauron-authoring-cli-contracts` skill) |
| `data/state.md` | the feature reads or writes persisted state | which state document(s) and fields the feature owns or writes, the feature-specific read/write semantics, and the field→requirement (`FR-NNN`) realization for the fields it owns — **not** the schema, which is owned by [contracts/state.md](contracts/state.md) and linked from here | [Glossary](#glossary) terms; link the [state data contract](contracts/state.md). The contract never links back to feature requirements (one-directional, no cycle) |
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
`spec/architecture/ADR-NNNN-<slug>.md`, numbered project-wide, and is not
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
| artifact | A unit Sauron distributes: a skill, an agent, or a persona |
| skill | An artifact hosted under a registry's `.skills/` directory |
| agent | An artifact hosted under a registry's `.agents/` directory |
| persona | A first-class artifact that references a set of skills and agents within the same registry; installed, listed, and described like any artifact |
| membership | The set of skills and agents a persona references; resolved at install and re-resolved by `sync`/`upgrade` |
| registry | A registered source of artifacts |
| transport | A registry's type — `git`, `http`, or `filesystem` — selecting how the source is reached, validated, and fetched from; persisted as `spec.transport` and selected at the CLI by `--kind` |
| kind | In a manifest, the document type (`Registry`, `Skill`, `Agent`, `Persona`, `Provider`, `Schedule`). At the CLI, the `--kind` flag selects a registry's `transport` |
| ref | A git revision — a branch, tag, or commit — a `git`-transport registry is pinned to; persisted as `spec.ref` and selected at the CLI by `--ref`; when absent, the repository's default branch is used |
| catalogue | The live, paginated view of what a registry offers, fetched fresh at command time; it is never persisted and has no offline form |
| provider | The destination environment where artifacts are installed (`claude`, `zencoder`); a single global setting recorded as the `Provider` document in `settings` |
| namespacing | The installed-target naming `sauron-<registry>-<name>`, which lets two registries offer the same artifact name without conflict; the `sauron-` prefix marks Sauron ownership |
| install | Fetching named artifacts from a registry and placing them under the provider; installing a persona installs its members |
| uninstall | Removing named installed artifacts; uninstalling a persona removes the members it brought in |
| sync | The full reconcile of the installed set against its sources: refresh, drift repair, removal of what vanished upstream, and persona membership re-resolution (additions and removals) |
| upgrade | The non-destructive refresh of the installed set: refresh changed artifacts and add newly-added persona members; never removes |
| plan | The printed list of pending changes — `+` additions, `~` updates, `-` removals |
| digest | The content identity recorded per artifact, used to detect change and local drift; always present |
| version | An optional, human-meaningful artifact label; derivable for `git` (the last commit touching the artifact directory), declared-only otherwise |
| provenance | The origin recorded for each installed artifact in the `track file`: whether it was installed directly and which personas brought it in |
| track file | `track.yaml`, the multi-document stream of `Skill`/`Agent`/`Persona` manifests recording installed artifacts and provenance |
| state | The set of files Sauron persists under `~/.sauron/` — `registries.yaml`, `track.yaml`, and `settings.yaml` — whose schema is owned by the [state data contract](contracts/state.md). Distinct from the `Configuration` DI struct, which is app configuration (resolved home), not persisted state |
| settings | `settings.yaml`, holding the `Provider` document and the `Schedule` documents |
| manifest | A persisted document carrying `apiVersion` (`sauron.raitonbl.com/v1`) and `kind`, with `metadata` and `spec`, in the spirit of a Kubernetes object |

## Canonical boilerplate

Shared semantics use these sentences verbatim (entity substituted):

- Idempotent deletion: `When a user uninstalls a <entity> that is not installed,
  Sauron shall exit successfully and report that nothing was removed.`
- Dry run: `Where --dry-run is provided, Sauron shall print the plan without
  changing the environment or the track file.`
- Validation transactionality: `While a <entity> is being validated, Sauron
  shall leave the existing state unchanged until validation succeeds.`
- Usage error: `If required arguments or flags are missing or invalid, then
  Sauron shall exit with code 2 without executing the command.`
- Failure output: `If a command fails, then Sauron shall write exactly one
  human-readable message to stderr.`

## CLI conventions

The command conventions — command grammar, shared flags, exit-status semantics,
and output discipline — are normative in the [CLI contract](contracts/cli.md).
Every command's `contracts/<verb>-<noun>.md` conforms to them; a feature contract
may refine which conditions map to which exit code but may not redefine the
meanings.

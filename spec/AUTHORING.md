# Spec Authoring Rules

Normative rules for authoring and organizing specifications in this repository.
Domain concepts live in [the spec README](README.md) and the
[glossary](GLOSSARY.md); the conventions every CLI
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
  `Depends on: [install](../0005-install-artifacts/spec.md)`, and a
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
  ├── architecture/ADR-NNNN-*.md   optional decision records
  ├── plan.md                      optional implementation plan (how the work is built)
  └── TASKS.md                     optional verifiable task breakdown, paired with plan.md
  ```

  A feature owning a command family (e.g. `list skills` / `list agents`) has one
  contract file per command (`contracts/list-skills.md`, …), each specifying that
  command's own output.

- Global, cross-feature contracts live in `spec/contracts/` — the
  [CLI contract](contracts/cli.md) (the command conventions), the
  [state data contract](contracts/state.md) (the schema of every document Sauron
  persists), the [architecture contract](contracts/architecture.md) (code
  structure and wiring), the [delivery contract](contracts/delivery.md)
  (build, gates, CI/CD, versioning), and the
  [HTTP Registry API contract](contracts/registry-http-api.oas3.yaml) (the REST
  API an `http` registry server implements).
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
| `data/state.md` | the feature reads or writes persisted state | which state document(s) and fields the feature owns or writes, the feature-specific read/write semantics, and the field→requirement (`FR-NNN`) realization for the fields it owns — **not** the schema, which is owned by [contracts/state.md](contracts/state.md) and linked from here | [glossary](GLOSSARY.md) terms; link the [state data contract](contracts/state.md). The contract never links back to feature requirements (one-directional, no cycle) |
| `capabilities/<name>.md` | the feature introduces a capability | one nested technical capability with no CLI surface | [Required sections](#required-sections) |
| `architecture/ADR-NNNN-<slug>.md` | a significant decision needs recording | one architectural decision and its rationale | [ADR structure](#adr-structure) |
| `plan.md` | a feature needs an implementation plan | how the work is built: goal & scope, pre-requirements, design, **checkpoints** (each with a verify command), key decisions, a link to `TASKS.md`, and — once the feature ships — an optional `## Distill` section (see below) | each checkpoint states the command/criterion that verifies it; the plan links `TASKS.md` |
| `TASKS.md` | a `plan.md` is present | the executable task breakdown — one task per unit of work, each owning its files, a single verification command, and its dependencies, with an overall order | **every task is independently verifiable**: it states the command/criterion that confirms it. A task without a pass/fail check is not a task |

The `## Distill` section of a `plan.md` is the loop back from delivery to the spec
([Constitution Ch. IV, Art. 5](../CONSTITUTION.md)). It is added only when there is
something to record — a feature that ships exactly as planned has nothing to
distill, so the section is omitted, never left as an empty placeholder. When
present it is a table, one row per insight:

| Insight | Source | Disposition |
|---|---|---|
| what delivery or use revealed, in one line | `deliver` \| `usage` \| `bug` \| `retro` | how it was closed: `→ FR-NNN amendment`, `→ PROPOSAL #NN`, or `won't-fix` |

## Required sections

`spec.md`, in this order:

1. `# <Title>`
2. Header block, each field on its own line with a blank line between:
   - `**Type:** feature`.
   - `**Status:**` — `Built` (shipped and verified), `Specified` (approved spec,
     not yet implemented), or `Partial — <what ships> (see Notes)` when a feature
     ships some requirements and defers others. This is the single source of a
     feature's build status; the [spec index](README.md) aggregates it as a view.
   - `**Audience:**` — optional; include only when the feature narrows from the
     project-wide audience declared in the [spec README](README.md) (e.g. a
     capability only an architect or security analyst needs). Omit otherwise.
   - Cross-links (`**Realized by:**` / `**Depends on:**`) as markdown links.
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

The canonical domain vocabulary lives in [GLOSSARY.md](GLOSSARY.md) — one term
per concept, no synonyms. Use those terms in every spec and contract; link the
glossary rather than restating a definition.

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

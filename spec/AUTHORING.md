# Spec Authoring Rules

Normative rules for authoring and organizing specifications in this repository.
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
  relative markdown link to the target file, e.g.
  `Depends on: [sync](../0009-sync/spec.md)`. Bare ids or unlinked feature
  names are not allowed.

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

- Global, cross-feature contracts live in `spec/contracts/` — e.g. the compiled
  [CLI command reference](contracts/cli.md) and the
  [architecture contract](contracts/architecture.md).

Each file has a fixed purpose and a section that defines how its content is
written:

| Path | Present when | Holds | Content rules |
|---|---|---|---|
| `spec.md` | always | the feature or capability: overview, EARS requirements, key entities | [Required sections](#required-sections), [EARS templates](#ears-templates-normative) |
| `contracts/command-line.md` | the feature owns a command | the command's synopsis, arguments, flags, output, and exit codes | [CLI conventions](#cli-conventions) (and the `authoring-cli-contracts` skill) |
| `data/configuration.md` | the feature reads or writes config | the `settings`/`track file` schema the feature touches: location, format, schema, write semantics | [Glossary](#glossary) terms `settings` and `track file` |
| `capabilities/<name>.md` | the feature introduces a capability | one nested technical capability with no CLI surface | [Required sections](#required-sections) |
| `architecture/ADR-NNNN-<slug>.md` | a significant decision needs recording | one architectural decision and its rationale | [ADR structure](#adr-structure) |

## Required sections

`spec.md`, in this order:

1. `# <Title>`
2. Header block: `**Type:** feature` plus cross-links
   (`**Realized by:**` / `**Depends on:**`) as markdown links.
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
`## Decision Records` section.

Structure, in this order:

1. `# ADR-NNNN: <decision as a short declarative title>`
2. Header fields, one per line:
   - `**Status**:` — `Accepted`, or `Superseded by [ADR-NNNN](ADR-NNNN-<slug>.md)`
     once replaced.
   - `**Date**:` — `YYYY-MM-DD`.
   - `**Feature**:` — the feature's human-readable name.
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

## Glossary

One canonical term per concept; specs do not use synonyms for these:

| Term | Meaning |
|---|---|
| artifact | A skill or an agent distributed by Sauron |
| skill | An artifact under a repository's `.skills/` directory |
| agent | An artifact under a repository's `.agents/` directory |
| repository | A registered source of artifacts |
| kind | A repository's type: `http`, `filesystem`, or `git` |
| persona | A named set of artifacts shared by a group |
| target | The provider destination (e.g. `claude`, `zencoder`) |
| priority | Integer precedence; lower value wins |
| sync | The operation that reconciles the target with repositories/personas |
| plan | The printed list of pending additions/removals (`+`/`-` lines) |
| track file | `track.yaml`, recording installed artifacts and provenance |
| settings | `settings.yaml`, the persisted configuration |

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

- Verb–noun hierarchy: `add repository`, `list personas`,
  `set priority repository`, `cron sync`.
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
| `--priority <n>` | Optional integer precedence (lower value wins), unique within its kind; the first resource is `0`, an omitted value appends at the end (`max + 1`). See the [priority model](0005-import-persona/architecture/ADR-0002-unified-priority-model.md) |
| `--kind <kind>` | Repository kind: `http` (default), `filesystem`, or `git` |
| `--search <term>` | Case-insensitive substring filter |
| `--sort <field>` | Sort field for list output |
| `--order <asc\|desc>` | Sort direction, default `asc` |
| `--persona <name>` | Scope the operation to one persona's artifacts |
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
- Commands that apply changes in bulk (`sync`, `prune`, `clear`, `set target`)
  print a shared plan/report format: artifacts grouped under `skills:` and
  `agents:` headings, one artifact per line, prefixed `+` for additions/updates
  and `-` for removals, followed by a summary count line when changes are
  applied. Per-artifact failures are reported without stopping the run.

# Spec Authoring Rules

Normative rules for authoring and organizing specifications in this repository.
Domain concepts live in [the spec README](README.md); CLI-wide behavior lives in
[the CLI contract](contracts/cli.md); project principles live in
[CONSTITUTION.md](../CONSTITUTION.md). When authoring specs in this repo, the
`authoring-specs` skill loads and points here.

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

- Global, cross-feature contracts live in `spec/contracts/`
  (e.g. [the CLI contract](contracts/cli.md)).

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

## CLI policy

Command grammar, shared flags, exit-status semantics, and output discipline are
owned by [the CLI contract](contracts/cli.md). Per-feature contracts may refine
which conditions map to which exit code; they may not redefine the meanings.

# Describe Persona

**Type:** feature
**Depends on:** [backend](../0012-backend/spec.md), [select personas](../0014-select-personas/spec.md)

## Overview

A person responsible for a team's agentic-AI setup needs the full detail of a
single persona — not the whole list — so that they can inspect one persona
before deciding to install it, or review the details of one already installed.
Listing the whole [catalog](../0005-list-personas/spec.md) answers "what is
there"; describing one persona answers "what exactly is this one". Describing
resolves the persona by name: an
[installed persona](../0014-select-personas/spec.md) is described from its full
definition in [personas.yaml](../contracts/configuration.md#personasyaml) and is
describable even offline (marked installed); a persona that is not installed is
fetched **live** from the [backend](../0012-backend/spec.md) (marked available).
Resolution is the [live persona view](../contracts/configuration.md#live-persona-view):
the installed personas plus, when the backend is reachable, a live fetch. It is
read-only; it never writes the configuration or the track file.

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to describe a single persona by
  name, resolving it from the
  [installed personas](../0014-select-personas/spec.md) when installed or from a
  live [backend](../0012-backend/spec.md) fetch when it is not.

### Event-driven

- **FR-002**: When a user describes a persona, Sauron shall resolve the persona
  by name against the
  [live persona view](../contracts/configuration.md#live-persona-view) — taking
  an [installed persona](../0014-select-personas/spec.md) from
  [personas.yaml](../contracts/configuration.md#personasyaml) and, for a persona
  that is not installed, fetching it live from the
  [backend](../0012-backend/spec.md) — and print one field per line as
  `field: value` with the identity field `name` first.
- **FR-003**: When a user describes a persona, Sauron shall make available the
  fields `description`, `tags`, `installed`, `priority`, `skills`, `agents`,
  `registry`, `last-updated`, and `last-synced`, showing `skills` and `agents`
  as integer counts and leaving an absent value empty.
- **FR-004**: When the described persona is not
  [installed](../0014-select-personas/spec.md) — resolved live from the
  [backend](../0012-backend/spec.md) — Sauron shall show `priority`,
  `last-updated`, and `last-synced` as empty, since those values exist only for
  installed personas.

### Unwanted behavior

- **FR-005**: If the named persona is neither installed nor offered by a
  reachable [backend](../0012-backend/spec.md) — because it is unknown, or
  because it is not installed and the backend is unreachable — then Sauron shall
  reject the request with a runtime error and report that the persona could not
  be resolved.
- **FR-006**: If the `<name>` argument is missing, then Sauron shall exit with
  code 2 without executing the command.
- **FR-007**: If `--fields` names a field outside the valid set, then Sauron
  shall exit with code 2 without executing the command and report the allowed
  fields.
- **FR-008**: If [personas.yaml](../contracts/configuration.md#personasyaml) or
  the track file exists but cannot be read or parsed, then Sauron shall reject
  the request with a runtime error and report that they cannot be read.
- **FR-009**: If the persona is not installed and the live
  [backend](../0012-backend/spec.md) fetch fails because the backend is
  unreachable, then Sauron shall reject the request with a runtime error and
  report that the persona could not be resolved.
- **FR-010**: If a command fails, then Sauron shall write exactly one
  human-readable message to stderr.

### Optional

- **FR-011**: Where `--fields` is provided, Sauron shall print the named fields
  in the given order, always keeping `name` present and first, and shall print
  the full field set when it is omitted.

## Key Entities

- **Installed persona**: a persona activated locally by
  [select personas](../0014-select-personas/spec.md), stored with its full
  definition in [personas.yaml](../contracts/configuration.md#personasyaml); it
  carries a `priority` and a local last-synced time, both shown only when the
  persona is installed, and is describable even offline. Its schema is owned by
  the [configuration data contract](../contracts/configuration.md#personasyaml).
- **Available persona**: a persona the [backend](../0012-backend/spec.md) offers
  live but that is not installed; it is fetched at command time and carries its
  `description`, `tags`, `skills`/`agents` artifacts, owning `registry`, and the
  backend's last-updated time, but no `priority` or last-synced time. Its
  connection is owned by the
  [backend data contract](../contracts/configuration.md#backendyaml).

## Notes

- **Catalog-removal redesign.** This spec was redefined for the no-persisted-catalog
  model. Sauron no longer reads a persisted `catalog` from `settings.yaml`; the
  glossary [catalog](../AUTHORING.md#glossary) is now the *live* computed view,
  never persisted. An installed persona is described from
  [personas.yaml](../contracts/configuration.md#personasyaml) (offline-capable);
  a not-installed persona is fetched live from the
  [backend](../0012-backend/spec.md). FR ids are preserved and redefined in
  place. **Intentional behavior change:** describing a not-installed persona now
  requires a reachable backend — when the backend is unreachable, a
  not-installed (and any unknown) name fails with a runtime error (exit 1)
  rather than resolving against a local mirror. The prior FR-009 (settings file
  absent ⇒ persona not found) is redefined as the offline / unreachable-backend
  failure; the absence of installed personas is no longer a distinct condition,
  since an unresolved name already fails via FR-005.

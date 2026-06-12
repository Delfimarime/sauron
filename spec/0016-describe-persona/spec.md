# Describe Persona

**Type:** feature
**Depends on:** [backend](../0012-backend/spec.md), [select personas](../0014-select-personas/spec.md)

## Overview

A person responsible for a team's agentic-AI setup needs the full detail of a
single [catalog](../0012-backend/spec.md) persona — not the whole list — so that
they can inspect one persona before deciding to install it, or review the
details of one already installed. Listing the whole
[catalog](../0005-list-personas/spec.md) answers "what is there"; describing one
persona answers "what exactly is this one". Describing resolves the persona by
name against the [catalog](../0012-backend/spec.md) whether it is
[installed](../0014-select-personas/spec.md) or merely available — a
not-yet-installed persona is still describable. It reads the
[catalog](../0012-backend/spec.md) and the
[installed set](../0014-select-personas/spec.md), joins them, and prints one
record. It is read-only and works offline against the local mirror; it never
writes the settings or the track file.

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to describe a single
  [catalog](../0012-backend/spec.md) persona by name, resolving it whether it is
  [installed](../0014-select-personas/spec.md) or available.

### Event-driven

- **FR-002**: When a user describes a persona, Sauron shall read the
  [catalog](../0012-backend/spec.md) and the
  [installed set](../0014-select-personas/spec.md) from the settings, resolve the
  persona by name, and print one field per line as `field: value` with the
  identity field `name` first.
- **FR-003**: When a user describes a persona, Sauron shall make available the
  fields `description`, `tags`, `installed`, `priority`, `skills`, `agents`,
  `registry`, `last-updated`, and `last-synced`, showing `skills` and `agents`
  as integer counts and leaving an absent value empty.
- **FR-004**: When the described persona is not
  [installed](../0014-select-personas/spec.md), Sauron shall show `priority`,
  `last-updated`, and `last-synced` as empty, since those values exist only for
  installed personas.

### Unwanted behavior

- **FR-005**: If no persona in the [catalog](../0012-backend/spec.md) matches the
  given name, then Sauron shall reject the request with a runtime error and
  report that the persona was not found.
- **FR-006**: If the `<name>` argument is missing, then Sauron shall exit with
  code 2 without executing the command.
- **FR-007**: If `--fields` names a field outside the valid set, then Sauron
  shall exit with code 2 without executing the command and report the allowed
  fields.
- **FR-008**: If the settings or the track file exist but cannot be read or
  parsed, then Sauron shall reject the request with a runtime error and report
  that they cannot be read.
- **FR-009**: If the settings file does not exist, then Sauron shall treat the
  catalog as empty and report that the persona was not found.
- **FR-010**: If a command fails, then Sauron shall write exactly one
  human-readable message to stderr.

### Optional

- **FR-011**: Where `--fields` is provided, Sauron shall print the named fields
  in the given order, always keeping `name` present and first, and shall print
  the full field set when it is omitted.

## Key Entities

- **Catalog persona**: an entry in the [catalog](../0012-backend/spec.md) — the
  local read-only mirror of persona definitions — identified by its `name` and
  carrying its `description`, `tags`, `skills`/`agents` artifacts, owning
  `registry`, and the backend's last-updated time. Its schema is owned by the
  [backend data model](../0012-backend/data/configuration.md).
- **Installed persona**: a catalog persona activated locally by
  [select personas](../0014-select-personas/spec.md); it carries a `priority` and
  a local last-synced time, both shown only when the persona is installed. Its
  schema is owned by the
  [select personas data model](../0014-select-personas/data/configuration.md).

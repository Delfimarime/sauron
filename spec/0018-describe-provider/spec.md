# Describe Provider

**Type:** feature

**Depends on:** [set provider](../0009-set-provider/spec.md)

## Overview

A person responsible for a team's agentic-AI setup needs the full detail of the
active provider — which provider Sauron delivers to and where it persists skills
and agents — so that they can confirm or troubleshoot the delivery destination
without inspecting the raw settings. The provider is the single global setting
owned by [set provider](../0009-set-provider/spec.md); describing it is the
single-record counterpart that shows the active provider's detail. Because the
provider is a singleton that always has a value (it defaults to `claude`), the
command takes no name and there is no "not configured" case. It is read-only and
works offline: it never contacts an external resource, and never writes the
settings or the track file.

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to describe the active provider,
  taking no name because the provider is a singleton owned by
  [set provider](../0009-set-provider/spec.md).
- **FR-002**: Sauron shall treat describing the provider as read-only and
  offline, contacting no external resource and writing neither the settings nor
  the track file.

### Event-driven

- **FR-003**: When a user describes the provider, Sauron shall read the active
  provider from the settings and display its fields, one per line as
  `field: value`, with the identity field `name` first.
- **FR-004**: When displaying the provider's fields, Sauron shall present `name`
  as the identity — the active provider (e.g. `claude` or `zencoder`) — and
  `skills-location` and `agents-location` as the available fields, naming where
  that provider persists skills and agents.
- **FR-005**: When the active provider has never been set, Sauron shall describe
  the default provider `claude`, since the provider always has a value (see
  [set provider](../0009-set-provider/spec.md)).
- **FR-006**: When a field has no value, Sauron shall show it with an empty
  value.

### Unwanted behavior

- **FR-007**: If a name argument is supplied (e.g. `describe provider claude`),
  then Sauron shall exit with code 2 without executing the command and report
  that the command takes no name.
- **FR-008**: If `--fields` names a field other than `name`, `skills-location`,
  or `agents-location`, then Sauron shall exit with code 2 without executing the
  command and report the allowed fields.
- **FR-009**: If the settings exist but cannot be read or parsed, then Sauron
  shall reject the request and report that the settings cannot be read.
- **FR-010**: If a command fails, then Sauron shall write exactly one
  human-readable message to stderr.

### Optional

- **FR-011**: Where `--fields` is provided, Sauron shall display only the named
  fields in the given order, with the identity field `name` always present and
  first; where it is omitted, Sauron shall display the full field set.

## Key Entities

- **Provider**: the single active provider destination — `claude` (default) or
  `zencoder` — owned by [set provider](../0009-set-provider/spec.md) and stored
  in the settings (`~/.sauron/settings.yaml`). Described by its identity `name`
  and the fields `skills-location` and `agents-location`, which name where that
  provider persists skills and agents. The provider always has a value; its
  schema is owned by the
  [set provider data model](../0009-set-provider/data/configuration.md).

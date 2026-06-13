# Describe Artifact

**Type:** feature

**Depends on:** [sync artifacts](../0006-sync-artifacts/spec.md)

## Overview

A person responsible for a team's agentic-AI setup needs the full detail of a
single managed skill or agent — where it came from, where it was installed, and
whether it is pinned — so that a delivered artifact can be inspected without
reading the track file by hand.

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to describe a managed skill or
  agent by name.

### Event-driven

- **FR-002**: When `describe skill <name>` or `describe agent <name>` runs, Sauron
  shall show the managed artifact's detail from the track file — `name`, `type`,
  source `registry`, `provider`, installed `path`, `pinned` state, and the
  installed `persona` when set.
- **FR-003**: When `--fields <list>` is provided, Sauron shall show only those
  fields, in the given order, with the name first.

### Unwanted behavior

- **FR-004**: If a name is missing, or `--fields` names an unknown field, then
  Sauron shall reject the request without describing.
- **FR-005**: If no managed artifact of that type and name exists, then Sauron
  shall report that the artifact is not installed and exit with an error.
- **FR-006**: If the track file cannot be read or parsed, then Sauron shall reject
  the request and report that it cannot be read.

## Key Entities

- **Installed Artifact**: a delivered artifact recorded in the
  [track file](../0006-sync-artifacts/spec.md) — its `type`, `name`, source
  `registry`, `provider`, installed `path`, `pinned` state, and (when set)
  `persona`.

## Notes

- Describe covers **managed** (installed) artifacts only; to inspect what a
  registry offers, use [`list ... --available`](../0021-list-artifacts/spec.md).

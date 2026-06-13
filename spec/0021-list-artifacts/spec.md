# List Artifacts

**Type:** feature

**Depends on:** [sync artifacts](../0006-sync-artifacts/spec.md)

## Overview

A person responsible for a team's agentic-AI setup needs to review the managed
skills and agents — what is installed and from which registry — and, on demand,
what a registry offers or what the resolved catalog looks like, so that the
delivered set can be audited.

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to list the managed skills and the
  managed agents.

### Event-driven

- **FR-002**: When `list skills` or `list agents` runs, Sauron shall list the
  installed artifacts of that type recorded in the track file, each with its name
  and source registry.
- **FR-003**: When `--registry <name>` is given without `--available`, Sauron
  shall list only managed artifacts whose source registry is that registry.
- **FR-004**: When `--available --registry <name>` is given, Sauron shall list
  the artifacts of that type the named registry offers, read live from the
  registry.
- **FR-005**: When `--available` is given without `--registry`, Sauron shall list
  the resolved catalog — for each name the winning registry under pin then
  priority, scoped to the installed personas (or all artifacts when no personas
  are installed).
- **FR-006**: When listing, Sauron shall include each artifact's source registry
  and pinned state, and display the columns named by `--fields` in order with the
  name first; the valid fields include `source` (the source registry) and
  `pinned`.

### Unwanted behavior

- **FR-007**: If `--sort`, `--order`, or `--fields` is given an unsupported
  value, then Sauron shall reject the request without listing.
- **FR-008**: If the track file cannot be read or parsed, then Sauron shall reject
  the request and report that it cannot be read.
- **FR-009**: If `--available` is requested and the backing registry cannot be
  reached, then Sauron shall report the failure and exit with an error.

### Optional

- **FR-010**: Where `--search <term>` is provided, Sauron shall filter
  case-insensitively on the artifact name and source registry.
- **FR-011**: Where `--sort <field>` is provided, Sauron shall sort by `name`
  (default), `registry`, or `type`; `--order` selects `asc` (default) or `desc`.
- **FR-012**: Where no artifact matches (an empty managed set or no matches),
  Sauron shall report that there is nothing to list and exit successfully.

## Key Entities

- **Installed Artifact**: a delivered artifact recorded in the
  [track file](../0006-sync-artifacts/spec.md) — its `type`, `name`, source
  `registry`, `provider`, installed `path`, `pinned` state, and (when set)
  `persona`.
- **Available Artifact**: an artifact a registry offers (live), or the resolved
  catalog entry — the winning registry for a name under pin then priority. Not
  persisted; see the
  [live resolution](../0006-sync-artifacts/architecture/ADR-0001-conflict-resolution-by-registry-priority.md).

## Notes

- `list skills`/`list agents` show managed artifacts; `--available` switches to
  what registries offer. The resolved `--available` view (no `--registry`)
  overlaps with [`sync artifacts --dry-run`](../0006-sync-artifacts/spec.md): this
  is a read of the available/winning set, whereas the dry run is the diff against
  what is installed.

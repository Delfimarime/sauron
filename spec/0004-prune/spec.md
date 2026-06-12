# Prune

**Type:** feature

**Depends on:** [sync artifacts](../0006-sync-artifacts/spec.md)

## Overview

A person responsible for a team's agentic-AI setup needs to remove artifacts
left behind by registries Sauron no longer tracks, so that the team's provider
carries only artifacts from currently registered registries.

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to prune installed artifacts
  whose source registry is no longer registered with Sauron.

### Event-driven

- **FR-002**: When a user runs `prune artifacts`, Sauron shall consider both
  skills and agents.
- **FR-003**: When a user runs `prune skills` or `prune agents`, Sauron shall
  consider only that type.
- **FR-004**: When pruning, Sauron shall identify installed artifacts,
  recorded in the track file (`~/.sauron/track.yaml`), whose source registry
  is not among the currently registered registries.
- **FR-005**: When an orphaned artifact is identified, Sauron shall delete it
  from its provider location and remove its entry from the track file.
- **FR-006**: When pruning completes, Sauron shall report what was removed,
  grouped by skills and agents with a `-` for each removed artifact (the same
  plan format as [sync artifacts](../0006-sync-artifacts/spec.md)).

### State-driven

- **FR-007**: While pruning, Sauron shall leave artifacts whose source
  registry is still registered untouched.

### Unwanted behavior

- **FR-008**: If a noun other than `artifacts`, `skills`, or `agents` is given,
  or none is given, then Sauron shall reject the request and report the allowed
  nouns.
- **FR-009**: If no orphaned artifacts are found, then Sauron shall report
  that there is nothing to prune and exit successfully.
- **FR-010**: If `registries.yaml` or the track file cannot be read, then Sauron
  shall reject the request and report that it cannot be read.
- **FR-011**: If an orphaned artifact cannot be deleted, then Sauron shall
  report the failure, continue pruning the remainder, and exit with an error.

### Optional

- **FR-012**: Where `--dry-run` is provided, Sauron shall print the plan
  without changing the environment or the track file.

## Key Entities

- **Installed Artifact**: an artifact that Sauron has delivered to a provider,
  recorded in the track file (`~/.sauron/track.yaml`). It has a type (skill or
  agent), a provider, an installed path, a source registry (its provenance),
  and optionally the persona that brought it into scope.
- **Orphaned Artifact**: an installed artifact whose source registry is no
  longer among the registered registries. Prune's subject.

## Decision Records

- [Prune scope and provenance](architecture/ADR-0001-prune-scope-and-provenance.md)
  — prune providers artifacts from unregistered registries only, using
  provenance recorded in the track file.

## Notes

- The track file (`~/.sauron/track.yaml`) is populated by
  [sync artifacts](../0006-sync-artifacts/spec.md); prune reads it and removes entries for the
  artifacts it prunes. See [data/configuration.md](data/configuration.md).
- Configuration is now split across files per the
  [configuration data contract](../contracts/configuration.md); file
  references updated accordingly.

# Delete Artifacts

**Type:** feature
**Depends on:** [sync artifacts](../0006-sync-artifacts/spec.md)

## Overview

A person responsible for a team's agentic-AI setup needs to remove every
artifact Sauron has installed, so that the environment can be reset to a clean
state.

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to erase the artifacts it
  manages.

### Event-driven

- **FR-002**: When a user runs `delete artifacts`, Sauron shall consider both
  skill and agent artifacts recorded in the track file (`~/.sauron/track.yaml`),
  across all providers.
- **FR-003**: When an artifact is in scope, Sauron shall delete it from its
  installed path and remove its entry from the track file.
- **FR-004**: When `delete artifacts` completes, Sauron shall report what was
  removed, grouped by skills and agents.
- **FR-012**: When a user runs `delete skills` or `delete agents`, Sauron shall
  consider only that artifact type.

### State-driven

- **FR-005**: While deleting artifacts, Sauron shall only remove artifacts it
  has recorded in the track file; artifacts it does not track are never touched.

### Unwanted behavior

- **FR-006**: If `--persona` is given without a value, then Sauron shall
  reject the request and report that a persona name is required.
- **FR-007**: If no artifacts are in scope, then Sauron shall report that
  there is nothing to delete and exit successfully.
- **FR-008**: If the track file cannot be read or parsed, then Sauron shall
  reject the request and report that it cannot be read.
- **FR-009**: If an artifact cannot be deleted, then Sauron shall report the
  failure, continue with the remainder, and exit with an error.
- **FR-013**: If a noun other than `artifacts`, `skills`, or `agents` is given,
  or none is given, then Sauron shall reject the request and report the allowed
  nouns.

### Optional

- **FR-010**: Where `--persona` is provided, Sauron shall consider only the
  artifacts recorded with that persona, regardless of whether the persona is
  still defined.
- **FR-011**: Where `--dry-run` is provided, Sauron shall print the plan
  without changing the environment or the track file.

## Key Entities

- **Installed Artifact**: an artifact recorded in the track file
  (`~/.sauron/track.yaml`). `delete artifacts` deletes it from its path and
  removes its entry. Unlike [prune](../0004-prune/spec.md) — which removes only
  artifacts whose source registry is unregistered — `delete artifacts` removes
  everything in scope regardless of whether the source registry still exists.

## Notes

- `--persona` matches the `persona` field recorded in the track file, so
  artifacts left by an already-deleted persona can still be deleted.
- `delete artifacts` differs from a registry teardown:
  [unset backend](../0012-backend/spec.md) performs a
  registry-scoped teardown that, by default, also removes the delivered
  artifacts (unless `--keep-artifacts`), whereas `delete artifacts` removes all
  installed artifacts in scope regardless of their source.

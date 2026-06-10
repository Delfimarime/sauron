# Clear

**Type:** feature
**Depends on:** [sync](../0009-sync/spec.md)

## Overview

A person responsible for a team's agentic-AI setup needs to remove every
artifact Sauron has installed, so that the environment can be reset to a clean
state.

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to erase the artifacts it
  manages.

### Event-driven

- **FR-002**: When clear runs without `--persona`, Sauron shall consider every
  artifact recorded in the track file (`~/.sauron/track.yaml`), across all
  targets.
- **FR-003**: When an artifact is in scope, Sauron shall delete it from its
  installed path and remove its entry from the track file.
- **FR-004**: When clear completes, Sauron shall report what was removed,
  grouped by skills and agents.

### State-driven

- **FR-005**: While clearing, Sauron shall only remove artifacts it has
  recorded in the track file; artifacts it does not track are never touched.

### Unwanted behavior

- **FR-006**: If `--persona` is given without a value, then Sauron shall
  reject the request and report that a persona name is required.
- **FR-007**: If no artifacts are in scope, then Sauron shall report that
  there is nothing to clear and exit successfully.
- **FR-008**: If the track file cannot be read or parsed, then Sauron shall
  reject the request and report that it cannot be read.
- **FR-009**: If an artifact cannot be deleted, then Sauron shall report the
  failure, continue with the remainder, and exit with an error.

### Optional

- **FR-010**: Where `--persona` is provided, Sauron shall consider only the
  artifacts recorded with that persona, regardless of whether the persona is
  still defined.
- **FR-011**: Where `--dry-run` is provided, Sauron shall print the plan
  without changing the environment or the track file.

## Key Entities

- **Installed Artifact**: an artifact recorded in the track file
  (`~/.sauron/track.yaml`). Clear deletes it from its path and removes its
  entry. Unlike [prune](../0004-prune/spec.md) — which removes only artifacts
  whose source repository is unregistered — clear removes everything in scope
  regardless of whether the source repository still exists.

## Notes

- `--persona` matches the `persona` field recorded in the track file, so
  artifacts left by an already-deleted persona can still be cleared.

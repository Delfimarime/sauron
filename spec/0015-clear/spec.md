# Feature Specification: Clear

**Created**: 2026-06-10

**Status**: Draft

**Input**: "Erase every agent and skill Sauron manages, optionally scoped to a persona."

## Overview

A person responsible for a team's agentic-AI setup needs to remove every skill and agent Sauron has installed, so that the environment can be reset to a clean state.

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to erase the skills and agents it manages.

### Event-driven (*When*)

- **FR-002**: When clear runs without `--persona`, Sauron shall consider every artifact recorded in `~/.sauron/track.yaml`, across all targets.
- **FR-003**: When clear runs with `--persona`, Sauron shall consider only the artifacts recorded with that persona, regardless of whether the persona is still defined.
- **FR-004**: When an artifact is in scope, Sauron shall delete it from its installed path and remove its entry from `~/.sauron/track.yaml`.
- **FR-005**: When clear completes, Sauron shall report what was removed, grouped by skills and agents.
- **FR-006**: When `--dry-run` is provided, Sauron shall report what would be removed without deleting anything or modifying `~/.sauron/track.yaml`.

### State-driven (*While*)

- **FR-007**: While clearing, Sauron shall only remove artifacts it has recorded in `~/.sauron/track.yaml`; artifacts it does not track are never touched.

### Unwanted-behavior (*If / then*)

- **FR-008**: If `--persona` is given without a value, then Sauron shall reject the request and report that a persona name is required.
- **FR-009**: If no artifacts are in scope, then Sauron shall report that there is nothing to clear and succeed.
- **FR-010**: If `~/.sauron/track.yaml` cannot be read or parsed, then Sauron shall reject the request and report that it cannot be read.
- **FR-011**: If an artifact cannot be deleted, then Sauron shall report the failure, continue with the remainder, and exit with an error.

## Key Entities

- **Installed Artifact**: a skill or agent recorded in `~/.sauron/track.yaml`. Clear deletes it from its path and removes its entry. Unlike prune — which removes only artifacts whose source repository is unregistered — clear removes everything in scope regardless of whether the source repository still exists.

## Notes

- `--persona` matches the `persona` field recorded in `~/.sauron/track.yaml`, so artifacts left by an already-deleted persona can still be cleared.

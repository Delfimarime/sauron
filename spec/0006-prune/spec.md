# Feature Specification: Prune Orphaned Skills and Agents

**Created**: 2026-06-09

**Status**: Draft

**Input**: "Delete skills and/or agents that belong to repositories that are no longer part of Sauron's registered repositories."

## Overview

A person responsible for a team's agentic-AI setup needs to remove skills and agents left behind by repositories Sauron no longer tracks, so that the team's targets carry only artifacts from currently registered sources.

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to prune installed skills and agents whose source repository is no longer registered with Sauron.

### Event-driven (*When*)

- **FR-002**: When the user runs prune without a type, Sauron shall consider both skills and agents.
- **FR-003**: When the user runs prune with `skills` or `agents`, Sauron shall consider only that type.
- **FR-004**: When pruning, Sauron shall identify installed artifacts, recorded in `~/.sauron/track.yaml`, whose source repository is not among the currently registered repositories.
- **FR-005**: When an orphaned artifact is identified, Sauron shall delete it from its target location and remove its entry from `~/.sauron/track.yaml`.
- **FR-006**: When pruning completes, Sauron shall report what was removed (the count and the names).
- **FR-007**: When `--dry-run` is provided, Sauron shall report what would be removed without deleting anything or modifying `~/.sauron/track.yaml`.

### State-driven (*While*)

- **FR-008**: While pruning, Sauron shall leave artifacts whose source repository is still registered untouched.

### Unwanted-behavior (*If / then*)

- **FR-009**: If a type other than `skills` or `agents` is given, then Sauron shall reject the request and report the allowed types.
- **FR-010**: If no orphaned artifacts are found, then Sauron shall report that there is nothing to prune and succeed.
- **FR-011**: If `~/.sauron/settings.yaml` or `~/.sauron/track.yaml` cannot be read, then Sauron shall reject the request and report that it cannot be read.
- **FR-012**: If an orphaned artifact cannot be deleted, then Sauron shall report the failure, continue pruning the remainder, and exit with an error.

## Key Entities

- **Installed Artifact**: a skill or agent that Sauron has delivered to a target, recorded in `~/.sauron/track.yaml`. It has a type (skill or agent), a target location, and a source repository (its provenance).
- **Orphaned Artifact**: an installed artifact whose source repository is no longer among the registered repositories. Prune's subject.

## Notes

- `~/.sauron/track.yaml` is populated by the deliver/install feature; prune reads it and removes entries for the artifacts it prunes. See `data/configuration.md`.

## Decision Records

- `architecture/ADR-0001-prune-scope-and-provenance.md` — prune targets artifacts from unregistered repositories only, using provenance recorded in `track.yaml`.

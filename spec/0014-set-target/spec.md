# Feature Specification: Set Target

**Created**: 2026-06-10

**Status**: Draft

**Input**: "Allow a user to set the target provider for Sauron; changing it moves all agents and skills to the new target unless --copy-only is given."

## Overview

A person responsible for a team's agentic-AI setup needs to choose which provider Sauron delivers to, so that the team's skills and agents land where their tooling expects them.

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to set the active target provider.
- **FR-002**: Sauron shall maintain a single active target in its configuration, defaulting to `claude` when never set. (See ADR-0001.)

### Event-driven (*When*)

- **FR-003**: When the user sets a target, Sauron shall require a target value of `claude` or `zencoder`.
- **FR-004**: When the new target differs from the current one, Sauron shall relocate every installed artifact recorded on the previous target to the new target's locations, and update its entry in `~/.sauron/track.yaml`.
- **FR-005**: When relocating without `--copy-only`, Sauron shall install each artifact at the new target and delete it from the previous target.
- **FR-006**: When relocating with `--copy-only`, Sauron shall install each artifact at the new target and leave the previous target's artifacts in place, recording the new target as an additional `~/.sauron/track.yaml` entry.
- **FR-007**: When the target is changed, Sauron shall persist the new active target to its configuration.
- **FR-008**: When the operation completes, Sauron shall report the new target and what was moved or copied.

### State-driven (*While*)

- **FR-009**: While setting the target, Sauron shall leave the existing configuration and tracking record unchanged until the change is persisted; both files are left untouched on any failure.

### Unwanted-behavior (*If / then*)

- **FR-010**: If no target value is provided, then Sauron shall reject the request and report that a target is required.
- **FR-011**: If the target value is not `claude` or `zencoder`, then Sauron shall reject the request and report the supported targets.
- **FR-012**: If the new target equals the current target, then Sauron shall make no change and report that the target is already set (no-op).
- **FR-013**: If `~/.sauron/settings.yaml` or `~/.sauron/track.yaml` cannot be read or parsed, then Sauron shall reject the request and report that it cannot be read.
- **FR-014**: If an artifact cannot be moved or copied, then Sauron shall report the failure, continue with the remainder, and exit with an error.

## Key Entities

- **Target**: the single active provider destination — `claude` (default) or `zencoder` — stored in `~/.sauron/settings.yaml` and used by sync, cron, and migration. Each target defines where skills and agents are persisted.
- **Installed Artifact**: a delivered skill or agent recorded in `~/.sauron/track.yaml`; its `target` and `path` are rewritten (move) or duplicated (`--copy-only`) when the active target changes.

## Decision Records

- `architecture/ADR-0001-global-target.md` — the target is a single global setting, defaulting to `claude`; changing it migrates installed artifacts.

# Set Target

**Type:** feature

## Overview

A person responsible for a team's agentic-AI setup needs to choose which
provider Sauron delivers to, so that the team's artifacts land where their
tooling expects them. The target is a single global setting (see
[ADR-0001](architecture/ADR-0001-global-target.md)); changing it migrates
installed artifacts to the new target's locations.

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to set the active target
  provider.
- **FR-002**: Sauron shall maintain a single active target in the settings,
  defaulting to `claude` when never set (see
  [ADR-0001](architecture/ADR-0001-global-target.md)).

### Event-driven

- **FR-003**: When a user sets a target, Sauron shall require a target value
  of `claude` or `zencoder`.
- **FR-004**: When the new target differs from the current one, Sauron shall
  relocate every installed artifact recorded on the previous target to the new
  target's locations, and update its entry in the track file
  (`~/.sauron/track.yaml`).
- **FR-005**: When relocating without `--copy-only`, Sauron shall install each
  artifact at the new target and delete it from the previous target.
- **FR-006**: When the target is changed, Sauron shall persist the new active
  target to the settings.
- **FR-007**: When the operation completes, Sauron shall report the new target
  and what was moved or copied.

### State-driven

- **FR-008**: While setting the target, Sauron shall leave the existing
  configuration and the track file unchanged until the change is persisted;
  both files are left untouched on any failure.

### Unwanted behavior

- **FR-009**: If no target value is provided, then Sauron shall reject the
  request and report that a target is required.
- **FR-010**: If the target value is not `claude` or `zencoder`, then Sauron
  shall reject the request and report the supported targets.
- **FR-011**: If the new target equals the current target, then Sauron shall
  make no change and report that the target is already set (no-op).
- **FR-012**: If the settings or the track file cannot be read or parsed, then
  Sauron shall reject the request and report that it cannot be read.
- **FR-013**: If an artifact cannot be moved or copied, then Sauron shall
  report the failure, continue with the remainder, and exit with an error.

### Optional

- **FR-014**: Where `--copy-only` is provided, Sauron shall, when relocating,
  install each artifact at the new target and leave the previous target's
  artifacts in place, recording the new target as an additional track file
  entry.

## Key Entities

- **Target**: the single active provider destination — `claude` (default) or
  `zencoder` — stored in the settings (`~/.sauron/settings.yaml`) and used by
  [sync](../0009-sync/spec.md), [cron sync](../0014-cron-sync/spec.md), and
  migration. Each target defines where skills and agents are persisted.
- **Installed Artifact**: a delivered artifact recorded in the track file
  (`~/.sauron/track.yaml`); its `target` and `path` are rewritten (move) or
  duplicated (`--copy-only`) when the active target changes.

## Decision Records

- [Global target](architecture/ADR-0001-global-target.md) — the target is a
  single global setting, defaulting to `claude`; changing it migrates
  installed artifacts.

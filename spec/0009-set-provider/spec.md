# Set Provider

**Type:** feature

## Overview

A person responsible for a team's agentic-AI setup needs to choose which
provider Sauron delivers to, so that the team's artifacts land where their
tooling expects them. The provider is a single global setting (see
[ADR-0001](architecture/ADR-0001-global-provider.md)); changing it migrates
installed artifacts to the new provider's locations.

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to set the active provider
  provider.
- **FR-002**: Sauron shall maintain a single active provider in the settings,
  defaulting to `claude` when never set (see
  [ADR-0001](architecture/ADR-0001-global-provider.md)).

### Event-driven

- **FR-003**: When a user sets a provider, Sauron shall require a provider value
  of `claude` or `zencoder`.
- **FR-004**: When the new provider differs from the current one, Sauron shall
  relocate every installed artifact recorded on the previous provider to the new
  provider's locations, and update its entry in the track file
  (`~/.sauron/track.yaml`).
- **FR-005**: When relocating without `--copy-only`, Sauron shall install each
  artifact at the new provider and delete it from the previous provider.
- **FR-006**: When the provider is changed, Sauron shall persist the new active
  provider to the settings.
- **FR-007**: When the operation completes, Sauron shall report the new provider
  and what was moved or copied.

### State-driven

- **FR-008**: While setting the provider, Sauron shall leave the existing
  configuration and the track file unchanged until the change is persisted;
  both files are left untouched on any failure.

### Unwanted behavior

- **FR-009**: If no provider value is provided, then Sauron shall reject the
  request and report that a provider is required.
- **FR-010**: If the provider value is not `claude` or `zencoder`, then Sauron
  shall reject the request and report the supported providers.
- **FR-011**: If the new provider equals the current provider, then Sauron shall
  make no change and report that the provider is already set (no-op).
- **FR-012**: If the settings or the track file cannot be read or parsed, then
  Sauron shall reject the request and report that it cannot be read.
- **FR-013**: If an artifact cannot be moved or copied, then Sauron shall
  report the failure, continue with the remainder, and exit with an error.

### Optional

- **FR-014**: Where `--copy-only` is provided, Sauron shall, when relocating,
  install each artifact at the new provider and leave the previous provider's
  artifacts in place, recording the new provider as an additional track file
  entry.

## Key Entities

- **Provider**: the single active provider destination — `claude` (default) or
  `zencoder` — stored in the settings (`~/.sauron/settings.yaml`) and used by
  [sync](../0006-sync-artifacts/spec.md),
  [scheduled sync](../0011-schedule-sync/spec.md), and migration. Each provider
  defines where skills and agents are persisted.
- **Installed Artifact**: a delivered artifact recorded in the track file
  (`~/.sauron/track.yaml`); its `provider` and `path` are rewritten (move) or
  duplicated (`--copy-only`) when the active provider changes.

## Decision Records

- [Global provider](architecture/ADR-0001-global-provider.md) — the provider is a
  single global setting, defaulting to `claude`; changing it migrates
  installed artifacts.

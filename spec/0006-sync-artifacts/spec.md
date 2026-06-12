# Sync Artifacts

**Type:** feature

**Depends on:** [add registry](../0001-add-registry/spec.md) (kind behavior: [http](../0001-add-registry/capabilities/http.md), [filesystem](../0001-add-registry/capabilities/filesystem.md), [git](../0001-add-registry/capabilities/git.md)), [set personas](../0014-select-personas/spec.md), [set provider](../0009-set-provider/spec.md)

## Overview

A person responsible for a team's agentic-AI setup needs to synchronize the
artifacts in the current environment with the registered registries, so that
the active provider carries exactly the artifacts the team's installed personas
call for, at their latest versions.

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to synchronize artifacts from
  the registered registries to the active provider environment.

### Event-driven

- **FR-002**: When sync artifacts runs without `--persona` and at least one
  persona is installed, Sauron shall use the union of all installed personas'
  artifacts as the desired set.
- **FR-003**: When sync artifacts runs without `--persona` and no personas are
  installed, Sauron shall use all artifacts from all registered registries as
  the desired set.
- **FR-004**: When the same artifact name is provided by more than one
  registry and is not pinned, Sauron shall take it from the registry with the
  higher precedence (lower priority value) (see
  [ADR-0001](architecture/ADR-0001-conflict-resolution-by-registry-priority.md)).
- **FR-018**: When an artifact is pinned to a registry (its track entry's
  `pinned` is `true`, set by [pin artifact](../0020-pin-artifact/spec.md)), Sauron
  shall take it from the pinned registry, overriding priority-based conflict
  resolution (see
  [ADR-0001](architecture/ADR-0001-conflict-resolution-by-registry-priority.md)).
- **FR-005**: When computing changes, Sauron shall compare the desired set
  against the artifacts recorded for the active provider in the track file
  (`~/.sauron/track.yaml`) and produce a plan of additions/updates and
  removals.
- **FR-006**: When sync artifacts applies the plan, Sauron shall install or
  update each artifact at the active provider's location, remove tracked
  artifacts no longer in the desired set, and record each installed artifact in
  the track file with its source registry, the installed persona that brought
  it into the desired set (when personas are in play; when several do, the
  highest-precedence installed persona per
  [priority model](../AUTHORING.md#priority-model)),
  the provider, and the installed path.
- **FR-007**: When sync artifacts runs, Sauron shall deliver to the configured
  global provider (`claude` by default; set with
  [set provider](../0009-set-provider/spec.md)).
- **FR-008**: When sync artifacts completes, Sauron shall report a summary of
  what was added, updated, and removed.

### State-driven

- **FR-009**: While sync artifacts runs, Sauron shall only remove artifacts it
  has recorded in the track file; artifacts it does not track are never
  touched.

### Unwanted behavior

- **FR-010**: If `--persona` names a persona that does not exist, then Sauron
  shall reject the request and report that the persona is not found.
- **FR-011**: If the configuration or the track file cannot be read or parsed,
  then Sauron shall reject the request and report that it cannot be read.
- **FR-012**: If a registry cannot be reached during sync artifacts, then
  Sauron shall report the failure, continue with the remaining registries,
  and exit with an error.
- **FR-013**: If an artifact cannot be installed or removed, then Sauron shall
  report the failure, continue with the remainder, and exit with an error.
- **FR-014**: If the desired set references an artifact that no registered
  registry provides, then Sauron shall report the missing artifact, continue
  with the remainder, and exit with an error.
- **FR-015**: If the desired set already matches the tracked state for the
  active provider, then Sauron shall report that the provider is up to date and
  exit successfully.

### Optional

- **FR-016**: Where `--persona` is provided, Sauron shall scope the desired
  set to that persona's artifacts.
- **FR-017**: Where `--dry-run` is provided, Sauron shall print the plan
  without changing the environment or the track file.

## Key Entities

- **Desired Set**: the artifacts that should be present on the provider — one
  persona's artifacts, the union of all installed personas' artifacts, or
  everything from all registries (FR-002, FR-003, FR-016), with name
  conflicts resolved by registry priority (FR-004).
- **Plan**: the difference between the desired set and the tracked state —
  additions/updates (`+`) and removals (`-`), grouped by skills and agents.
- **Provider**: the active provider environment artifacts are delivered to,
  configured globally (`claude` by default; see
  [set provider](../0009-set-provider/spec.md)). Each provider defines where skills
  and agents are persisted.
- **Installed Artifact**: a delivered artifact recorded in the track file
  (`~/.sauron/track.yaml`) with its type, name, provider, installed path, source
  registry, and (when applicable) installed persona.

## Decision Records

- [Conflict resolution — pin, then registry priority](architecture/ADR-0001-conflict-resolution-by-registry-priority.md)
  — a pinned artifact is taken from its pinned registry; otherwise the
  lowest-priority-value registry wins.

## Notes

- `sync artifacts` (this command) and
  [sync personas](../0013-sync-personas/spec.md) are independent operations.
  `sync personas` refreshes the stored *definitions* of the installed personas
  from the backend; `sync artifacts` reconciles the active provider with the
  artifacts the *installed* personas call for. Installing a persona is owned by
  [set personas](../0014-select-personas/spec.md).
- The recommended order is therefore
  [sync personas](../0013-sync-personas/spec.md) →
  [set personas](../0014-select-personas/spec.md) → `sync artifacts`: pull
  the latest persona definitions, install the ones that should participate, then
  reconcile the provider with their artifacts.
- Configuration is now split across files per the
  [configuration data contract](../contracts/configuration.md); file
  references updated accordingly.

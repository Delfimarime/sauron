# Sync

**Type:** feature
**Depends on:** [add repository](../0001-add-repository/spec.md) (kind behavior: [http](../0001-add-repository/capabilities/http.md), [filesystem](../0001-add-repository/capabilities/filesystem.md), [git](../0001-add-repository/capabilities/git.md)), [import persona](../0005-import-persona/spec.md), [set target](../0012-set-target/spec.md)

## Overview

A person responsible for a team's agentic-AI setup needs to synchronize the
artifacts in the current environment with the registered repositories, so that
the active target carries exactly the artifacts the team's personas call for,
at their latest versions.

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to synchronize artifacts from
  the registered repositories to the active target environment.

### Event-driven

- **FR-002**: When sync runs without `--persona` and at least one persona is
  defined, Sauron shall use the union of all personas' artifacts as the
  desired set.
- **FR-003**: When sync runs without `--persona` and no personas are defined,
  Sauron shall use all artifacts from all registered repositories as the
  desired set.
- **FR-004**: When the same artifact name is provided by more than one
  repository, Sauron shall take it from the repository with the higher
  precedence (lower priority value) (see
  [ADR-0001](architecture/ADR-0001-conflict-resolution-by-repository-priority.md)).
- **FR-005**: When computing changes, Sauron shall compare the desired set
  against the artifacts recorded for the active target in the track file
  (`~/.sauron/track.yaml`) and produce a plan of additions/updates and
  removals.
- **FR-006**: When sync applies the plan, Sauron shall install or update each
  artifact at the active target's location, remove tracked artifacts no longer
  in the desired set, and record each installed artifact in the track file
  with its source repository, the persona that brought it into the desired set
  (when personas are in play; when several do, the highest-precedence persona
  per
  [import persona ADR-0002](../0005-import-persona/architecture/ADR-0002-unified-priority-model.md)),
  the target, and the installed path.
- **FR-007**: When sync runs, Sauron shall deliver to the configured global
  target (`claude` by default; set with
  [set target](../0012-set-target/spec.md)).
- **FR-008**: When sync completes, Sauron shall report a summary of what was
  added, updated, and removed.

### State-driven

- **FR-009**: While sync runs, Sauron shall only remove artifacts it has
  recorded in the track file; artifacts it does not track are never touched.

### Unwanted behavior

- **FR-010**: If `--persona` names a persona that does not exist, then Sauron
  shall reject the request and report that the persona is not found.
- **FR-011**: If the settings or the track file cannot be read or parsed, then
  Sauron shall reject the request and report that it cannot be read.
- **FR-012**: If a repository cannot be reached during sync, then Sauron shall
  report the failure, continue with the remaining repositories, and exit with
  an error.
- **FR-013**: If an artifact cannot be installed or removed, then Sauron shall
  report the failure, continue with the remainder, and exit with an error.
- **FR-014**: If the desired set references an artifact that no registered
  repository provides, then Sauron shall report the missing artifact, continue
  with the remainder, and exit with an error.
- **FR-015**: If the desired set already matches the tracked state for the
  active target, then Sauron shall report that the target is up to date and
  exit successfully.

### Optional

- **FR-016**: Where `--persona` is provided, Sauron shall scope the desired
  set to that persona's artifacts.
- **FR-017**: Where `--dry-run` is provided, Sauron shall print the plan
  without changing the environment or the track file.

## Key Entities

- **Desired Set**: the artifacts that should be present on the target — one
  persona's artifacts, the union of all personas' artifacts, or everything
  from all repositories (FR-002, FR-003, FR-016), with name conflicts resolved
  by repository priority (FR-004).
- **Plan**: the difference between the desired set and the tracked state —
  additions/updates (`+`) and removals (`-`), grouped by skills and agents.
- **Target**: the active provider environment artifacts are delivered to,
  configured globally (`claude` by default; see
  [set target](../0012-set-target/spec.md)). Each target defines where skills
  and agents are persisted.
- **Installed Artifact**: a delivered artifact recorded in the track file
  (`~/.sauron/track.yaml`) with its type, name, target, installed path, source
  repository, and (when applicable) persona.

## Decision Records

- [Conflict resolution by repository priority](architecture/ADR-0001-conflict-resolution-by-repository-priority.md)
  — same-named artifacts resolve by repository priority.

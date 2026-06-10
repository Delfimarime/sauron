# Feature Specification: Sync

**Created**: 2026-06-10

**Status**: Draft

**Input**: "Synchronize skills and agents between the registered repositories and the current environment, with a dry-run plan and optional persona scoping."

## Overview

A person responsible for a team's agentic-AI setup needs to synchronize the skills and agents in the current environment with the registered repositories, so that the active target carries exactly the artifacts the team's personas call for, at their latest versions.

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to synchronize skills and agents from the registered repositories to the active target environment.

### Event-driven (*When*)

- **FR-002**: When sync runs with `--persona`, Sauron shall scope the desired set to that persona's skills and agents.
- **FR-003**: When sync runs without `--persona` and at least one persona is defined, Sauron shall use the union of all personas' skills and agents as the desired set.
- **FR-004**: When sync runs without `--persona` and no personas are defined, Sauron shall use all skills and agents from all registered repositories as the desired set.
- **FR-005**: When the same artifact name is provided by more than one repository, Sauron shall take it from the repository with the higher precedence (lower priority value). (See ADR-0001.)
- **FR-006**: When computing changes, Sauron shall compare the desired set against the artifacts recorded for the active target in `~/.sauron/track.yaml` and produce a plan of additions/updates and removals.
- **FR-007**: When sync applies the plan, Sauron shall install or update each artifact at the active target's location, remove tracked artifacts no longer in the desired set, and record each installed artifact in `~/.sauron/track.yaml` with its source repository, the persona that brought it into the desired set (when personas are in play; when several do, the highest-precedence persona per `0007-import-persona` ADR-0001), the target, and the installed path.
- **FR-008**: When `--dry-run` is provided, Sauron shall print the plan — grouped by skills and agents, `+` for additions/updates and `-` for removals — without changing the environment or `~/.sauron/track.yaml`.
- **FR-009**: When sync runs, Sauron shall deliver to the configured global target (`claude` by default; set with `0014-set-target`).
- **FR-010**: When sync completes, Sauron shall report a summary of what was added, updated, and removed.

### State-driven (*While*)

- **FR-011**: While sync runs, Sauron shall only remove artifacts it has recorded in `~/.sauron/track.yaml`; artifacts it does not track are never touched.

### Unwanted-behavior (*If / then*)

- **FR-012**: If `--persona` names a persona that does not exist, then Sauron shall reject the request and report that the persona is not found.
- **FR-013**: If `~/.sauron/settings.yaml` or `~/.sauron/track.yaml` cannot be read or parsed, then Sauron shall reject the request and report that it cannot be read.
- **FR-014**: If a repository cannot be reached during sync, then Sauron shall report the failure, continue with the remaining repositories, and exit with an error.
- **FR-015**: If an artifact cannot be installed or removed, then Sauron shall report the failure, continue with the remainder, and exit with an error.
- **FR-016**: If the desired set references a skill or agent that no registered repository provides, then Sauron shall report the missing artifact, continue with the remainder, and exit with an error.
- **FR-017**: If the desired set already matches the tracked state for the active target, then Sauron shall report that the target is up to date and succeed.

## Key Entities

- **Desired Set**: the skills and agents that should be present on the target — one persona's artifacts, the union of all personas' artifacts, or everything from all repositories (FR-002–FR-004), with name conflicts resolved by repository priority (FR-005).
- **Plan**: the difference between the desired set and the tracked state — additions/updates (`+`) and removals (`-`), grouped by skills and agents.
- **Target**: the active provider environment artifacts are delivered to, configured globally (`claude` by default; see `0014-set-target`). Each target defines where skills and agents are persisted.
- **Installed Artifact**: a delivered skill or agent recorded in `~/.sauron/track.yaml` with its type, name, target, installed path, source repository, and (when applicable) persona.

## Decision Records

- `architecture/ADR-0001-conflict-resolution-by-repository-priority.md` — same-named artifacts resolve by repository priority.

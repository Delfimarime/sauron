# Feature Specification: Add Filesystem Repository

**Created**: 2026-06-09

**Status**: Draft

**Input**: "Allow a user to register a filesystem directory repository as a source of skills and agents that Sauron can watch and deliver."

## Overview

A person responsible for a team's agentic-AI setup needs to register a filesystem directory as a source of skills and agents, so that Sauron can watch it and keep the team's targets in sync with its latest contents.

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to register a filesystem directory as a repository source of skills and agents.
- **FR-014**: Every registered repository shall have a name that is unique across all repositories, regardless of kind.
- **FR-015**: Every registered repository shall have a priority that is unique across all repositories, regardless of kind.

### Event-driven (*When*)

- **FR-002**: When a user submits a request to add a repository, Sauron shall require a repository kind, a name, a priority, and a directory path.
- **FR-003**: When a user submits a directory path, Sauron shall verify that the directory exists and is accessible before registering it.
- **FR-004**: When a user submits a directory path, Sauron shall verify that the directory contains at least one artifact under `.agents/` or `.skills/` before registering it.
- **FR-005**: When a repository is registered, Sauron shall identify it by its name.
- **FR-006**: When a repository passes validation, Sauron shall persist it to its configuration (`~/.sauron/settings.yaml`) so that it becomes a watched source in subsequent runs.
- **FR-007**: When a repository is successfully registered, Sauron shall report success to the user.
- **FR-016**: When a repository is registered, Sauron shall record its name and priority alongside its kind and path.

### State-driven (*While*)

- **FR-008**: While a repository is being validated, Sauron shall leave the existing configuration unchanged until validation succeeds.

### Unwanted-behavior (*If / then*)

- **FR-009**: If no directory path is provided, then Sauron shall reject the request and report that a path is required.
- **FR-010**: If `--kind filesystem` is not specified, then Sauron shall not treat the request as a filesystem repository; because `http` is the default kind, `filesystem` must be selected explicitly. (The `sauron add repository` interface is shared across kinds, each covered by its own feature.)
- **FR-011**: If the directory does not exist or cannot be accessed, then Sauron shall reject the request, leave the configuration unchanged, and report that the repository cannot be accessed.
- **FR-012**: If the directory contains neither a populated `.agents/` nor a populated `.skills/`, then Sauron shall reject the request and report that no skills or agents were found.
- **FR-013**: If a repository with the same path is already registered, then Sauron shall still register the new repository; duplicate paths are permitted and shall not, on their own, cause rejection.
- **FR-017**: If no name is provided, then Sauron shall reject the request and report that a name is required.
- **FR-018**: If the name is not a valid slug (`^[a-z0-9]+(-[a-z0-9]+)*$`), then Sauron shall reject the request and report that the name format is invalid.
- **FR-019**: If no priority is provided, or the priority is not a positive integer, then Sauron shall reject the request and report that a valid priority is required.
- **FR-021**: If the name is already used by another repository, then Sauron shall reject the request, leave the configuration unchanged, and report that the name must be unique.
- **FR-022**: If the priority is already used by another repository, then Sauron shall reject the request, leave the configuration unchanged, and report that the priority must be unique.

## Key Entities

- **Repository**: A registered filesystem directory that provides skills and/or agents, identified by its **name**. Holds artifacts under `.agents/` and `.skills/`. Each repository carries:
  - **name** — a unique slug (`^[a-z0-9]+(-[a-z0-9]+)*$`) that identifies the repository and is unique across all repositories regardless of kind.
  - **priority** — a unique positive integer, unique across all repositories regardless of kind, where a lower number means higher precedence.
  - **path** — the resolved absolute directory location. Paths are not required to be unique; two repositories may point at the same directory.

  Downstream, personas select from its artifacts and targets receive them.

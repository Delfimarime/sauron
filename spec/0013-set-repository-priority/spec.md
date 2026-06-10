# Feature Specification: Set Repository Priority

**Created**: 2026-06-10

**Status**: Draft

**Input**: "Allow a user to change a repository's priority by name."

## Overview

A person responsible for a team's agentic-AI setup needs to change a repository's priority, so that the precedence order among repositories reflects the team's intent.

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to set a repository's priority by name.

### Event-driven (*When*)

- **FR-002**: When the user sets a priority, Sauron shall require a repository name and a priority value.
- **FR-003**: When the request is valid, Sauron shall assign the value as the repository's priority and persist the updated configuration.
- **FR-004**: When the value equals the repository's current priority, Sauron shall make no change and report success (no-op).

### State-driven (*While*)

- **FR-005**: While setting a priority, Sauron shall leave the existing configuration unchanged until the change is persisted; the file is left untouched on any failure.

### Unwanted-behavior (*If / then*)

- **FR-006**: If the name or the value is missing, then Sauron shall reject the request and report what is required.
- **FR-007**: If the value is not a positive integer (`1` or greater), then Sauron shall reject the request and report that a positive integer is required.
- **FR-008**: If no repository with the given name exists, then Sauron shall reject the request and report that the repository is not found.
- **FR-009**: If the value is already used by another repository, then Sauron shall reject the request, leave the configuration unchanged, and report that the priority must be unique.
- **FR-010**: If the configuration cannot be read or parsed, then Sauron shall reject the request and report that the configuration cannot be read.

## Key Entities

- **Repository**: a registered source identified by its name. Its priority is a positive integer, unique across all repositories regardless of kind, where a lower value means higher precedence. Unlike persona priority, repository priority is always defined — there is no zero-anchor, no undefined value, and no single-repository guard; this feature is the only way to change it after the repository is added.

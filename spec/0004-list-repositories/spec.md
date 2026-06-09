# Feature Specification: List Repositories

**Created**: 2026-06-09

**Status**: Draft

**Input**: "Allow a user to list the repositories registered with Sauron, optionally filtered by a search term and sorted by an attribute."

## Overview

A person responsible for a team's agentic-AI setup needs to see which repositories Sauron is watching, so that they can review the team's configured sources of skills and agents.

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to list the registered repositories.

### Event-driven (*When*)

- **FR-002**: When the user lists repositories, Sauron shall read the registered repositories from its configuration and display each one's name, kind, priority, and location.
- **FR-003**: When `--search` is provided, Sauron shall include only repositories whose name or location contains the term, matched case-insensitively.
- **FR-004**: When `--sort` is provided, Sauron shall order the repositories by the chosen attribute — `name`, `priority`, or `kind`; when it is omitted, Sauron shall order by `priority`.
- **FR-005**: When `--order` is provided, Sauron shall order the repositories ascending or descending accordingly; when it is omitted, Sauron shall order ascending.
- **FR-006**: When displaying a repository's location, Sauron shall use the kind-appropriate locator field — `path` for filesystem, `url` for http, `uri` for git.

### Unwanted-behavior (*If / then*)

- **FR-007**: If no repositories are registered, then Sauron shall report that none are registered and succeed.
- **FR-008**: If `--search` matches no repositories, then Sauron shall report that none match and succeed.
- **FR-009**: If the configuration file does not exist, then Sauron shall treat it as no repositories registered.
- **FR-010**: If the configuration exists but cannot be read or parsed, then Sauron shall reject the request and report that the configuration cannot be read.
- **FR-011**: If `--sort` is not one of `name`, `priority`, or `kind`, then Sauron shall reject the request and report the allowed sort attributes.
- **FR-012**: If `--order` is not `asc` or `desc`, then Sauron shall reject the request and report the allowed order values.

## Key Entities

- **Repository**: a registered source of skills and agents, shown by its name, kind, priority, and location. The location is the kind's locator (`path`/`url`/`uri`). Listing spans all kinds.

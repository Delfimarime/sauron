# Feature Specification: List Personas

**Created**: 2026-06-10

**Status**: Draft

**Input**: "Allow a user to list the personas registered with Sauron, optionally filtered by a search term and tags, and sorted by an attribute."

## Overview

A person responsible for a team's agentic-AI setup needs to see which personas are defined, so that they can review how skills and agents are grouped for delivery.

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to list the registered personas.

### Event-driven (*When*)

- **FR-002**: When the user lists personas, Sauron shall read the registered personas from its configuration and display each one's name, priority, tags, and number of skills and agents.
- **FR-003**: When `--search` is provided, Sauron shall include only personas whose name or description contains the term, matched case-insensitively.
- **FR-004**: When `--tag` is provided (repeatable), Sauron shall include only personas that carry every given tag.
- **FR-005**: When both `--search` and `--tag` are provided, Sauron shall include only personas that satisfy both.
- **FR-006**: When `--sort` is provided, Sauron shall order the personas by the chosen attribute — `name` or `priority`; when it is omitted, Sauron shall order by `priority`.
- **FR-007**: When `--order` is provided, Sauron shall order the personas ascending or descending accordingly; when it is omitted, Sauron shall order ascending.
- **FR-008**: When ordering by priority, Sauron shall rank defined priorities first, ascending (`0` first), followed by personas with undefined priority ordered by name among themselves (per `0007-import-persona` ADR-0001); descending reverses that order.

### Unwanted-behavior (*If / then*)

- **FR-009**: If no personas are registered, then Sauron shall report that none are registered and succeed.
- **FR-010**: If the filters match no personas, then Sauron shall report that none match and succeed.
- **FR-011**: If the configuration file does not exist, then Sauron shall treat it as no personas registered.
- **FR-012**: If the configuration exists but cannot be read or parsed, then Sauron shall reject the request and report that the configuration cannot be read.
- **FR-013**: If `--sort` is not one of `name` or `priority`, then Sauron shall reject the request and report the allowed sort attributes.
- **FR-014**: If `--order` is not `asc` or `desc`, then Sauron shall reject the request and report the allowed order values.

## Key Entities

- **Persona**: a registered named set of agents and skills, shown by its name, priority (`-` when undefined), tags, and skill/agent counts.

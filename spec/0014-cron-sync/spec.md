# Feature Specification: Cron Sync

**Created**: 2026-06-10

**Status**: Draft

**Input**: "Schedule Sauron to run sync automatically via the operating system's cron; support disabling it."

## Overview

A person responsible for a team's agentic-AI setup needs Sauron to synchronize on a schedule without being run by hand, so that targets stay current automatically.

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to schedule `sauron sync` to run automatically via the operating system's cron. (See ADR-0001.)

### Event-driven (*When*)

- **FR-002**: When the user provides a cron expression, Sauron shall validate it and install or replace a managed crontab entry that runs `sauron sync`.
- **FR-003**: When the schedule is installed, Sauron shall record it in its configuration (`~/.sauron/settings.yaml`).
- **FR-004**: When the scheduled entry runs, it shall invoke `sauron sync` with no flags, so the sync follows the configured global target and the union of personas (or everything when no personas are defined).
- **FR-005**: When the schedule is installed or replaced, Sauron shall report the active schedule.
- **FR-006**: When `--disable` is provided, Sauron shall remove the managed crontab entry and the recorded schedule, and report that scheduling is disabled.

### State-driven (*While*)

- **FR-007**: While installing or removing the schedule, Sauron shall change only its own managed crontab entry, leaving any other crontab entries untouched.

### Unwanted-behavior (*If / then*)

- **FR-008**: If neither a cron expression nor `--disable` is provided, then Sauron shall reject the request and report that one is required.
- **FR-009**: If both a cron expression and `--disable` are provided, then Sauron shall reject the request and report that they are mutually exclusive.
- **FR-010**: If the cron expression is invalid, then Sauron shall reject the request and report that the expression is invalid.
- **FR-011**: If `--disable` is given when no schedule is installed, then Sauron shall make no change and report that scheduling is already disabled (treated as success).
- **FR-012**: If the crontab cannot be read or written, then Sauron shall reject the request and report that the schedule cannot be updated.

## Key Entities

- **Schedule**: the cron expression under which `sauron sync` runs, recorded in `~/.sauron/settings.yaml` and realized as a managed entry in the user's crontab.

## Decision Records

- `architecture/ADR-0001-cron-via-os-crontab.md` — scheduling uses the operating system's crontab with a managed marker.

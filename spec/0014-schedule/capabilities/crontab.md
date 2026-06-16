# Crontab Scheduling

**Type:** capability

**Enables:** [schedule](../spec.md)

## Overview

Scheduling is implemented through the operating system's crontab. This capability
manages the lifecycle of the managed crontab entries — adding, replacing, and
removing them — and keeps them in step with the `Schedule` documents in settings.

## Requirements

### Ubiquitous

- FR-001: Sauron shall manage its crontab entries through the OS crontab, marking
  each managed entry so it can be identified and removed without disturbing
  unmanaged entries.
- FR-002: Sauron shall keep the managed crontab entries and the `Schedule`
  documents consistent: registering or removing one updates the other.

### Event-driven

- FR-003: When a schedule for an operation is replaced, Sauron shall update the
  single managed crontab entry for that operation in place.

### Unwanted behavior

- FR-004: If the crontab cannot be read or written, then Sauron shall fail with a
  runtime error and leave existing entries unchanged.

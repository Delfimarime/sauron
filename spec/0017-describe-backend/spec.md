# Describe Backend

**Type:** feature
**Depends on:** [backend](../0012-backend/spec.md)

## Overview

A person responsible for a team's agentic-AI setup needs to inspect the single
configured backend in full — its kind, where it lives, how it authenticates, and
the state of the catalog pulled from it — so that they can review or troubleshoot
the persona source without re-reading the raw settings. The backend is the
**singleton** owned by [backend](../0012-backend/spec.md); describing it is the
single-record counterpart that reads that one record. Because there is exactly
one backend per instance, the command takes no name. It is read-only and offline:
it never contacts the backend or any other external resource, and never writes
the settings or the track file.

When no backend is configured, describing it is not an error: Sauron reports that
no backend is configured and exits successfully, in the spirit of an idempotent
no-op.

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to describe the singleton
  [backend](../0012-backend/spec.md), without taking a name.
- **FR-002**: Sauron shall treat describing the backend as read-only and offline,
  contacting no external resource and writing neither the settings nor the track
  file.

### Event-driven

- **FR-003**: When a user describes the backend and one is configured, Sauron
  shall read the singleton backend from the settings and display its fields, one
  per line as `field: value`, with the identity field `location` first.
- **FR-004**: When displaying the backend's fields, Sauron shall present
  `location` as the identity and `kind`, `auth`, `catalog-size`, `installed`, and
  `last-synced` as the available fields.
- **FR-005**: When displaying `auth`, Sauron shall render it as `REDACTED` and
  shall never print a credential reference or a resolved secret.
- **FR-006**: When displaying a count field (`catalog-size`, `installed`), Sauron
  shall print it as an integer.
- **FR-007**: When a field has no value, Sauron shall show it with an empty
  value.

### State-driven

- **FR-008**: While no backend is configured, Sauron shall report that no backend
  is configured and exit successfully.

### Unwanted behavior

- **FR-009**: If a name argument is supplied, then Sauron shall exit with code 2
  without executing the command and report that `describe backend` takes no name.
- **FR-010**: If `--fields` names a field other than `location`, `kind`, `auth`,
  `catalog-size`, `installed`, or `last-synced`, then Sauron shall exit with code
  2 without executing the command and report the allowed fields.
- **FR-011**: If the settings exist but cannot be read or parsed, then Sauron
  shall reject the request and report that the settings cannot be read.
- **FR-012**: If a command fails, then Sauron shall write exactly one
  human-readable message to stderr.

### Optional

- **FR-013**: Where `--fields` is provided, Sauron shall display only the named
  fields in the given order, with the identity field `location` always present
  and first; where it is omitted, Sauron shall display the full field set.

## Key Entities

- **Backend**: the singleton backend that owns persona definitions (see
  [backend](../0012-backend/spec.md)), described by its identity `location` and
  the fields `kind`, `auth`, `catalog-size`, `installed`, and `last-synced`. The
  `location` is the kind's locator (`url`/`path`/`uri`). `auth` is always shown
  `REDACTED` and never reveals a credential reference or a resolved secret;
  `catalog-size` is the number of personas in the local
  [catalog](../0012-backend/spec.md#key-entities), `installed` is the number of
  those activated locally by
  [select personas](../0014-select-personas/spec.md), and `last-synced` is when
  the catalog was last refreshed by
  [sync personas](../0013-sync-personas/spec.md). Fields with no value are shown
  empty.

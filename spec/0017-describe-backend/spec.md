# Describe Backend

**Type:** feature

**Depends on:** [backend](../0012-backend/spec.md)

## Overview

A person responsible for a team's agentic-AI setup needs to inspect the single
configured backend in full — its kind, where it lives, how it authenticates, and
when its persona definitions were last refreshed — so that they can review or
troubleshoot the persona source without re-reading the raw configuration. The
backend is the **singleton** owned by [backend](../0012-backend/spec.md);
describing it is the single-record counterpart that reads that one record.
Because there is exactly one backend per instance, the command takes no name. It
is read-only and offline: it never contacts the backend or any other external
resource, and never writes the configuration or the track file.

When no backend is configured, describing it is not an error: Sauron reports that
no backend is configured and exits successfully, in the spirit of an idempotent
no-op.

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to describe the singleton
  [backend](../0012-backend/spec.md), without taking a name.
- **FR-002**: Sauron shall treat describing the backend as read-only and offline,
  contacting no external resource and writing neither the configuration nor the track
  file.

### Event-driven

- **FR-003**: When a user describes the backend and one is configured, Sauron
  shall read the singleton backend connection from
  [backend.yaml](../contracts/configuration.md#backendyaml) and display its
  fields, one per line as `field: value`, with the identity field `uri` first.
- **FR-004**: When displaying the backend's fields, Sauron shall present `uri` as
  the identity and `kind`, `auth`, `timeout`, and `last-synced` as the available
  fields.
- **FR-005**: When displaying `auth`, Sauron shall render it as `REDACTED` and
  shall never print a credential reference or a resolved secret.
- **FR-006**: When displaying `last-synced`, Sauron shall print the
  `last_synced_at` timestamp recorded by
  [sync personas](../0013-sync-personas/spec.md).
- **FR-007**: When a field has no value, Sauron shall show it with an empty
  value.

### State-driven

- **FR-008**: While no backend is configured, Sauron shall report that no backend
  is configured and exit successfully.

### Unwanted behavior

- **FR-009**: If a name argument is supplied, then Sauron shall exit with code 2
  without executing the command and report that `describe backend` takes no name.
- **FR-010**: If `--fields` names a field other than `uri`, `kind`, `auth`,
  `timeout`, or `last-synced`, then Sauron shall exit with code 2 without
  executing the command and report the allowed fields.
- **FR-011**: If [backend.yaml](../contracts/configuration.md#backendyaml)
  exists but cannot be read or parsed, then Sauron shall reject the request and
  report that the backend connection cannot be read.
- **FR-012**: If a command fails, then Sauron shall write exactly one
  human-readable message to stderr.

### Optional

- **FR-013**: Where `--fields` is provided, Sauron shall display only the named
  fields in the given order, with the identity field `uri` always present and
  first; where it is omitted, Sauron shall display the full field set.

## Key Entities

- **Backend**: the singleton backend that owns persona definitions (see
  [backend](../0012-backend/spec.md)), described by its identity `uri` and the
  fields `kind`, `auth`, `timeout`, and `last-synced`, read from
  [backend.yaml](../contracts/configuration.md#backendyaml). The `uri` is the
  kind's locator (the `http`/`filesystem`/`git` source where persona definitions
  live). `auth` is always shown `REDACTED` and never reveals a credential
  reference or a resolved secret; `timeout` bounds network operations; and
  `last-synced` is the `last_synced_at` timestamp recording when persona
  definitions were last refreshed by
  [sync personas](../0013-sync-personas/spec.md). Fields with no value are shown
  empty.

## Notes

This spec was revised to the persona model in which there is **no persisted
catalog** (the set of available personas is a live view; see
[Live persona view](../contracts/configuration.md#live-persona-view)).
Behavior changes recorded for the meaning-preservation guard:

- The identity/locator field was renamed `location` → `uri`, matching
  [backend.yaml](../contracts/configuration.md#backendyaml), and the source
  term `personaRegistry` was renamed to `backend` throughout.
- The catalog markers `catalog-size` and `installed` were **removed** from the
  displayed and `--fields` field set. Because no catalog is persisted, the count
  of catalog entries and of those installed locally is no longer shown. FR-004
  is redefined to present `uri`, `kind`, `auth`, `timeout`, and `last-synced`;
  FR-006 (formerly "print a count field as an integer") is redefined to govern
  the `last-synced` timestamp, the nearest applicable display requirement.
- The source file was clarified from "the settings" to
  [backend.yaml](../contracts/configuration.md#backendyaml), which owns the
  singleton backend connection; the read-only/offline and no-write guarantees
  are unchanged.

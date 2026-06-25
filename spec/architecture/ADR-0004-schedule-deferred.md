# ADR-0004: Schedule is deferred; v1 relies on the user's own OS scheduler

**Status**: Accepted
**Date**: 2026-06-25
**Scope**: Project-wide

## Context

Sauron's reconcile operations — `sync` and `upgrade` — bring the environment
into line with the desired state. A natural follow-on want is to run them
*automatically*, on a periodic cadence, without a person invoking the command
each time.

A feature was specified for this (the `schedule` feature, `0014`). Its shape was:

- `sauron schedule (sync|upgrade) <expression>` registers an OS-crontab entry
  that runs the corresponding sauron command on the given cron expression, and
  records a `Schedule` document in `settings.yaml`.
- `sauron unschedule (sync|upgrade)` removes that operation's crontab entry and
  its `Schedule` document.
- Each operation has at most one schedule; re-scheduling replaces the single
  managed crontab entry in place rather than adding a second.
- The chosen mechanism is the operating system's own crontab facility (an OS
  facility, not a library). Managed entries are marked so they can be identified
  and removed without disturbing unmanaged entries, and the crontab entries are
  kept consistent with the `Schedule` documents: registering or removing one
  updates the other.
- Failure shape: an invalid cron expression or missing/invalid arguments exits
  with a usage error (code `2`) and registers nothing; a crontab that cannot be
  read or written fails with a runtime error and leaves existing entries
  unchanged; unscheduling an operation that is not scheduled exits `0` and
  reports that nothing was removed.

The `Schedule` document is a manifest kind sharing the common `metadata`
envelope: `metadata.name` is the operation (`sync` or `upgrade`) and
`spec.cron` is the cron expression registered in the OS crontab.

Owning this feature in v1 means owning OS-crontab manipulation — reading,
marking, replacing, and removing managed entries — and the consistency contract
between those entries and the persisted `Schedule` documents. That is a
meaningful surface to build and verify. Against it, the user's need is already
met by an OS facility every target platform already provides: a person can add
`sauron sync` (or `sauron upgrade`) to their own OS scheduler directly. v1.0.0
does not need to own periodic reconciliation to deliver value.

Deleting the `0014` feature directory would lose the design above. This ADR is
its durable home so the feature can be removed without losing the decision.

## Decision

The **schedule** feature is **not implemented in v1.0.0**. Sauron v1 ships no
`schedule` or `unschedule` command and manages no OS-crontab entries.

A user who wants periodic reconciliation in v1 wires it themselves: they add
`sauron sync` or `sauron upgrade` to their own OS scheduler (for example, an OS
crontab entry they author and own). Sauron does not register, mark, replace, or
remove crontab entries on the user's behalf in v1.

This ADR records the deferred design — the command grammar, the OS-crontab
mechanism, the one-schedule-per-operation rule, and the persisted `Schedule`
document — so that none of it is lost when the `0014` feature directory is
deleted. The `Schedule` manifest schema is **retained** as the reference
artifact: see
[Schedule.schema.json](../contracts/schemas/Schedule.schema.json).

The chosen mechanism, if and when the feature is built, is the operating
system's crontab facility — an OS facility, not a scheduling library. No
scheduling library is selected by this decision.

## Consequences

**Positive**

- v1.0.0 ships a smaller, simpler surface: no OS-crontab manipulation and no
  managed-entry consistency contract to build, test, or maintain.
- The periodic-reconciliation need is still met in v1 by an OS facility the user
  already has, with no sauron code in the path.
- The full schedule design and its rationale survive the deletion of the `0014`
  feature directory, recorded here and anchored to the retained
  `Schedule.schema.json`.
- Choosing the OS crontab facility, not a library, is recorded so a future
  implementation starts from the intended mechanism.

**Negative**

- Periodic reconciliation in v1 is a manual, user-owned setup; sauron neither
  registers it nor reports on it, and there is no single sauron command to
  inspect or change it.
- A user's hand-wired scheduler entry is invisible to sauron and unmanaged: it
  is not reconciled against any `Schedule` document and can drift from what
  sauron would have written.
- The retained `Schedule.schema.json` describes a document no v1 command writes,
  so the schema and the shipped behavior are intentionally out of step until the
  feature is built.

## Revisit when

Users need sauron to manage periodic reconciliation itself — registering,
replacing, and removing the scheduled runs — rather than wiring their own OS
scheduler by hand.

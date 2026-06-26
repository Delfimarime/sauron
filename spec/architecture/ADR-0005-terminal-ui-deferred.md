# ADR-0005: The terminal UI is deferred until the headless CLI is complete

**Status**: Accepted
**Date**: 2026-06-26
**Scope**: Project-wide

## Context

Sauron has two presentations of the same behavior: the **headless CLI** (the
[CLI contract](../contracts/cli.md)) and the interactive **terminal UI (TUI)**
(the [terminal UI contract](../contracts/terminal-ui.md)). Both are binding
design contracts, and both are deliberately thin: a use case holds all behavior
and returns a presentation-agnostic result, the CLI renders it through its
views, the TUI renders it on screen. Neither surface introduces behavior the
other lacks, and both draw from one shared vocabulary of field, flag, and value
names that must never diverge.

The headless CLI is the foundation, not a peer that happens to ship first. It is
the scriptable, CI-runnable, OS-crontab-schedulable surface, and it is the
surface that exercises the use-case layer end to end — every command drives a
use case and renders its result. The TUI renders those same use cases; it has
nothing to render until they exist and stable.

That use-case layer is still being built. Several headless commands — `install`,
`uninstall`, `sync`, `upgrade`, `list`, `describe` (artifact), and
`set provider` — are not yet implemented and verified. Their result shapes and
the shared field/flag vocabulary are still settling as those commands land.

Building the TUI now would fork effort onto a second presentation before the use
cases it renders are complete and stable. It would stand up a full-screen
application — and the full PTY-driven test matrix that proves it — over a
vocabulary and a set of results that are still moving, then demand that this
second surface be kept in lockstep with the core as the core changes. The cost
is a duplicated, drifting surface; the near-term value is zero, because the
headless CLI already delivers the whole install/uninstall/sync/upgrade/list/
describe/set-provider product without it.

## Decision

The **terminal UI** is **not implemented or supported** until the headless CLI
feature set is fully implemented and verified. No TUI code is written, no TUI
feature is built, and `sauron` does not launch an interactive full-screen
application in the meantime.

The [terminal UI contract](../contracts/terminal-ui.md) **remains a binding
contract** for the eventual TUI. It is retained, not deleted: it is the durable
home of the TUI design — surface selection, the shared vocabulary, and the
layouts — so that when the TUI is built it conforms to a contract already worked
out against the same use cases the CLI renders. Its companion
[terminal-ui.html](../contracts/terminal-ui.html) palette and layout reference is
retained alongside it.

Until then the headless CLI is sauron's only presentation: every command runs
through the [CLI contract](../contracts/cli.md), and the use-case layer is
proven end to end through that one surface before a second surface is laid over
it.

## Consequences

**Positive**

- Effort stays on the foundation: the use-case layer is completed and verified
  through one presentation, with no second surface to keep in sync while the
  result shapes and shared vocabulary are still settling.
- The TUI, when built, targets stable use cases and a settled vocabulary, so its
  rendering and its PTY test matrix are written once against a fixed target
  rather than chased against a moving one.
- The TUI design is not lost: the binding [terminal UI contract](../contracts/terminal-ui.md)
  and its visual companion are retained, so the eventual implementation starts
  from a contract already reconciled with the CLI.

**Negative**

- Until the headless CLI is complete, sauron has no interactive surface; users
  drive every operation through verb-noun commands, even where a full-screen,
  navigable view would be more convenient.
- A binding contract describes a surface no shipped build provides, so the
  contract and the running product are intentionally out of step until the TUI
  is built.
- Reconciling the TUI against any vocabulary or result-shape changes made while
  it was deferred is deferred work too: the contract must be re-checked against
  the finished use cases before the TUI is implemented.

## Revisit when

The headless CLI feature set — the `install`, `uninstall`, `sync`, `upgrade`,
`list`, `describe` (artifact), and `set provider` commands — is fully
implemented and verified, at which point the TUI can be planned and built
against stable use cases and a settled shared vocabulary.

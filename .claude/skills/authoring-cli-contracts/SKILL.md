---
name: authoring-cli-contracts
description: Use when writing or editing a feature's contracts/command-line.md, or the compiled spec/contracts/cli.md, in this repository — covers the per-command contract section order, the CLI conventions every command obeys (command grammar, shared flags, exit-status semantics, output discipline), and keeping the compiled command reference in sync. Specializes authoring-specs for the CLI surface.
---

# Authoring Sauron CLI Contracts

When creating or modifying a command's `contracts/command-line.md`, or the
compiled [spec/contracts/cli.md](../../../spec/contracts/cli.md), follow the
normative CLI conventions in
[spec/AUTHORING.md § CLI conventions](../../../spec/AUTHORING.md#cli-conventions).
This skill specializes [authoring-specs](../authoring-specs/SKILL.md) for the
CLI surface — both apply, and this one wins on conflict about command contracts.

## Three documents, three roles

- **`spec/AUTHORING.md` § CLI conventions** — the single source of truth for
  command grammar, shared-flag meanings, exit-status semantics, and output
  discipline. Never restate these in a command contract; conform to them.
- **`spec/NNNN-<feature>/contracts/command-line.md`** — the authoritative,
  full behavior of one command.
- **`spec/contracts/cli.md`** — the compiled command reference: one section per
  command with synopsis, one-line intent, key flags, and a link to the owning
  feature contract. It is derived, not authoritative.

## Per-command contract (`command-line.md`)

1. Title `# Contract: Command Line — <Command>`, then a pointer line
   `Conventions: [CLI conventions](../../AUTHORING.md).` and a `**Spec**:`
   markdown link to the feature `spec.md`.
2. Section order: `## Synopsis`, `## Arguments`, `## Flags`, `## Output`,
   `## Exit codes`, `## Examples`.
3. In `## Exit codes`, open with `Exit-status meanings are owned by the
   [CLI conventions](../../AUTHORING.md);` then list only the command-specific
   conditions — never contradict the global `0`/`2`/`1` meanings.
4. Every requirement reference is a relative markdown link, e.g.
   `Realizes [spec](../spec.md) FR-002`.
5. Reuse a shared flag with its conventional meaning; a contract may narrow a
   shared flag (e.g. restrict `--sort` values) but may not contradict it.

## Keep the compiled reference in sync

Whenever a command is added, renamed, removed, or its synopsis or flags change,
update its section in [spec/contracts/cli.md](../../../spec/contracts/cli.md):

- One `## <command>` section, ordered the same as the numbered specs.
- A synopsis (code block), a one-line intent, a compact `Key flags:` line, and
  `Full contract → [NNNN-<name>](../NNNN-<name>/contracts/command-line.md).`
- Do not copy whole flag or exit tables into the reference; link to the
  contract for detail. The reference summarizes; the contract governs.

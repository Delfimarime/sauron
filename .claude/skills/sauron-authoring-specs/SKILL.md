---
name: sauron-authoring-specs
description: Use when writing or editing any specification under spec/ in this repository — covers EARS requirement templates, the feature/capability taxonomy, chronological NNNN numbering, the shared glossary, markdown-link cross-references, required section order, canonical boilerplate, and ADR placement (feature-scoped vs project-level). The normative rules live in spec/AUTHORING.md.
---

# Authoring Sauron Specs

When creating or modifying any file under `spec/`, follow the normative rules in
[spec/AUTHORING.md](../../../spec/AUTHORING.md). Procedural reminders:

1. **Type first** — decide feature (user-observable, owns a command) vs
   capability (technical, nested under its feature, no CLI surface). Declare it
   in the header block.
2. **EARS + glossary** — phrase each requirement with a canonical EARS template;
   use glossary terms (no synonyms); reuse the canonical boilerplate verbatim
   for shared semantics (idempotent deletion, dry run, validation
   transactionality, usage error, failure output).
3. **Links** — write every cross-reference as a relative markdown link to the
   target file; never a bare id or name.
4. **Structure** — keep the required section order. CLI-wide conventions
   (grammar, shared flags, exit status, output) are owned by
   [spec/contracts/cli.md](../../../spec/contracts/cli.md), not restated.
5. **Meaning-preservation guard** — never let a wording change alter what a
   requirement demands. If a rewrite would change the meaning, keep the original
   sentence and record the deviation in the spec's `## Notes`.
6. **ADRs** — feature-scoped decisions live under
   `spec/NNNN-<feature>/architecture/ADR-NNNN-*.md`; cross-cutting decisions owned
   by no single feature live under the project-level `spec/architecture/`
   directory (with a `**Scope**` header in place of `**Feature**`). A
   `PROPOSAL:` issue's EARS and intended ADRs are formalized here once accepted.
   **Never author an ADR without explicit user permission** — propose it and wait.
7. **Executable behaviour.** A feature's user-observable behaviour is verified
   end-to-end by Gherkin scenarios under `test/e2e/testdata`; keep requirements
   expressible as `.feature` scenarios. See
   [`sauron-implementing-integration-tests`](../sauron-implementing-integration-tests/SKILL.md).

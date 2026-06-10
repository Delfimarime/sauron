---
name: authoring-specs
description: Use when writing or editing any specification under spec/ in this repository — covers EARS requirement templates, the feature/capability taxonomy, chronological NNNN numbering, the shared glossary, markdown-link cross-references, required section order, and canonical boilerplate. The normative rules live in spec/AUTHORING.md.
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

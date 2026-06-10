# CLAUDE.md

Project instructions for sauron.

## Writing specifications

When writing or editing anything under `spec/`, use the **authoring-specs**
skill (`.claude/skills/authoring-specs/`). The normative authoring rules live in
[spec/AUTHORING.md](spec/AUTHORING.md): spec types (feature vs capability), EARS
templates, NNNN numbering, the glossary, required sections, and markdown-link
cross-references. The CLI conventions every command obeys (grammar, shared
flags, exit status, output discipline) live in the same file under
[§ CLI conventions](spec/AUTHORING.md#cli-conventions).

When writing or editing a command's `contracts/command-line.md` or the compiled
[spec/contracts/cli.md](spec/contracts/cli.md), use the
**authoring-cli-contracts** skill (`.claude/skills/authoring-cli-contracts/`).

## Implementation

Project and implementation principles live in
[CONSTITUTION.md](CONSTITUTION.md).

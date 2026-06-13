# AGENTS.md

Agent and contributor instructions for sauron. This is the source of truth for
how agents and humans work in this repository; `CLAUDE.md` (and any other
assistant config) points here. It **routes** to the project's rules and carries
only the meta-knowledge that lives nowhere else.

## Precedence

Project-scoped skills and agents under `.claude/` **take precedence over**
user-level (`~/.claude/`) skills and agents. Where they conflict — e.g. this
repo's Use Case/Action architecture versus a personal "services-as-interfaces"
convention — the **project** definition wins. The user-level definitions still
apply where the project says nothing.

## Where the rules live

- **Principles & governance** — [CONSTITUTION.md](CONSTITUTION.md) (spec-driven,
  EARS, ADRs, ports & adapters, Use Case orchestration, the verification gate,
  dependency/license/security discipline, Trunk-based development).
- **Spec authoring** — [spec/AUTHORING.md](spec/AUTHORING.md) (spec types, EARS,
  numbering, glossary, ADR structure, CLI conventions).
- **Technical contract** — [spec/contracts/architecture.md](spec/contracts/architecture.md)
  (layout, fx wiring, Use Case/Action interfaces, storage, root command,
  versioning, CI/CD, gates, approved dependencies).
- **Contributing & workflow** — [CONTRIBUTING.md](CONTRIBUTING.md) (Trunk flow,
  Conventional Commits, SemVer bump, `PROPOSAL:` and bug issues, templates).

## Project skills (`.claude/skills/`)

- `sauron-authoring-specs` — authoring/editing specs under `spec/`.
- `sauron-authoring-cli-contracts` — command-line contracts; specializes the above.
- `sauron-implementing-architecture` — writing Go (Use Case/Action, infrastructure,
  storage, versioning, pond, local gates). Overrides `golang-personal-architecture`.
- `sauron-operating-ci` — CI/CD pipeline files.

## Project agents (`.claude/agents/`)

- `sauron-architect` — read-only; audits Go against the architecture contract.
- `sauron-developer` — implements the Use Case/Action layer (write).
- `sauron-adr-author` — authors ADRs **only with explicit user permission** (write).
- `sauron-gatekeeper` — read-only; runs `task all` and reports the gate result.
- `sauron-ci-operator` — authors/maintains CI files under `.github/` / `.gitlab/`
  (write; no live-pipeline or API access).

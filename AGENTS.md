# AGENTS.md

Agent and contributor instructions for sauron. This is the source of truth for
how agents and humans work in this repository; `CLAUDE.md` (and any other
assistant config) points here. It **routes** to the project's rules and carries
only the meta-knowledge that lives nowhere else.

## Product context

What sauron is and how it behaves: the [README](README.md) (overview) and the
[domain documentation](spec/README.md) (the canonical model — registries,
personas, providers, the sync plan, and state). Read these before changing
behavior; the rules below govern *how* changes are made.

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
- **Architecture contract** — [spec/contracts/architecture.md](spec/contracts/architecture.md)
  (layout, fx wiring, Use Case/Action interfaces, storage, root command,
  approved dependencies).
- **Delivery contract** — [spec/contracts/delivery.md](spec/contracts/delivery.md)
  (build, the Taskfile gates, CI/CD, versioning).
- **CLI contract** — [spec/contracts/cli.md](spec/contracts/cli.md) (command
  grammar, shared flags, exit status, output discipline).
- **State data contract** — [spec/contracts/state.md](spec/contracts/state.md)
  (the persisted `~/.sauron/` documents; structure owned by the
  [schemas](spec/contracts/schemas/)).
- **Integration tests** — [test/e2e/CONSTITUTION.md](test/e2e/CONSTITUTION.md)
  (the black-box `test/e2e` harness: intent, runtime/Source architecture,
  controllers, the `#{}` resolver, fixtures, tags, and the integration gate).
- **Contributing & workflow** — [CONTRIBUTING.md](CONTRIBUTING.md) (Trunk flow,
  Conventional Commits, SemVer bump, `PROPOSAL:` and bug issues, templates).

## Task routing

Match the task to its governing doc, skill, and agent:

| Task | Governing docs | Skill | Agent |
|---|---|---|---|
| Author / edit a spec | [AUTHORING.md](spec/AUTHORING.md); for a command, the [CLI contract](spec/contracts/cli.md) | `sauron-authoring-specs`, `sauron-authoring-cli-contracts` | — |
| Write / modify Go | [architecture.md](spec/contracts/architecture.md), [CONSTITUTION Ch. III](CONSTITUTION.md) | `sauron-implementing-architecture` | `sauron-developer` (write) · `sauron-architect` (audit) |
| Integration tests | [test/e2e/CONSTITUTION.md](test/e2e/CONSTITUTION.md) | `sauron-implementing-integration-tests` | `sauron-integration-test-developer` |
| CI/CD pipeline files | [delivery.md](spec/contracts/delivery.md) | `sauron-operating-ci` | `sauron-ci-operator` |
| Record a decision (ADR) | [CONSTITUTION Ch. I, Art. 4–5](CONSTITUTION.md), [ADR structure](spec/AUTHORING.md#adr-structure) | `sauron-authoring-specs` | `sauron-adr-author` (**explicit user permission required**) |
| Verify before merge | [delivery.md](spec/contracts/delivery.md), [CONSTITUTION Ch. IV, Art. 2](CONSTITUTION.md) | — | `sauron-gatekeeper` (`task all`) |

## Project skills (`.claude/skills/`)

- `sauron-authoring-specs` — authoring/editing specs under `spec/`.
- `sauron-authoring-cli-contracts` — command-line contracts; specializes the above.
- `sauron-implementing-architecture` — writing Go (Use Case/Action, infrastructure,
  storage, versioning, pond, local gates). Overrides `golang-personal-architecture`.
- `sauron-implementing-integration-tests` — writing the black-box `test/e2e` suite
  (godog, testcontainers, the graybox `pkg/`-only pattern).
- `sauron-operating-ci` — CI/CD pipeline files.

## Project agents (`.claude/agents/`)

- `sauron-architect` — read-only; audits Go against the architecture contract.
- `sauron-developer` — implements the Use Case/Action layer (write).
- `sauron-integration-test-developer` — implements the `test/e2e` harness (write;
  touches only `test/e2e/**`).
- `sauron-adr-author` — authors ADRs **only with explicit user permission** (write).
- `sauron-gatekeeper` — read-only; runs `task all` and reports the gate result.
- `sauron-ci-operator` — authors/maintains CI files under `.github/` / `.gitlab/`
  (write; no live-pipeline or API access).

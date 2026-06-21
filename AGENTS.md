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

## Operating rules

How an agent conducts itself in this repo, in three tiers. **When unsure which
tier a change falls in, treat it as Ask first.**

### Always do — routine hygiene, no permission needed

- Invoke the matching skill before acting, and follow the [task routing](#task-routing) table.
- Work from an approved spec and cite the `FR-NNN` ids the change realizes.
- Keep tests hermetic: `afero.MemMapFs`, injected lookups, a temp `SAURON_HOME` —
  never the real `~/.sauron`, never mutate process env.
- Run the relevant local gate and **show the output** before claiming done
  (`task test` / `gate-lint` / `gate-coverage`; `gate-integration` for e2e).
  Evidence before assertions.
- Keep it `gofmt`/`go vet`-clean, gocognit ≤ 15, one concise doc line per exported symbol.
- Use glossary terms; write cross-references as relative markdown links; when you
  rename or move a file, fix every inbound link.
- Single-source: link the owning contract, never restate it.

### Ask first — high-impact or risky, explicit user permission required

- Author or modify an ADR — propose and wait; never auto-generate. An ADR records
  a decision, not a library.
- Add, remove, or upgrade a dependency — needs license review and amending the
  approved-dependency table.
- Change a normative contract (architecture / cli / state / delivery), the
  Constitution, AUTHORING, or the shared glossary/vocabulary — anything that
  ripples across the corpus.
- Change a public port (`pkg/sauron/extension`) or the persisted-state schema
  (a breaking on-disk change).
- Restructure the spec corpus — split, rename, or move a contract or feature.
- Resolve an open/pending spec question that constrains design — answer first,
  record as an ADR once approved.

### Never do — hard stop

- Write implementation without an approved spec, or proceed past an open question
  by guessing in code.
- Commit, push, or open a PR/MR unless the user explicitly asks; never add a
  `Co-Authored-By` trailer.
- Name a specific library or tool in an ADR.
- Write a resolved secret value to any file — credentials persist only as
  `${env:VAR}` references.
- Start a bare goroutine — all concurrency runs on the injected `pond` pool.
- Import `internal/` from `test/e2e` or from `pkg/`.
- Introduce CGO or otherwise break the `CGO_ENABLED=0` cross-compile.
- Merge with a failing gate or coverage below the 80% floor.

## Where the rules live

- **Principles & governance** — [CONSTITUTION.md](CONSTITUTION.md) (spec-driven,
  EARS, ADRs, ports & adapters, Use Case orchestration, the verification gate,
  dependency/license/security discipline, Trunk-based development).
- **Spec authoring** — [spec/AUTHORING.md](spec/AUTHORING.md) (spec types, EARS,
  numbering, ADR structure, CLI conventions).
- **Glossary** — [spec/GLOSSARY.md](spec/GLOSSARY.md) (the canonical domain
  vocabulary; use these terms, no synonyms).
- **Workflow walkthrough** — [spec/WORKFLOW.md](spec/WORKFLOW.md) (one slice from
  requirement through spec, contract, code, and the verification gate).
- **Security overview** — [spec/SECURITY.md](spec/SECURITY.md) (the security
  posture — secret handling, TLS, write integrity, file permissions — for
  security analysts).
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
- **Integration tests** — [test/e2e/HARNESS.md](test/e2e/HARNESS.md)
  (the black-box `test/e2e` harness: runtime/Source architecture,
  controllers, the `#{}` resolver, fixtures, tags, and the integration gate).
- **Contributing & workflow** — [CONTRIBUTING.md](CONTRIBUTING.md) (Trunk flow,
  Conventional Commits, SemVer bump, `PROPOSAL:` and bug issues, templates).

## Task routing

Match the task to its governing doc, skill, and agent:

| Task | Governing docs | Skill | Agent |
|---|---|---|---|
| Author / edit a spec | [AUTHORING.md](spec/AUTHORING.md); for a command, the [CLI contract](spec/contracts/cli.md) | `sauron-authoring-specs`, `sauron-authoring-cli-contracts` | — |
| Write / modify Go | [architecture.md](spec/contracts/architecture.md), [CONSTITUTION Ch. III](CONSTITUTION.md) | `sauron-implementing-architecture` | `sauron-developer` (write) · `sauron-architect` (audit) |
| Integration tests | [test/e2e/HARNESS.md](test/e2e/HARNESS.md) | `sauron-implementing-integration-tests` | `sauron-integration-test-developer` |
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

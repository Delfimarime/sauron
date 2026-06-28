# Workflow

How one change travels from a requirement to shipped, verified code in this
spec-driven project. This is the connective how-to: it names the stages, the
artifacts each produces, and the commands that gate them — and links to the
documents that own each rule rather than restating them
([Constitution Ch. IV, Art. 4](../CONSTITUTION.md)).

**Audience:** developers and architects working on sauron, and engineers new to
the repository who need the lifecycle in one place.

The lifecycle is a **loop, not a line** — the four stages below map to the
SpecDriven flywheel phases **Discover → Design → Deliver → Distill**, and
Distill feeds the next Discover. The [set registry](0001-set-registry/spec.md)
feature (the first built feature) is the running example.

## 1. Discover

*Flywheel: Discover.* A change starts as intent, not code.

- Open a **`PROPOSAL: <name>`** issue (or **`BUG: <summary>`** for a defect),
  per [CONTRIBUTING.md](../CONTRIBUTING.md#proposing-a-feature). A proposal
  states the intent, the requirements in
  [EARS form](AUTHORING.md#ears-templates-normative), and any **intended ADRs**
  — it *describes* decisions; it does not commit them.
- The proposal carries no implementation. It is the seed an approved spec grows
  from.

For set registry, the seed was "a developer must tell Sauron where artifacts
come from before anything can be installed" — which became FR-001 onward.

## 2. Design

*Flywheel: Design.* An accepted proposal becomes an approved spec under
`spec/NNNN-<name>/` ([Constitution Ch. I](../CONSTITUTION.md), Articles 1–3),
numbered chronologically. Authoring mechanics are normative in
[AUTHORING.md](AUTHORING.md) — the `sauron-authoring-specs` skill loads them.

Author, in this order, omitting what the feature does not touch:

| Artifact | When | Owns |
|---|---|---|
| [`spec.md`](0001-set-registry/spec.md) | always | overview, EARS `FR-NNN` requirements, key entities; **`Status:` Specified** |
| [`contracts/<verb>-<noun>.md`](0001-set-registry/contracts/set-registry.md) | per owned command | synopsis, args, flags, output, exit codes — conforms to the [CLI contract](contracts/cli.md) |
| [`data/state.md`](contracts/state.md) | touches persisted state | which documents/fields it owns + the `FR-NNN` realization; the schema stays in the [state data contract](contracts/state.md) |
| `capabilities/<name>.md` | needs technical sub-behavior | one capability, no CLI surface ([set registry](0001-set-registry/spec.md) has git and http transports) |
| `architecture/ADR-NNNN-*.md` | a significant decision needs recording | one decision — authored **only with the maintainer's explicit intent** ([Ch. I, Art. 4](../CONSTITUTION.md)), never auto-generated |

Requirements are EARS, describing observable behavior, not implementation
([Ch. I, Art. 2](../CONSTITUTION.md)). Every cross-reference is a relative
markdown link ([AUTHORING.md cross-references](AUTHORING.md#cross-references)).

**Gate to leave Design:** no open or pending questions. While any ambiguity
remains the spec is not ready for implementation
([Ch. I, Art. 6](../CONSTITUTION.md)); resolve it — recording an ADR where the
answer constrains the design — before writing code, never by guesswork in
source.

## 3. Deliver

*Flywheel: Deliver.* Build the code that satisfies the contracts
([Constitution Ch. III](../CONSTITUTION.md)) and pass the verification gate
([Ch. IV, Art. 2](../CONSTITUTION.md)).

**a. Plan the work.** Write `plan.md` (goal & scope, design, checkpoints — each
with a verify command) and the paired `TASKS.md` (one independently verifiable
task per unit of work, with dependency order). Both trace back to the `FR-NNN`
they fulfill ([Ch. IV, Art. 1](../CONSTITUTION.md)), and both are ordered
**test-first**: the test — and, for user-observable behavior, the `test/e2e`
Gherkin scenario — comes before the implementation it verifies
([Ch. III, Art. 1](../CONSTITUTION.md)). The `sauron-developer` agent
executes a planned unit; the `plan-implementation` skill shapes the plan.

**b. Branch and implement.** Take a short-lived branch off `main`
([CONTRIBUTING Trunk flow](../CONTRIBUTING.md#branching-model--trunk-based)).
Implement under `internal/` following the
[architecture contract](contracts/architecture.md): a command's logic is a
**Use Case** that orchestrates **Actions**, stateless with collaborators
injected by uberfx and per-invocation context arriving through a `Request`
([Ch. III, Art. 3–4](../CONSTITUTION.md)). Set registry, for example, ships
`SetRegistryUseCase` behind the `extension.Registry` port. Work is **test-first
to the 90% coverage target** ([Ch. III, Art. 1](../CONSTITUTION.md)).

**c. Pass the gate.** The [Taskfile](../Taskfile.yml) targets match the CI jobs
1:1; run them as you go, then `task all` before opening the PR:

| Target | Checks |
|---|---|
| `task test` | unit tests (race detector) + coverage report |
| `task gate-lint` | Uber style + `gocognit ≤15` (golangci-lint) |
| `task gate-coverage` | 90% ideal, fails below the 80% floor |
| `task gate-security` | Trivy scan — 0 CRITICAL, ≤2 HIGH |
| `task gate-integration` | the black-box `test/e2e` BDD suite against the built binary |
| `task all` | every gate, in dependency order |

The `sauron-gatekeeper` agent runs the gate and reports; an exception to a
security threshold requires a project-level ADR ([Ch. IV, Art. 2](../CONSTITUTION.md)).

**d. Land it.** Open a PR whose title follows
[Conventional Commits](../CONTRIBUTING.md#commits--versioning), traces to its
spec / `FR-NNN`, and bumps `package.json` by change type (feat → minor,
fix → patch, breaking → major). On merge, flip the spec's **`Status:`** to
**Built** (or `Partial` if it defers requirements). That field is the single
source of build status; the [spec index](README.md#specifications) aggregates it
as a view.

## 4. Distill

*Flywheel: Distill — the loop back to Discover.* Delivery and use teach things
the spec could not predict ([Constitution Ch. IV, Art. 5](../CONSTITUTION.md)).

- Every command leaves an observable record — its terminal outcome through the
  structured zap + ECS logger — so behavior is inspected, not guessed.
- An insight (a spec that mispredicted behavior, a missed edge case, a usage
  pattern that contradicts the design) is recorded in the owning feature's
  [`plan.md` `## Distill`](AUTHORING.md#numbering-and-layout) section — a table,
  one row per insight, added only when there is something to record.
- Each row is **closed out**, not just logged: as a spec amendment carrying its
  `FR-NNN` trace, or as a new `PROPOSAL:` issue that reopens Discover. An
  insight that changes a requirement is reflected in the spec **before** the
  code. The loop turns.

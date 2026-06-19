# Contributing

How work is proposed, branched, and landed on sauron. This complements the
[Constitution](CONSTITUTION.md) (governing principles), the spec authoring rules
in [spec/AUTHORING.md](spec/AUTHORING.md), and the technical
[architecture contract](spec/contracts/architecture.md).

## Branching model — Trunk-based

- `main`/`master` is the **only** long-lived branch and is always releasable.
- Work happens on **short-lived feature branches** taken off `main`/`master` and
  merged back through a pull/merge request. There are no `develop` or `release`
  branches — this is **not** Gitflow.
- A feature branch is kept small and short-lived, and is deleted after it merges.

## Commits & versioning

- Commits and PR/MR titles follow
  [Conventional Commits](https://www.conventionalcommits.org):
  `(feat|refactor|chore|fix): <intent>` — the message states the intent of the
  change.
- The `version` in `package.json` is the SemVer source of truth and is **bumped
  by hand in the same PR/MR**, to match the change type:
  - a **feature** (`feat`) → **minor** bump,
  - a **fix** (`fix`) → **patch** bump,
  - a **breaking change** → **major** bump,
  - `refactor`/`chore` that change no behavior → no bump.
- CI decorates that version into the build artifact label (`-RELEASE`,
  `-PRE-RELEASE.<n>`, `-SNAPSHOT.<n>`); see the
  [delivery contract](spec/contracts/delivery.md#versioning).

## Proposing a feature

Open an issue titled **`PROPOSAL: <name>`**. The description states the intent and
includes:

- the requirements in **EARS** form (see the
  [EARS templates](spec/AUTHORING.md#ears-templates-normative)), and
- any **intended ADRs** describing the technical decisions behind the proposal.

An accepted proposal becomes a spec under `spec/NNNN-<name>/`, where the EARS
requirements and the committed ADRs live. ADRs are authored only with explicit
intent at that point — the proposal *describes* the intended decisions; it does
not commit them.

## Reporting a bug

Open an issue titled **`BUG: <summary>`**, written with the **CPE** pattern so the
defect is unambiguous:

- **Context** — the `sauron` version (`sauron --version`), the OS, and the steps
  or state that led to the problem.
- **Problem** — what actually happened: the observed behavior, error, or output.
- **Expected Outcome** — what should have happened instead.

State the issue type clearly so it is triaged correctly.

## Pull / merge requests

- Title follows Conventional Commits (above); the description states the intent
  of the changes.
- Trace the change to its spec / `FR-NNN` (or the `PROPOSAL:` issue it realizes).
- The change bumps `package.json` per its type, and passes the
  [verification gate](CONSTITUTION.md) locally (`task all`): tests, lint,
  coverage (≥ 80%), and the dependency security scan.

## Issue & PR templates

The same content is provided per platform; pick the one matching the host:

- **GitHub** — `.github/ISSUE_TEMPLATE/feature_proposal.md`,
  `.github/ISSUE_TEMPLATE/bug_report.md`, and
  `.github/PULL_REQUEST_TEMPLATE.md`.
- **GitLab** — `.gitlab/issue_templates/Feature_Proposal.md`,
  `.gitlab/issue_templates/Bug_Report.md`, and
  `.gitlab/merge_request_templates/Default.md`.

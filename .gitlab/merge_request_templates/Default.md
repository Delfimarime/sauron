<!--
  Set the MR title to follow Conventional Commits:
    (feat|refactor|chore|fix): <intent>
  See CONTRIBUTING.md.
-->

## Intent

<!-- What this change does and why. -->

## Traceability

<!-- The spec / FR-NNN this realizes, or the PROPOSAL: issue it comes from. -->

## Checklist

- [ ] Title follows Conventional Commits: `(feat|refactor|chore|fix): <intent>`
- [ ] Change traces to an approved spec under `spec/` (or the `PROPOSAL:` issue it realizes)
- [ ] Any significant technical decision is recorded as an ADR (or N/A)
- [ ] Any new dependency is added to the approved-dependency table with its license (or N/A)
- [ ] `package.json` `version` bumped to match the change type (feat → minor, fix → patch, breaking → major)
- [ ] Verification gate passes locally: `task all` (tests, lint, coverage ≥ 80%, dependency scan, integration tests)

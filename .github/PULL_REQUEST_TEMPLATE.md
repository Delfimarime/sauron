<!--
  Title must follow Conventional Commits:
    (feat|refactor|chore|fix): <intent>
  See CONTRIBUTING.md.
-->

## Intent

<!-- What this change does and why. -->

## Traceability

<!-- The spec / FR-NNN this realizes, or the PROPOSAL: issue it comes from. -->

## Checklist

- [ ] Title follows Conventional Commits: `(feat|refactor|chore|fix): <intent>`
- [ ] `package.json` `version` bumped to match the change type (feat → minor, fix → patch, breaking → major)
- [ ] Verification gate passes locally: `task all` (tests, lint, coverage ≥ 80%, dependency scan)

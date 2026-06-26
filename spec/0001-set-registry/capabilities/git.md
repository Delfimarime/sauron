# Git Transport

**Type:** capability

**Enables:** [set registry](../spec.md)

**Enables:** [list catalogue](../../0004-list-catalogue/spec.md)

**Enables:** [install](../../0007-install-artifacts/spec.md)

## Overview

The git transport reaches a registry hosted in a git repository. It validates the
source when set and fetches artifact content for browsing, installing, and
reconciling. Artifacts are read from the repository's `.skills/` and `.agents/`
directories; a skill or agent is the directory under one of those.

## Requirements

### Ubiquitous

- FR-001: Sauron shall reach git registries over the URI's scheme, supporting SSH
  remotes with a private key (`--ssh-key`) and HTTPS remotes with credentials
  passed as environment references.
- FR-002: Sauron shall treat each directory under `.skills/` or `.agents/` as one
  skill or agent.
- FR-003: Sauron shall compute an artifact's `digest` from the tree-object hash of
  its directory at the resolved commit.

### Event-driven

- FR-004: When validating a git registry, Sauron shall confirm the repository is
  reachable and hosts at least one skill or agent.

### Optional

- FR-005: Where an explicit version is not declared, Sauron shall derive a git
  artifact's optional `version` from the most recent commit that touched the
  artifact's directory (that commit's SHA).
- FR-007: Where a ref is provided, Sauron shall resolve the registry's artifacts
  from that ref (a branch, tag, or commit); where no ref is provided, Sauron shall
  resolve from the repository's default branch.

### Unwanted behavior

- FR-006: If the repository is unreachable or authentication fails, then Sauron
  shall fail with a runtime error.
- FR-008: If the provided ref cannot be resolved in the repository, then Sauron
  shall fail with a runtime error.

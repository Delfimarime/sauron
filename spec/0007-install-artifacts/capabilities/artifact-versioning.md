# Artifact Versioning

**Type:** capability

**Enables:** [install](../spec.md)

**Enables:** [sync](../../0011-sync/spec.md)

**Enables:** [upgrade](../../0012-upgrade/spec.md)

## Overview

Every installed artifact carries a content `digest` and an optional, human-readable
`version`. The `digest` is the always-present identity that reconcile operations
compare to detect change and local drift; the `version` is a label whose
availability depends on the registry transport. There is no artifact-level
version pinning: install takes what the registry currently offers, and
sync/upgrade move to the source's latest.

## Requirements

### Ubiquitous

- FR-001: Sauron shall record a content `digest` for every installed artifact and
  use it to decide whether an artifact has changed upstream or drifted locally.
- FR-002: Sauron shall treat the `version` as optional metadata only, never as a
  pin: reconcile decisions are made on `digest`, not `version`.

### Optional

- FR-003: Where the registry is `git`, Sauron shall derive an artifact's `version`
  from the most recent commit that touched the artifact's directory when no
  explicit version is declared, and compute its `digest` from the directory's
  tree-object hash.
- FR-004: Where the registry is `http` or `filesystem`, Sauron shall record a
  `version` only when the source or artifact declares one, and compute the
  `digest` from artifact content.

## Notes

`version` is metadata, not a pin: there is no artifact-level version pinning, and
none is reserved (YAGNI). A `git` registry's content — and therefore each
artifact's derived `version` and `digest` — is resolved at the registry's
[`spec.revision`](../../0001-set-registry/capabilities/git.md) when one is set.

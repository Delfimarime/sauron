# Artifact Versioning

**Type:** capability

**Enables:** [install](../spec.md)

**Enables:** [sync](../../0011-sync/spec.md)

**Enables:** [upgrade](../../0012-upgrade/spec.md)

## Overview

Every installed artifact carries a single `version`: the identity Sauron records at
install and compares on reconcile to decide whether the artifact changed upstream.
The `version` is read from the source, never computed by Sauron — its form depends on
the transport. There is no artifact-level version pinning: install records the
source's current `version`, and sync/upgrade move to the source's latest.

## Requirements

### Ubiquitous

- FR-001: Sauron shall record a `version` for every installed artifact, read from the
  source rather than computed, and use it as the artifact's identity to detect
  upstream change during `sync` and `upgrade`.
- FR-002: Sauron shall treat `version` as a recorded identity, never as a user pin:
  install records the source's current `version`, and `sync`/`upgrade` reconcile to
  the source's latest.

### Optional

- FR-003: Where the registry is `git`, Sauron shall set an artifact's `version` to the
  git tree-object hash of the artifact's directory, resolved at the registry's
  [`spec.revision`](../../0001-set-registry/capabilities/git.md) when one is set.
- FR-004: Where the registry is `http`, Sauron shall set an artifact's `version` to
  the object version the registry declares for it.

### Unwanted behavior

- FR-005: If a registry offers an artifact that declares no version, then
  Sauron shall report it and skip it, continuing with the remaining artifacts.

## Notes

`version` is an identity, not a pin: there is no artifact-level version pinning, and
none is reserved (YAGNI). Because the `version` is read from the source and nothing is
computed locally, Sauron detects upstream change but not local edits to an installed
artifact.

`install`, `sync`, and `upgrade` share one internal reconcile step that computes the
add/update/remove/unchanged plan by comparing each artifact's `version` against the
tracked set, so [sync](../../0011-sync/spec.md) and
[upgrade](../../0012-upgrade/spec.md) reference one mechanism.

A `git` artifact's `version` (the directory's tree hash) is never empty, so only an
`http` artifact can trigger FR-005.

# Data Model: Configuration — Sauron Settings (Describing)

**Spec**: [Describe Registry](../spec.md)

Describes the configuration that the Describe Registry feature reads. Describing
is read-only and offline: it never writes the settings or the track file, and
defines no schema of its own.

## Location & format

- **Path**: `~/.sauron/settings.yaml` (home directory resolved per platform).
- **Format**: a single YAML document. A missing file is treated as the registry
  not found. Realizes [spec](../spec.md) FR-010.

## Read shape

The registry schema — the `registries` array, the common fields (`name`,
`kind`, `priority`), and the kind-scoped fields (`url`/`path`/`uri`, `auth`,
`tls`, `timeout`, `ssh`) — is owned by
[add registry](../../0001-add-registry/data/configuration.md). Describe Registry
reads that schema; it adds nothing to it.

The named registry is matched by its identity `name`. Its `location` is the
kind-appropriate locator (`path`/`url`/`uri`); fields that do not apply to the
resolved kind are shown empty. `auth` is read only as its `${env:VAR}`
reference, and a resolved secret is never read into output — it is rendered
`REDACTED`. Realizes [spec](../spec.md) FR-004, FR-005, FR-006.

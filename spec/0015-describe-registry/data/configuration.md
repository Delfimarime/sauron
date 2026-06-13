# Data Model: Configuration — Describe Registry (`registries.yaml`)

**Spec**: [Describe Registry](../spec.md)

Describe Registry reads one `items` entry from `registries.yaml` by name
and shows its fields; it never writes. The schema is owned by the
[configuration data contract](../../contracts/configuration.md#registriesyaml);
this document does not restate it.

## Reads

- A single `items` entry of `registries.yaml`, matched by its identity
  `name` — its `name`, `kind`, `priority`, `uri`, and the kind-scoped
  credential fields (`auth`, `tls`, `timeout`, `ssh`). A missing file is treated
  as the registry not found. Realizes FR-003, FR-004, FR-005, FR-010.
- `auth` is read only as its `${env:VAR}` reference; a resolved secret is never
  read into output — it is rendered `REDACTED`. Realizes FR-006.

## Owns / Writes

- Nothing. Describing is read-only and offline.

## Notes

Configuration is now split across files per the
[configuration data contract](../../contracts/configuration.md); file references
updated accordingly.

# Data Model: Configuration — Add Registry (registries.yaml)

**Spec**: [Add Registry](../spec.md)

Add Registry owns `registries.yaml` and the `items` entries it appends.
The schema — fields, constraints, and write semantics — is owned by the
[configuration data contract](../../contracts/configuration.md#registriesyaml);
this document does not restate it.

## Owns

- `registries.yaml` `items`: each registered source, keyed by `name`, with `kind`, `priority`, `uri`, and kind-scoped credentials. See the [contract](../../contracts/configuration.md#registriesyaml).
- Validation before persist: an entry is appended and written back only after kind validation passes; the existing configuration is left unchanged until then (FR-006, FR-011). See the [contract](../../contracts/configuration.md#registriesyaml).
- Priority assignment for a new entry follows the [priority model](../../AUTHORING.md#priority-model) (`0` for the first registry, `max + 1` when omitted later), as governed by the [contract](../../contracts/configuration.md#registriesyaml).

## Realizes

- `registries.yaml` `items` entry write → [spec](../spec.md) FR-002, FR-005 (name identity), FR-004 (kind), FR-003, FR-009, FR-010, FR-017 (priority), FR-007 (records name/priority/kind/uri), FR-018 (name/priority uniqueness), FR-006, FR-011 (validated atomic write).

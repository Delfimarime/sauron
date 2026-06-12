# Data Model: Configuration — Select Personas (personas.yaml)

**Spec**: [Select Personas](../spec.md)

Select Personas owns `personas.yaml` `items`, storing each installed
persona with its full definition (fetched live from the backend at install
time). The schema and write semantics are owned by the
[configuration data contract](../../contracts/configuration.md#personasyaml);
this document does not restate them.

## Owns / writes
- `personas.yaml` `items`: `set persona` full-replaces it (positional `priority`) and stores each persona's full definition fetched from the [backend](../../contracts/configuration.md#backendyaml); `unset persona` removes entries.

## Realizes

- `personas.yaml` `items` (`name`, `priority`) → [spec](../spec.md) FR-004
  (install named personas), FR-005 (positional priority), FR-007 (full-replace
  re-declaration), FR-010 (transactional validation), FR-015 (read failure).
- `personas.yaml` `items` definition fields → [spec](../spec.md) FR-017 (store
  each persona's full definition at install time).

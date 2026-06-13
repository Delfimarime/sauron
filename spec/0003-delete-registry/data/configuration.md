# Data Model: Configuration — Delete Registry (registries.yaml)

**Spec**: [Delete Registry](../spec.md)

This feature mutates `registries.yaml`'s `items` block, removing the
entry whose `name` matches the argument; installed artifacts are untouched. The
schema is owned by the
[configuration data contract](../../contracts/configuration.md#registriesyaml);
this document does not restate it.

## Reads

- `registries.yaml` `items`: the entry whose `name` matches the argument.

## Owns

- Nothing. `registries.yaml` is owned by
  [add registry](../../0001-add-registry/spec.md).

## Writes

- `registries.yaml` `items`: removes the matching entry; all other
  entries are preserved. When no entry matches, no write is performed
  (idempotent no-op). Realizes [spec](../spec.md) FR-002, FR-005, FR-006. The
  cross-file write semantics and atomic single-file write are defined in the
  [configuration data contract](../../contracts/configuration.md#cross-file-write-semantics).
  Installed skills and agents are never referenced or modified
  ([ADR-0001](../architecture/ADR-0001-unregister-keeps-installed-artifacts.md)).

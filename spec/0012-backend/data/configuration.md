# Data Model: Configuration — Backend (backend.yaml)

**Spec**: [Set Backend](../spec.md)

Backend owns `backend.yaml`, the singleton backend connection. The schema —
fields, credentials, and write semantics — is owned by the
[configuration data contract](../../contracts/configuration.md#backendyaml);
this document does not restate it. There is no persisted catalog; the
available personas are a [live view](../../contracts/configuration.md#live-persona-view).

## Owns / writes

- `backend.yaml` (root fields): the connection (`kind`, `uri`, credentials, `timeout`) — a singleton, so the fields sit at the file root with no wrapper key. `set backend` upserts it; `unset backend` removes it and cascades to `personas.yaml` and (unless `--keep-artifacts`) `track.yaml` + delivered artifacts, per the contract's [cross-file write semantics](../../contracts/configuration.md#cross-file-write-semantics).

## Realizes

- `backend.yaml` connection (`kind`, `uri`, credentials) and its atomic write/teardown → [spec](../spec.md) FR-002, FR-004, FR-005, FR-006, FR-011, FR-017.

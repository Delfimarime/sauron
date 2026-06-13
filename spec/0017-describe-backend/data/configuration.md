# Data Model: Configuration — Describe Backend (backend.yaml)

**Spec**: [Describe Backend](../spec.md)

Describe Backend reads the singleton backend connection from `backend.yaml`;
it never writes. The schema is owned by the
[configuration data contract](../../contracts/configuration.md#backendyaml);
this document does not restate it. Credentials are redacted on display.

## Reads
- `backend.yaml` (root fields) — the connection (`kind`, `uri`, credentials, `timeout`, `last_synced_at`).

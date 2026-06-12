# Data Model: Configuration — Sync Personas (personas.yaml, backend.yaml)

**Spec**: [Sync Personas](../spec.md)

Sync Personas refreshes the stored definitions of the installed personas. The
schema is owned by the [configuration data contract](../../contracts/configuration.md#personasyaml);
this document does not restate it. There is no persisted catalog — the
available personas are a [live view](../../contracts/configuration.md#live-persona-view).

## Writes
- `personas.yaml` `items` definition fields (`description`, `tags`, `skills`, `agents`, `last_modified_at`, `last_synced_at`) — refreshed from the backend.
- `backend.yaml` `last_synced_at` (a root field) — on a full refresh.
- Never writes the [track file](../../contracts/configuration.md#trackyaml).

## Realizes

- `personas.yaml` definition fields → [spec](../spec.md) FR-002, FR-003 (refresh
  and stamp each installed persona's definition), FR-006 (`--force` re-pull and
  uninstall), FR-009 (no write when already up to date).
- `backend.yaml` `last_synced_at` → [spec](../spec.md) FR-004 (set on a full
  refresh).

# Data Model: Configuration — Sauron Settings (Provider)

**Spec**: [Set Provider](../spec.md)

Describes how the Set Provider feature reads and updates the persisted configuration and the tracking record.

## Active provider — `~/.sauron/settings.yaml`

- **Path**: `~/.sauron/settings.yaml` (home directory resolved per platform).
- **Format**: a single YAML document.

Top-level field:

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `provider` | string | No | `claude` | The active provider: `claude` or `zencoder`. Absent means `claude`. Realizes [spec](../spec.md) FR-002. |

Example:

```yaml
provider: zencoder
registries: []
personas: []
```

## Tracking record — `~/.sauron/track.yaml`

The installed-artifact record owned by [sync](../../0006-sync-artifacts/spec.md) (which defines its full schema). Set Provider reads it and rewrites the affected entries when migrating.

- Move (default): the entry's `provider` and `path` are rewritten to the new provider. Realizes [spec](../spec.md) FR-005.
- Copy (`--copy-only`): a new entry is added for the new provider; the previous entry is kept. Realizes [spec](../spec.md) FR-014.

## Operation

- The artifacts considered for migration are the `track.yaml` entries whose `provider` is the previous active provider. Realizes [spec](../spec.md) FR-004.
- Entries already on other providers (e.g. left by an earlier `--copy-only`) are not touched.
- With no tracked artifacts, setting the provider only updates `settings.yaml`.

## Write semantics

- `settings.yaml` (the `provider` field) and `track.yaml` (the migrated entries) are each written atomically: serialize to a temporary file in `~/.sauron/`, then rename over the destination. Both are left untouched on any failure. Realizes [spec](../spec.md) FR-008.
- When the new provider equals the current one, neither file is written (no-op). Realizes [spec](../spec.md) FR-011.

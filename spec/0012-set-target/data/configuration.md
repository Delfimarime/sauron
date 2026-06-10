# Data Model: Configuration — Sauron Settings (Target)

**Spec**: [Set Target](../spec.md)

Describes how the Set Target feature reads and updates the persisted configuration and the tracking record.

## Active target — `~/.sauron/settings.yaml`

- **Path**: `~/.sauron/settings.yaml` (home directory resolved per platform).
- **Format**: a single YAML document.

Top-level field:

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `target` | string | No | `claude` | The active provider: `claude` or `zencoder`. Absent means `claude`. Realizes [spec](../spec.md) FR-002. |

Example:

```yaml
target: zencoder
repositories: []
personas: []
```

## Tracking record — `~/.sauron/track.yaml`

The installed-artifact record owned by [sync](../../0009-sync/spec.md) (which defines its full schema). Set Target reads it and rewrites the affected entries when migrating.

- Move (default): the entry's `target` and `path` are rewritten to the new target. Realizes [spec](../spec.md) FR-005.
- Copy (`--copy-only`): a new entry is added for the new target; the previous entry is kept. Realizes [spec](../spec.md) FR-014.

## Operation

- The artifacts considered for migration are the `track.yaml` entries whose `target` is the previous active target. Realizes [spec](../spec.md) FR-004.
- Entries already on other targets (e.g. left by an earlier `--copy-only`) are not touched.
- With no tracked artifacts, setting the target only updates `settings.yaml`.

## Write semantics

- `settings.yaml` (the `target` field) and `track.yaml` (the migrated entries) are each written atomically: serialize to a temporary file in `~/.sauron/`, then rename over the destination. Both are left untouched on any failure. Realizes [spec](../spec.md) FR-008.
- When the new target equals the current one, neither file is written (no-op). Realizes [spec](../spec.md) FR-011.

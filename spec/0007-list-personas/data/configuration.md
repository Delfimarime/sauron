# Data Model: Configuration — Sauron Settings (Persona Listing)

**Spec**: [List Personas](../spec.md)

Describes the configuration that the List Personas feature reads. Listing never writes.

## Location & format

- **Path**: `~/.sauron/settings.yaml` (home directory resolved per platform).
- **Format**: a single YAML document. A missing file is treated as no personas.

## Read shape

Top-level `personas` array; each entry contributes the columns shown:

| Column | Source field | Notes |
|--------|--------------|-------|
| NAME | `name` | Identity; sortable. |
| PRIORITY | `priority` | Sortable; default sort attribute; always a defined non-negative integer. |
| TAGS | `tags` | Comma-separated; empty when absent. |
| SKILLS | `skills` | Count of entries. |
| AGENTS | `agents` | Count of entries. |

## Sorting

- `--sort` selects the attribute: `name` or `priority` (default). Realizes [spec](../spec.md) FR-013.
- `--order` selects the direction: `asc` (default) or `desc`. Realizes [spec](../spec.md) FR-014.
- Priority ordering follows [import persona ADR-0002](../../0005-import-persona/architecture/ADR-0002-unified-priority-model.md): every persona has a defined, unique priority, ordered ascending (`0` first). Realizes [spec](../spec.md) FR-003.

## Filtering

- `--search` compares its term, case-insensitively, against NAME and the persona's description. Realizes [spec](../spec.md) FR-010.
- `--tag` (repeatable) keeps only personas carrying **every** given tag. Realizes [spec](../spec.md) FR-011.
- Both filters combine with AND. Realizes [spec](../spec.md) FR-012.

# Data Model: Configuration — Sauron Settings (Persona Listing)

**Spec**: [List Personas](../spec.md)

Describes the configuration that the List Personas feature reads. Listing never
writes. The feature reads two blocks of `settings.yaml` and joins them on
persona name: the read-only `catalog` (owned by the
[backend configuration](../../0012-backend/data/configuration.md))
and the `installed` set (owned by the
[select personas configuration](../../0014-select-personas/data/configuration.md)).
Neither block is redefined here.

## Location & format

- **Path**: `~/.sauron/settings.yaml` (home directory resolved per platform).
- **Format**: a single YAML document. A missing file is treated as an empty
  catalog.

## Read shape

Each catalog entry contributes a row; a row is *installed* when the catalog
`name` also appears in the `installed` array. The columns derive from both
blocks:

| Column | Source | Notes |
|--------|--------|-------|
| NAME | `catalog[].name` | Identity; sortable; always shown and first. |
| INSTALLED | derived | `yes` when `name` is in `installed`, else `no`; sortable. |
| PRIORITY | `installed[].priority` | Sortable; default sort attribute; `-` when not installed. |
| TAGS | `catalog[].tags` | Comma-separated; empty when absent. |
| SKILLS | `catalog[].skills` | Count of entries. |
| AGENTS | `catalog[].agents` | Count of entries. |
| LAST UPDATED | `catalog[].lastModifiedAt` | Backend last-modified time; `-` when not installed. |
| LAST SYNCED | `catalog[].lastSyncedAt` | Local last-synced time; `-` when not installed. |

`LAST UPDATED` and `LAST SYNCED` are populated only for installed personas;
not-installed entries show `-`. Realizes [spec](../spec.md) FR-002, FR-003.

## Field selection

- `--fields` selects which columns are displayed and in what order; `name` is
  always present and first. Valid fields: `installed`, `priority`, `tags`,
  `skills`, `agents`, `last-updated`, `last-synced`. An unknown field is a usage
  error. Realizes [spec](../spec.md) FR-017, FR-012.

## Sorting

- `--sort` selects the attribute: `name`, `installed`, `priority` (default),
  `last-updated`, or `last-synced`. Realizes [spec](../spec.md) FR-018.
- `--order` selects the direction: `asc` (default) or `desc`. Realizes
  [spec](../spec.md) FR-019.
- Priority ordering follows the
  [unified priority model](../../AUTHORING.md#priority-model):
  installed personas have a non-negative, unique priority ordered ascending
  (`0` first); not-installed personas have no priority and are placed last
  regardless of direction. Realizes [spec](../spec.md) FR-004.

## Filtering

- `--search` compares its term, case-insensitively, against NAME and the
  persona's `description`. Realizes [spec](../spec.md) FR-013.
- `--tag` (repeatable) keeps only personas carrying **every** given tag.
  Realizes [spec](../spec.md) FR-014.
- `--installed` keeps only installed personas (`true`) or only not-installed
  personas (`false`); omitted keeps both. Realizes [spec](../spec.md) FR-015.
- These filters combine with AND. Realizes [spec](../spec.md) FR-016.

# Data Model: Configuration — Sauron Settings (Listing)

**Spec**: [List Repositories](../spec.md)

Describes the configuration that the List Repositories feature reads. Listing never writes.

## Location & format

- **Path**: `~/.sauron/settings.yaml` (home directory resolved per platform).
- **Format**: a single YAML document. A missing file is treated as no repositories.

## Read shape

Top-level `repositories` array; each entry contributes the columns shown:

| Column | Source field | Notes |
|--------|--------------|-------|
| NAME | `name` | Identity; sortable. |
| KIND | `kind` | filesystem, http, or git; sortable. |
| PRIORITY | `priority` | Sortable; default sort attribute. |
| LOCATION | `path` / `url` / `uri` | Kind-appropriate locator. Realizes [spec](../spec.md) FR-003. |

## Sorting

- `--sort` selects the attribute: `name`, `priority` (default), or `kind`. Realizes [spec](../spec.md) FR-011.
- `--order` selects the direction: `asc` (default) or `desc`. Realizes [spec](../spec.md) FR-012.

## Search

`--search` compares its term, case-insensitively, against NAME and LOCATION only. Realizes [spec](../spec.md) FR-010.

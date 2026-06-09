# Data Model: Configuration — Sauron Settings (Listing)

**Spec**: `../spec.md` (List Repositories)
**Status**: Draft

Describes the configuration that the List Repositories feature reads. Listing never writes.

## Location & format

- **Path**: `~/.sauron/settings.json` (home directory resolved per platform).
- **Format**: a single JSON document. A missing file is treated as no repositories.

## Read shape

Top-level `repositories` array; each entry contributes the columns shown:

| Column | Source field | Notes |
|--------|--------------|-------|
| NAME | `name` | Identity; sortable. |
| KIND | `kind` | filesystem, http, or git; sortable. |
| PRIORITY | `priority` | Sortable; default sort attribute. |
| LOCATION | `path` / `url` / `uri` | Kind-appropriate locator. Realizes FR-006. |

## Sorting

- `--sort` selects the attribute: `name`, `priority` (default), or `kind`. Realizes FR-004.
- `--order` selects the direction: `asc` (default) or `desc`. Realizes FR-005.

## Search

`--search` compares its term, case-insensitively, against NAME and LOCATION only. Realizes FR-003.

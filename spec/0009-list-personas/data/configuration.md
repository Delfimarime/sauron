# Data Model: Configuration — Sauron Settings (Persona Listing)

**Spec**: `../spec.md` (List Personas)
**Status**: Draft

Describes the configuration that the List Personas feature reads. Listing never writes.

## Location & format

- **Path**: `~/.sauron/settings.yaml` (home directory resolved per platform).
- **Format**: a single YAML document. A missing file is treated as no personas.

## Read shape

Top-level `personas` array; each entry contributes the columns shown:

| Column | Source field | Notes |
|--------|--------------|-------|
| NAME | `name` | Identity; sortable. |
| PRIORITY | `priority` | Sortable; default sort attribute; shown as `-` when undefined. |
| TAGS | `tags` | Comma-separated; empty when absent. |
| SKILLS | `skills` | Count of entries. |
| AGENTS | `agents` | Count of entries. |

## Sorting

- `--sort` selects the attribute: `name` or `priority` (default). Realizes FR-006.
- `--order` selects the direction: `asc` (default) or `desc`. Realizes FR-007.
- Priority ordering follows `0007-import-persona` ADR-0001: defined values first, ascending (`0` first); undefined priorities after all defined ones, ordered by name among themselves. Realizes FR-008.

## Filtering

- `--search` compares its term, case-insensitively, against NAME and the persona's description. Realizes FR-003.
- `--tag` (repeatable) keeps only personas carrying **every** given tag. Realizes FR-004.
- Both filters combine with AND. Realizes FR-005.

# Data Model: Configuration — Sauron Settings (Installed Personas)

**Spec**: [Select Personas](../spec.md)

Describes how the Select Personas feature reads and writes the persisted
configuration. The feature owns the `installed` block of `settings.yaml`; it
references catalog entries owned by the
[backend configuration](../../0012-backend/data/configuration.md)
and does not redefine the catalog.

## Location & format

- **Path**: `~/.sauron/settings.yaml` (home directory resolved per platform).
- **Format**: a single YAML document. A missing file is treated as an empty
  installed set.

## The `installed` block

`settings.yaml` carries an `installed` array. Each element records one installed
persona:

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | The persona's name. Must match a `name` in the catalog (owned by the [backend](../../0012-backend/data/configuration.md)). |
| `priority` | integer | Non-negative, unique across the `installed` array. Assigned positionally by `set persona` argument order — the first listed persona is `0`, the next `1`, and so on (see the [unified priority model](../../AUTHORING.md#priority-model)). |

```yaml
installed:
  - name: platform
    priority: 0
  - name: security
    priority: 1
  - name: data
    priority: 2
```

Each `name` references a catalog persona; the catalog's full schema (definitions
pulled from the backend) is owned by
[backend configuration](../../0012-backend/data/configuration.md)
and is not duplicated here. Uninstalling a persona removes its element from
`installed` only; the referenced catalog entry is left untouched.

## Operation

- `set persona <name>...` replaces the entire `installed` array with one element
  per given name, in argument order, assigning `priority` `0, 1, 2, …` by
  position. Personas previously in `installed` but not listed are removed.
  Realizes [spec](../spec.md) FR-004, FR-005, FR-007.
- Every given name is validated against the catalog before any write; if any
  name is absent the array is left unchanged. Realizes
  [spec](../spec.md) FR-011, FR-013.
- `unset persona X Y` removes the elements whose `name` matches; `unset persona`
  with no name empties the array. Catalog entries are not affected. Realizes
  [spec](../spec.md) FR-008, FR-009.
- An `unset persona` for a name not present in `installed` performs no change.
  Realizes [spec](../spec.md) FR-014.

## Write semantics

- The whole document is loaded, the `installed` array rebuilt, and the document
  written back only after all validation passes. The file is left untouched on
  any failure. Realizes [spec](../spec.md) FR-010, FR-015.
- Writes are atomic: serialize to a temporary file in `~/.sauron/`, then rename
  over `settings.yaml`.
- When the operation results in no change (an `unset` of an uninstalled
  persona), no write is performed. Realizes [spec](../spec.md) FR-014.

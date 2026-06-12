# Data Model: Configuration — Sauron Settings (Persona Description)

**Spec**: [Describe Persona](../spec.md)

Describes the configuration that the Describe Persona feature reads. Describing
never writes and defines no schema of its own. The feature reads two blocks of
`settings.yaml` and joins them on persona name: the read-only `catalog` (owned
by the [backend configuration](../../0012-backend/data/configuration.md)) and
the `installed` set (owned by the
[select personas configuration](../../0014-select-personas/data/configuration.md)).
Neither block is redefined here.

## Location & format

- **Path**: `~/.sauron/settings.yaml` (home directory resolved per platform).
- **Format**: a single YAML document. A missing file is treated as an empty
  catalog, so any name resolves as not found.

## Read shape

The persona named by `<name>` is resolved against the `catalog`; it is
*installed* when that name also appears in the `installed` set. Its fields derive
from both blocks. `priority`, `last-updated`, and `last-synced` are populated
only for an installed persona and are empty otherwise. Realizes
[spec](../spec.md) FR-002, FR-003, FR-004.

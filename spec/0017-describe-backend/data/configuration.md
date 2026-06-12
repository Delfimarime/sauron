# Data Model: Configuration — Sauron Settings (Describe Backend)

**Spec**: [Describe Backend](../spec.md)

Describes the configuration that the Describe Backend feature reads. This feature
defines **no schema of its own**: it is read-only and reads the singleton
backend (`personaRegistry`) and the `catalog` mirror whose schema is owned by
[backend](../../0012-backend/data/configuration.md). The `installed` count is
derived from the install records owned by
[select personas](../../0014-select-personas/spec.md).

## Location & format

- **Path**: `~/.sauron/settings.yaml` (home directory resolved per platform).
- **Format**: a single YAML document.

## Write semantics

This feature never writes the settings or the
[track file](../../0006-sync-artifacts/data/configuration.md). When no backend is
configured, it reports the absence and exits successfully without writing
anything. Realizes [spec](../spec.md) FR-002, FR-008.

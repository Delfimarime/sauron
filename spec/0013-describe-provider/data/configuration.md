# Describe Provider — configuration

This feature **reads** `settings.yaml` and writes nothing. The document schema is
owned by the [configuration data contract](../../contracts/configuration.md).

## Reads

- File: `settings.yaml`, matching the single `Provider` document.

## Field realization

| Field | Requirement |
|---|---|
| `Provider.metadata.name` | FR-001, FR-002, FR-003 |

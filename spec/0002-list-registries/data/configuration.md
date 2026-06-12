# Data Model: Configuration — List Registries (registries.yaml)

**Spec**: [List Registries](../spec.md)

This feature reads `registries.yaml`'s `items` block — each entry's
`name`, `kind`, `priority`, and `uri` — to render the listing; it never writes.
The schema is owned by the
[configuration data contract](../../contracts/configuration.md#registriesyaml);
this document does not restate it.

## Reads

- `registries.yaml` `items`: `name`, `kind`, `priority`, `uri` — the
  columns shown by the listing. `--search` matches case-insensitively against
  `name` and `uri`; `--sort` orders by `name`, `priority`, or `kind`.

## Owns

- Nothing.

## Writes

- Nothing. Listing is read-only.

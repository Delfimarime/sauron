# Describe Registry

**Type:** feature

**Depends on:** [add registry](../0001-add-registry/spec.md)

## Overview

A person responsible for a team's agentic-AI setup needs to inspect one
registry in full — its kind, uri, priority, and how it authenticates — so
that they can review or troubleshoot a single configured source without scanning
the whole list. Registries are registered by
[add registry](../0001-add-registry/spec.md); describing is the single-record
counterpart to [list registries](../0002-list-registries/spec.md). It is
read-only and offline: it never contacts the registry or any other external
resource, and never writes the configuration or the track file.

## Requirements

### Ubiquitous

- **FR-001**: Sauron shall provide the ability to describe a single registered
  registry by name.
- **FR-002**: Sauron shall treat describing a registry as read-only and offline,
  contacting no external resource and writing neither the configuration nor the
  track file.

### Event-driven

- **FR-003**: When a user describes a registry, Sauron shall read the named
  registry from `registries.yaml` and display its fields, one per line as
  `field: value`, with the identity field `name` first.
- **FR-004**: When displaying a registry's fields, Sauron shall present `name`
  as the identity and `kind`, `uri`, `priority`, `auth`, `tls`, `timeout`,
  and `ssh-key` as the available fields, where `uri` is the registry's source
  location.
- **FR-005**: When a field does not apply to the resolved registry's kind,
  Sauron shall show that field empty rather than rejecting the request — `tls`
  and `timeout` apply to `http`, `ssh-key` and `timeout` to `git`, and
  `filesystem` carries neither auth nor TLS.
- **FR-006**: When displaying `auth`, Sauron shall print the credential's
  `${env:VAR}` reference and shall never print a resolved secret, rendering any
  resolved secret value as `REDACTED`.
- **FR-007**: When a field has no value, Sauron shall show it with an empty
  value.

### Unwanted behavior

- **FR-008**: If the `<name>` argument is missing, then Sauron shall exit with
  code 2 without executing the command and report that a registry name is
  required.
- **FR-009**: If no registry with the given name is registered, then Sauron
  shall reject the request and report that the registry was not found.
- **FR-010**: If `registries.yaml` does not exist, then Sauron shall treat the
  registry as not found.
- **FR-011**: If `registries.yaml` exists but cannot be read or parsed, then
  Sauron shall reject the request and report that the configuration cannot be
  read.
- **FR-012**: If `--fields` names a field other than `name`, `kind`,
  `uri`, `priority`, `auth`, `tls`, `timeout`, or `ssh-key`, then Sauron
  shall exit with code 2 without executing the command and report the allowed
  fields.

### Optional

- **FR-013**: Where `--fields` is provided, Sauron shall display only the named
  fields in the given order, with the identity field `name` always present and
  first; where it is omitted, Sauron shall display the form's full field set.

## Key Entities

- **Registry**: a registered source of artifacts (see
  [add registry](../0001-add-registry/spec.md)), described by its identity
  `name` and the fields `kind`, `uri`, `priority`, `auth`, `tls`,
  `timeout`, and `ssh-key`. The `uri` is the registry's source location. Fields
  that do not apply to the resolved kind are shown empty; `auth` shows only the
  `${env:VAR}` reference and never a resolved secret.

## Notes

Configuration is now split across files per the
[configuration data contract](../contracts/configuration.md); file references
updated accordingly.

# Data Model: Configuration — Sauron Settings (Backend)

**Spec**: [Backend](../spec.md)

Describes the persisted configuration that the Backend feature reads
and writes, covering all backend kinds
([http](../capabilities/http.md), [filesystem](../capabilities/filesystem.md),
[git](../capabilities/git.md)). It defines two blocks: the singleton
`personaRegistry` config and the read-only `catalog` mirror.

## Location & format

- **Path**: `~/.sauron/settings.yaml` (home directory resolved per platform).
- **Format**: a single YAML document.
- **Lifecycle**: created on first successful write if absent.

## Schema

Top-level document (the blocks this feature owns):

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `personaRegistry` | object (singleton) | No | The configured backend; absent when none is configured. Realizes [spec](../spec.md) FR-002, FR-005. |
| `catalog` | array of Persona Definition | No | The local read-only mirror of persona definitions. Empty or absent until [sync personas](../../0013-sync-personas/spec.md) populates it. Realizes [spec](../spec.md) FR-008. |

### `personaRegistry` block (singleton)

Common fields (all kinds):

| Field | Type | Required | Constraints | Description |
|-------|------|----------|-------------|-------------|
| `kind` | string | Yes | enum: `http`, `filesystem`, `git` | Backend kind; selects which kind-scoped fields apply. Realizes [spec](../spec.md) FR-004. |
| `location` | string | Yes | per kind (`url`/`path`/`uri` form) | Where persona definitions live. Realizes [spec](../spec.md) FR-004, FR-017. |
| `auth` | object | No | http only; see below | HTTP Basic credentials (env references). |
| `tls` | object | No | http only; see below | TLS / mutual-TLS settings. |
| `ssh` | object | No | git only; see below | SSH authentication settings. |
| `timeout` | string | No | duration; default `30s` | Bounds network operations (the `HEAD` probe, `git ls-remote`, fetches). Realizes [http](../capabilities/http.md) FR-005, [git](../capabilities/git.md) FR-004. |
| `lastSyncedAt` | string | No | RFC 3339 timestamp | When the catalog was last refreshed from this backend. Set by [sync personas](../../0013-sync-personas/spec.md). |

`auth` object (`http`):

| Field | Type | Required | Constraints | Description |
|-------|------|----------|-------------|-------------|
| `type` | string | Yes | `"basic"` | Auth scheme. |
| `username` | string | No | `${env:VAR}` reference | Resolved from the environment at use time. Realizes [http](../capabilities/http.md) FR-004. |
| `password` | string | No | `${env:VAR}` reference | Resolved from the environment at use time. Realizes [http](../capabilities/http.md) FR-004. |

`tls` object (`http`):

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `skipVerify` | boolean | No | false | Skip server cert verification. |
| `caCert` | string | No | — | Path to a CA bundle. |
| `clientCert` | string | No | — | Path to the client certificate (mutual TLS). |
| `clientKey` | string | No | — | Path to the client key (mutual TLS). |

`ssh` object (`git`):

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `keyPath` | string | No | Path to the private key used to authenticate. Omitted = system SSH credentials. Realizes [git](../capabilities/git.md) FR-008. |

### `catalog` block (read-only mirror)

Each Persona Definition entry:

| Field | Type | Required | Constraints | Description |
|-------|------|----------|-------------|-------------|
| `name` | string | Yes | unique within the catalog | Persona identity. |
| `description` | string | No | — | Human-readable summary. |
| `tags` | array of string | No | — | Free-form labels for search and grouping. |
| `skills` | array of string | No | — | Skill artifacts the persona bundles. |
| `agents` | array of string | No | — | Agent artifacts the persona bundles. |
| `lastModifiedAt` | string | No | RFC 3339 timestamp | The backend's per-persona last-modified time (git: last commit touching the persona; http: `Last-Modified`/index metadata; filesystem: file mtime). Recorded by [sync personas](../../0013-sync-personas/spec.md). |
| `lastSyncedAt` | string | No | RFC 3339 timestamp | When this entry was last pulled into the local catalog. |

The catalog is **read-only** with respect to this feature: it is populated and
refreshed exclusively by [sync personas](../../0013-sync-personas/spec.md), and
`set backend` never writes catalog entries. Which catalog personas are
*installed* is a separate concern owned by
[select personas](../../0014-select-personas/spec.md) and is not defined here.

## Credentials & transport

- Per [ADR-0001](../../0001-add-registry/architecture/ADR-0001-credentials-via-env-only.md),
  `auth.username` / `auth.password` hold only an `${env:VAR}` reference; no
  secret is ever written to disk. TLS cert/key fields store **file paths**, not
  the certificate material.
- Per [ADR-0002](../../0001-add-registry/architecture/ADR-0002-ssh-only-remotes.md),
  only SSH git URIs are supported. `ssh.keyPath` stores a **file path**, not key
  material; when absent, the system's regular SSH credentials are used.

## Write semantics

- `set backend` loads the whole document, replaces the
  `personaRegistry` block, and writes it back only after all validation passes;
  the file is left untouched on any failure. The catalog is not modified.
  Realizes [spec](../spec.md) FR-005, FR-006, FR-011.
- `unset backend` loads the whole document, removes the
  `personaRegistry` block and the `catalog` array, and (unless
  `--keep-artifacts`) the corresponding install records and track-file entries,
  then writes back. Realizes [spec](../spec.md) FR-008, FR-009, FR-010.
- Writes are atomic: serialize to a temporary file in `~/.sauron/`, then rename
  over `settings.yaml`.

## Example

```yaml
personaRegistry:
  kind: http
  location: https://secure-personas.example.com
  auth:
    type: basic
    username: ${env:PERSONAS_USER}
    password: ${env:PERSONAS_PASS}
  tls:
    skipVerify: false
    caCert: /home/user/.sauron/ca.pem
  timeout: 30s
  lastSyncedAt: 2026-06-12T09:30:00Z
catalog:
  - name: backend-developer
    description: Backend service development persona.
    tags: [backend, go]
    skills:
      - design-oas3
      - code-review
    agents:
      - software-engineer
    lastModifiedAt: 2026-06-11T18:00:00Z
    lastSyncedAt: 2026-06-12T09:30:00Z
  - name: qa-engineer
    description: Test-authoring and quality persona.
    tags: [qa]
    skills:
      - test-plan
    agents: []
    lastModifiedAt: 2026-06-10T12:00:00Z
    lastSyncedAt: 2026-06-12T09:30:00Z
```

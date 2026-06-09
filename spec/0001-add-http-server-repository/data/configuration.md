# Data Model: Configuration — Sauron Settings (HTTP Repository)

**Spec**: `../spec.md` (Add HTTP Server Repository)
**Status**: Draft

Describes the persisted configuration that the Add HTTP Server Repository feature reads and writes.

## Location & format

- **Path**: `~/.sauron/settings.json` (home directory resolved per platform).
- **Format**: a single JSON document.
- **Lifecycle**: created on first successful write if absent.

## Schema

Top-level document:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `repositories` | array of Repository | Yes | Registered sources. Empty array when none. |

HTTP Repository entry:

| Field | Type | Required | Constraints | Description |
|-------|------|----------|-------------|-------------|
| `kind` | string | Yes | enum: `http` | Repository kind. Realizes FR-004. |
| `name` | string | Yes | slug; unique across all kinds | Repository identity. Realizes FR-008, FR-018. |
| `priority` | integer | Yes | positive; unique across all kinds; lower = higher | Ordering. Realizes FR-017, FR-019. |
| `url` | string | Yes | http/https URL | Server location. May repeat. Realizes FR-005. |
| `auth` | object | No | see below | HTTP Basic credentials (env references). |
| `tls` | object | No | see below | TLS / mutual-TLS settings. |
| `timeout` | string | No | duration; default 30s | Bounds the HEAD probe and fetches. Realizes FR-025. |

`auth` object:

| Field | Type | Required | Constraints | Description |
|-------|------|----------|-------------|-------------|
| `type` | string | Yes | `"basic"` | Auth scheme. |
| `username` | string | No | `${env:VAR}` reference | Resolved from the environment at use time. Realizes FR-007. |
| `password` | string | No | `${env:VAR}` reference | Resolved from the environment at use time. Realizes FR-007. |

`tls` object:

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `skipVerify` | boolean | No | false | Skip server cert verification. |
| `caCert` | string | No | — | Path to a CA bundle. |
| `clientCert` | string | No | — | Path to the client certificate (mutual TLS). |
| `clientKey` | string | No | — | Path to the client key (mutual TLS). |

## Identity

A repository is identified by its `name`. `name` and `priority` are each unique across all repositories regardless of kind. `url` is not an identity key — entries may share a URL.

## Credentials

Per ADR-0001, `auth.username` / `auth.password` hold only an `${env:VAR}` reference; no secret is ever written to disk. TLS cert/key fields store **file paths**, not the certificate material.

## Write semantics

- The whole document is loaded, the new entry appended, and written back only after all validation passes. The file is left untouched on any failure.
- Writes are atomic: serialize to a temporary file in `~/.sauron/`, then rename over `settings.json`.

## Example

```json
{
  "repositories": [
    {
      "kind": "http",
      "name": "team-secure",
      "priority": 2,
      "url": "https://secure.example.com",
      "auth": {
        "type": "basic",
        "username": "${env:SKILLS_USER}",
        "password": "${env:SKILLS_PASS}"
      },
      "tls": {
        "skipVerify": false,
        "caCert": "/home/user/.sauron/ca.pem",
        "clientCert": "/home/user/.sauron/client.pem",
        "clientKey": "/home/user/.sauron/client.key"
      },
      "timeout": "30s"
    }
  ]
}
```

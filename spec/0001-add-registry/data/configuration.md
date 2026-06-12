# Data Model: Configuration — Sauron Settings (Registry)

**Spec**: [Add Registry](../spec.md)

Describes the persisted configuration that the Add Registry feature reads
and writes, covering all registry kinds
([http](../capabilities/http.md), [filesystem](../capabilities/filesystem.md),
[git](../capabilities/git.md)).

## Location & format

- **Path**: `~/.sauron/settings.yaml` (home directory resolved per platform).
- **Format**: a single YAML document.
- **Lifecycle**: created on first successful write if absent.

## Schema

Top-level document:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `registries` | array of Registry | Yes | Registered sources. Empty array when none are registered. |

Registry entry — common fields (all kinds):

| Field | Type | Required | Constraints | Description |
|-------|------|----------|-------------|-------------|
| `kind` | string | Yes | enum: `http`, `filesystem`, `git` | Registry kind; selects which kind-scoped fields apply. Realizes [spec](../spec.md) FR-004. |
| `name` | string | Yes | slug `^[a-z0-9]+(-[a-z0-9]+)*$`; unique across all kinds | Registry identity. Realizes [spec](../spec.md) FR-002, FR-005, FR-016. |
| `priority` | integer | No | non-negative; unique across all kinds; `0` for the first registry, `max + 1` when omitted on a later add; lower = higher precedence. See the [unified priority model](../../AUTHORING.md#priority-model). | Registry ordering. Realizes [spec](../spec.md) FR-003, FR-009, FR-010, FR-017. |

Kind-scoped fields — `http` ([http](../capabilities/http.md)):

| Field | Type | Required | Constraints | Description |
|-------|------|----------|-------------|-------------|
| `url` | string | Yes | `http`/`https` URL | Server location. May repeat across entries. Realizes [http](../capabilities/http.md) FR-002. |
| `auth` | object | No | see below | HTTP Basic credentials (env references). |
| `tls` | object | No | see below | TLS / mutual-TLS settings. |
| `timeout` | string | No | duration; default `30s` | Bounds the `HEAD` probe and fetches. Realizes [http](../capabilities/http.md) FR-005. |

`auth` object:

| Field | Type | Required | Constraints | Description |
|-------|------|----------|-------------|-------------|
| `type` | string | Yes | `"basic"` | Auth scheme. |
| `username` | string | No | `${env:VAR}` reference | Resolved from the environment at use time. Realizes [http](../capabilities/http.md) FR-004. |
| `password` | string | No | `${env:VAR}` reference | Resolved from the environment at use time. Realizes [http](../capabilities/http.md) FR-004. |

`tls` object:

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `skipVerify` | boolean | No | false | Skip server cert verification. |
| `caCert` | string | No | — | Path to a CA bundle. |
| `clientCert` | string | No | — | Path to the client certificate (mutual TLS). |
| `clientKey` | string | No | — | Path to the client key (mutual TLS). |

Kind-scoped fields — `filesystem`
([filesystem](../capabilities/filesystem.md)):

| Field | Type | Required | Constraints | Description |
|-------|------|----------|-------------|-------------|
| `path` | string | Yes | absolute, symlink-resolved | Directory location. May repeat across entries. Realizes [filesystem](../capabilities/filesystem.md) FR-004. |

Kind-scoped fields — `git` ([git](../capabilities/git.md)):

| Field | Type | Required | Constraints | Description |
|-------|------|----------|-------------|-------------|
| `uri` | string | Yes | SSH-based git URI | Remote location. May repeat across entries. Realizes [git](../capabilities/git.md) FR-002. |
| `ssh` | object | No | see below | SSH authentication settings. |
| `timeout` | string | No | duration; default `30s` | Bounds network operations. Realizes [git](../capabilities/git.md) FR-004. |

`ssh` object:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `keyPath` | string | No | Path to the private key used to authenticate. Omitted = system SSH credentials. Realizes [git](../capabilities/git.md) FR-009. |

## Identity

A registry is identified by its `name`. `name` and `priority` are each
unique across all registries regardless of kind; `priority` is always a
defined, non-negative integer per the
[unified priority model](../../AUTHORING.md#priority-model).
The location field (`url`, `path`, or `uri`) is not an identity key — entries
may share a location. Realizes [spec](../spec.md) FR-018.

## Credentials & transport

- Per [ADR-0001](../architecture/ADR-0001-credentials-via-env-only.md),
  `auth.username` / `auth.password` hold only an `${env:VAR}` reference; no
  secret is ever written to disk. TLS cert/key fields store **file paths**,
  not the certificate material.
- Per [ADR-0002](../architecture/ADR-0002-ssh-only-remotes.md), only SSH git
  URIs are supported. `ssh.keyPath` stores a **file path**, not key material;
  when absent, the system's regular SSH credentials are used.

## Write semantics

- The whole document is loaded, the new entry appended, and written back only
  after all validation passes. The file is left untouched on any failure.
  Realizes [spec](../spec.md) FR-006, FR-011.
- Writes are atomic: serialize to a temporary file in `~/.sauron/`, then
  rename over `settings.yaml`.

## Example

```yaml
registries:
  - kind: http
    name: team-secure
    priority: 2
    url: https://secure.example.com
    auth:
      type: basic
      username: ${env:SKILLS_USER}
      password: ${env:SKILLS_PASS}
    tls:
      skipVerify: false
      caCert: /home/user/.sauron/ca.pem
      clientCert: /home/user/.sauron/client.pem
      clientKey: /home/user/.sauron/client.key
    timeout: 30s
  - kind: filesystem
    name: team-skills
    priority: 1
    path: /home/user/team-skills
  - kind: git
    name: team-deploy
    priority: 3 # added without --priority; appended at the end (max + 1)
    uri: ssh://git@github.com/acme/agents.git
    ssh:
      keyPath: /home/user/.ssh/deploy_ed25519
    timeout: 30s
```

Every `priority` is stored as a concrete non-negative integer; an entry added
without `--priority` is persisted with its assigned value (`0` for the first
registry, `max + 1` for a later one), never left blank.

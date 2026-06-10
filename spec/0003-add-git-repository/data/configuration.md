# Data Model: Configuration — Sauron Settings (Git Repository)

**Spec**: `../spec.md` (Add Git Repository)
**Status**: Draft

Describes the persisted configuration that the Add Git Repository feature reads and writes.

## Location & format

- **Path**: `~/.sauron/settings.yaml` (home directory resolved per platform).
- **Format**: a single YAML document.
- **Lifecycle**: created on first successful write if absent.

## Schema

Top-level document:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `repositories` | array of Repository | Yes | Registered sources. Empty array when none. |

Git Repository entry:

| Field | Type | Required | Constraints | Description |
|-------|------|----------|-------------|-------------|
| `kind` | string | Yes | enum: `git` | Repository kind. Realizes FR-004. |
| `name` | string | Yes | slug; unique across all kinds | Repository identity. Realizes FR-008, FR-019. |
| `priority` | integer | Yes | positive; unique across all kinds; lower = higher | Ordering. Realizes FR-018, FR-020. |
| `uri` | string | Yes | SSH-based git URI | Remote location. May repeat. Realizes FR-005. |
| `ssh` | object | No | see below | SSH authentication settings. |
| `timeout` | string | No | duration; default 30s | Bounds network operations. Realizes FR-011. |

`ssh` object:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `keyPath` | string | No | Path to the private key used to authenticate. Omitted = system SSH credentials. Realizes FR-007. |

## Identity

A repository is identified by its `name`. `name` and `priority` are each unique across all repositories regardless of kind. `uri` is not an identity key — entries may share a URI.

## Transport & credentials

Per ADR-0001, only SSH git URIs are supported. `ssh.keyPath` stores a **file path**, not key material; when absent, the system's regular SSH credentials are used.

## Write semantics

- The whole document is loaded, the new entry appended, and written back only after all validation passes. The file is left untouched on any failure.
- Writes are atomic: serialize to a temporary file in `~/.sauron/`, then rename over `settings.yaml`.

## Example

```yaml
repositories:
  - kind: git
    name: team-deploy
    priority: 2
    uri: ssh://git@github.com/acme/agents.git
    ssh:
      keyPath: /home/user/.ssh/deploy_ed25519
    timeout: 30s
```

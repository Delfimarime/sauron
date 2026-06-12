# Configuration Data Contract

The compiled data contract for everything Sauron persists. It owns the schema of
every configuration file under `~/.sauron/`, the sub-schemas they share, and the
key and write conventions they obey. A feature's `data/configuration.md`
declares which file and fields the feature owns or writes — and the requirements
(`FR-NNN`) that govern them — and links here for the schema; this contract does
not depend on any feature.

The machine-readable JSON Schemas for these files live under
[schemas/](schemas/), one per file, and are the normative validation source; the
tables below are their human-readable form.

## Files

Sauron's state is split by concern. Each file is an independent YAML document and
is written independently.

| File | Owned by | Holds | Schema |
|---|---|---|---|
| `registries.yaml` | [add registry](../0001-add-registry/spec.md) | the registered artifact sources | [registries.schema.json](schemas/registries.schema.json) |
| `backend.yaml` | [backend](../0012-backend/spec.md) | the singleton backend connection | [backend.schema.json](schemas/backend.schema.json) |
| `personas.yaml` | [select personas](../0014-select-personas/spec.md) | the installed personas, with their definitions | [personas.schema.json](schemas/personas.schema.json) |
| `track.yaml` | [sync artifacts](../0006-sync-artifacts/spec.md) | the installed artifacts and their provenance | [track.schema.json](schemas/track.schema.json) |
| `settings.yaml` | [set provider](../0009-set-provider/spec.md), [schedule artifact sync](../0011-schedule-sync/spec.md), [schedule persona sync](../0019-schedule-sync-personas/spec.md) | global settings: the active provider and the sync schedules | [settings.schema.json](schemas/settings.schema.json) |

- **Path**: each file lives at `~/.sauron/<file>` (home directory resolved per
  platform).
- **Lifecycle**: a file is created on the first successful write that needs it; a
  missing file is read as its empty state (no registries, no backend, no
  installed personas, nothing tracked, defaults for global settings).

There is **no persisted catalog**. The set of *available* personas is computed
live (see [Live persona view](#live-persona-view)).

## Conventions

- **Keys are `snake_case`.** Multi-word keys use underscores
  (`skip_verify`, `ca_cert`, `client_cert`, `client_key`, `key_path`,
  `last_modified_at`, `last_synced_at`); single-word keys are bare (`kind`,
  `name`, `priority`, `uri`, `timeout`, `provider`, `schedules`, `version`, `items`,
  `auth`, `tls`, `ssh`, `type`, `username`, `password`, `tags`, `skills`,
  `agents`, `description`, `path`, `registry`, `persona`).
- **Collections are `items`.** A file that holds a list keeps it in a top-level
  `items` array (`registries.yaml`, `personas.yaml`, `track.yaml`). Singleton
  files (`backend.yaml`, `settings.yaml`) carry their fields at the root.
- **Versioning.** Every file carries a top-level `version` (integer, starts at
  `1`) so a future schema change can be detected and migrated.
- **Atomic single-file writes.** A write serializes the whole document to a
  temporary file in `~/.sauron/`, then renames it over the destination; the file
  is left untouched on any failure.

## Cross-file write semantics

A single atomic rename covers one file, not a multi-file operation. The
operations that touch more than one file are written in a fixed order and are
**idempotent**: a run interrupted between two file writes is fully repaired by
re-running the same command.

- **set provider** writes `track.yaml` (migrated entries) **then** `settings.yaml`
  (`provider`). If interrupted after `track.yaml`, re-running `set provider` with
  the same target migrates nothing further and completes `settings.yaml`.
- **unset backend** writes `track.yaml` and removes delivered artifacts (unless
  `--keep-artifacts`), **then** `personas.yaml` (empties its `items`), **then**
  `backend.yaml`. Each step is a no-op on a re-run once already applied.

## Shared sub-schemas

Used by both [registries.yaml](#registriesyaml) entries and the
[backend.yaml](#backendyaml) connection.

`auth` (http kinds) — HTTP Basic credentials held as environment references:

| Field | Type | Required | Constraints | Description |
|---|---|---|---|---|
| `type` | string | Yes | `"basic"` | Auth scheme. |
| `username` | string | No | `${env:VAR}` reference | Resolved from the environment at use time. |
| `password` | string | No | `${env:VAR}` reference | Resolved from the environment at use time. |

`tls` (http kinds):

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `skip_verify` | boolean | No | false | Skip server cert verification. |
| `ca_cert` | string | No | — | Path to a CA bundle. |
| `client_cert` | string | No | — | Path to the client certificate (mutual TLS). |
| `client_key` | string | No | — | Path to the client key (mutual TLS). |

`ssh` (git kind):

| Field | Type | Required | Description |
|---|---|---|---|
| `key_path` | string | No | Path to the private key used to authenticate; omitted = system SSH credentials. |

**Credentials & transport (binding for both files):** per
[ADR-0001](../0001-add-registry/architecture/ADR-0001-credentials-via-env-only.md),
`auth.username` / `auth.password` hold only an `${env:VAR}` reference; no secret
is ever written to disk, and `tls` cert/key fields store **file paths**, not
material. Per
[ADR-0002](../0001-add-registry/architecture/ADR-0002-ssh-only-remotes.md), only
SSH git URIs are supported; `ssh.key_path` stores a **file path**, not key
material.

## settings.yaml

Global settings. Owned by [set provider](../0009-set-provider/spec.md)
(`provider`), [schedule artifact sync](../0011-schedule-sync/spec.md)
(`schedules.sync_artifacts`), and
[schedule persona sync](../0019-schedule-sync-personas/spec.md)
(`schedules.sync_personas`).
Schema: [settings.schema.json](schemas/settings.schema.json).

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `version` | integer | Yes | `1` | Schema version. |
| `provider` | string | No | `claude` | The active provider: `claude` or `zencoder`. Absent means `claude`. |
| `schedules` | object | No | — | The recorded sync schedules, keyed by operation; absent when nothing is scheduled. |

`schedules` object — one optional key per scheduled operation:

| Field | Type | Required | Description |
|---|---|---|---|
| `sync_artifacts` | object | No | The `sauron sync artifacts` schedule. |
| `sync_personas` | object | No | The `sauron sync personas` schedule. |

Each schedule object:

| Field | Type | Required | Description |
|---|---|---|---|
| `expression` | string | Yes | The cron expression Sauron installs into the OS crontab. |

```yaml
version: 1
provider: zencoder
schedules:
  sync_artifacts:
    expression: "0 * * * *"
  sync_personas:
    expression: "0 6 * * *"
```

## registries.yaml

The registered artifact sources. Owned by
[add registry](../0001-add-registry/spec.md).
Schema: [registries.schema.json](schemas/registries.schema.json).

| Field | Type | Required | Description |
|---|---|---|---|
| `version` | integer | Yes | Schema version. |
| `items` | array of Registry | Yes | Registered sources; empty array when none. |

Registry entry — common fields (all kinds):

| Field | Type | Required | Constraints | Description |
|---|---|---|---|---|
| `name` | string | Yes | slug `^[a-z0-9]+(-[a-z0-9]+)*$`; unique across all kinds | Registry identity. |
| `kind` | string | Yes | enum: `http`, `filesystem`, `git` | Registry kind; selects the transport and which credentials apply. |
| `priority` | integer | Yes | non-negative; unique; `0` for the first registry, `max + 1` when omitted on a later add; lower = higher precedence. See the [priority model](../AUTHORING.md#priority-model). | Registry ordering. |
| `uri` | string | Yes | kind-shaped: `http`/`https` URL (`http`), absolute symlink-resolved path (`filesystem`), SSH git URI (`git`) | Source location. Not an identity key — entries may share a `uri`. |

Kind-scoped fields:

| Kind | Adds | From sub-schema |
|---|---|---|
| `http` | `auth`, `tls`, `timeout` (duration, default `30s`) | [`auth`](#shared-sub-schemas), [`tls`](#shared-sub-schemas) |
| `filesystem` | — | — |
| `git` | `ssh`, `timeout` (duration, default `30s`) | [`ssh`](#shared-sub-schemas) |

A registry is identified by `name`; `name` and `priority` are each unique across
all registries regardless of kind.

```yaml
version: 1
items:
  - name: team-secure
    kind: http
    priority: 2
    uri: https://secure.example.com
    auth:
      type: basic
      username: ${env:SKILLS_USER}
      password: ${env:SKILLS_PASS}
    tls:
      skip_verify: false
      ca_cert: /home/user/.sauron/ca.pem
    timeout: 30s
  - name: team-skills
    kind: filesystem
    priority: 1
    uri: /home/user/team-skills
  - name: team-deploy
    kind: git
    priority: 3
    uri: ssh://git@github.com/acme/agents.git
    ssh:
      key_path: /home/user/.ssh/deploy_ed25519
    timeout: 30s
```

## backend.yaml

The singleton backend connection — the persona analog of a registry. Because it
is a single instance, its connection fields sit at the **root** of the file (no
wrapper key). Owned by [backend](../0012-backend/spec.md).
Schema: [backend.schema.json](schemas/backend.schema.json).

When no backend is configured, `backend.yaml` is absent (or carries only
`version`).

| Field | Type | Required | Constraints | Description |
|---|---|---|---|---|
| `version` | integer | Yes | — | Schema version. |
| `kind` | string | Yes | enum: `http`, `filesystem`, `git` | Backend kind. |
| `uri` | string | Yes | kind-shaped, as for a registry's `uri` | Where persona definitions live. |
| `auth` | object | No | http only | [`auth`](#shared-sub-schemas) sub-schema. |
| `tls` | object | No | http only | [`tls`](#shared-sub-schemas) sub-schema. |
| `ssh` | object | No | git only | [`ssh`](#shared-sub-schemas) sub-schema. |
| `timeout` | string | No | http/git; duration, default `30s` | Bounds network operations. |
| `last_synced_at` | string | No | RFC 3339 timestamp | When persona definitions were last refreshed from this backend, by [sync personas](../0013-sync-personas/spec.md). |

```yaml
version: 1
kind: http
uri: https://secure-personas.example.com
auth:
  type: basic
  username: ${env:PERSONAS_USER}
  password: ${env:PERSONAS_PASS}
tls:
  skip_verify: false
  ca_cert: /home/user/.sauron/ca.pem
timeout: 30s
last_synced_at: 2026-06-12T09:30:00Z
```

## personas.yaml

The installed personas, each stored **with its full definition** so that
[sync artifacts](../0006-sync-artifacts/spec.md) works without contacting the
backend. Owned by [select personas](../0014-select-personas/spec.md).
Schema: [personas.schema.json](schemas/personas.schema.json).

| Field | Type | Required | Description |
|---|---|---|---|
| `version` | integer | Yes | Schema version. |
| `items` | array of Installed Persona | Yes | The installed personas; empty array when none. |

Installed Persona entry:

| Field | Type | Required | Constraints | Description |
|---|---|---|---|---|
| `name` | string | Yes | unique within `items` | Persona identity, as offered by the backend. |
| `priority` | integer | Yes | non-negative; unique within `items`; assigned positionally by `set persona` argument order (`0` first). See the [priority model](../AUTHORING.md#priority-model). | Installed-persona ordering. |
| `description` | string | No | — | Human-readable summary from the definition. |
| `tags` | array of string | No | — | Free-form labels from the definition. |
| `skills` | array of string | No | — | Skill artifacts the persona bundles. |
| `agents` | array of string | No | — | Agent artifacts the persona bundles. |
| `last_modified_at` | string | No | RFC 3339 timestamp | The backend's last-modified time for the definition when it was fetched. |
| `last_synced_at` | string | No | RFC 3339 timestamp | When this definition was last refreshed into `personas.yaml`. |

```yaml
version: 1
items:
  - name: backend-developer
    priority: 0
    description: Backend service development persona.
    tags: [backend, go]
    skills:
      - design-oas3
      - code-review
    agents:
      - software-engineer
    last_modified_at: 2026-06-11T18:00:00Z
    last_synced_at: 2026-06-12T09:30:00Z
  - name: qa-engineer
    priority: 1
    description: Test-authoring and quality persona.
    tags: [qa]
    skills:
      - test-plan
    agents: []
    last_modified_at: 2026-06-10T12:00:00Z
    last_synced_at: 2026-06-12T09:30:00Z
```

## track.yaml

The record of installed artifacts and their provenance. Created and maintained by
[sync artifacts](../0006-sync-artifacts/spec.md).
Schema: [track.schema.json](schemas/track.schema.json).

| Field | Type | Required | Description |
|---|---|---|---|
| `version` | integer | Yes | Schema version. |
| `items` | array of Installed Artifact | Yes | Delivered artifacts; empty array when none. |

Installed Artifact entry:

| Field | Type | Required | Description |
|---|---|---|---|
| `type` | string | Yes | `skill` or `agent`. |
| `name` | string | Yes | Artifact name, as installed. |
| `provider` | string | Yes | Provider the artifact was delivered to (`claude` or `zencoder`). |
| `path` | string | Yes | Where it was installed (the provider's location for this artifact). |
| `registry` | string | Yes | Source registry name — the conflict winner per [ADR-0001](../0006-sync-artifacts/architecture/ADR-0001-conflict-resolution-by-registry-priority.md). |
| `persona` | string | No | Installed persona that brought the artifact into the desired set; the highest-precedence one when several do; absent when synced without personas. |

An entry is identified by (`provider`, `type`, `name`): the same artifact
delivered to two providers yields two entries.

```yaml
version: 1
items:
  - type: skill
    name: code-review
    provider: claude
    path: /home/user/.claude/skills/code-review
    registry: team-deploy
    persona: backend-developer
```

## Live persona view

Sauron does **not** persist a catalog. The set of *available* personas a user can
browse or install is assembled at command time:

- **Installed personas** come from [personas.yaml](#personasyaml) and are always
  available, including offline.
- **Backend personas** are fetched live from [backend.yaml](#backendyaml)'s
  connection when the command runs. When the backend is unreachable, the live
  fetch is skipped and only installed personas are shown.

Consequently [list personas](../0005-list-personas/spec.md) and
[describe persona](../0016-describe-persona/spec.md) degrade gracefully offline,
and [sync personas](../0013-sync-personas/spec.md) refreshes the stored
definitions of installed personas rather than maintaining a mirror.

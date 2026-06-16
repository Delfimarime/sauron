# Configuration Data Contract

The normative schema of every document Sauron persists. Feature
`data/configuration.md` files link here for the schema and never restate it; this
contract never links back to feature requirements (the relationship is
one-directional, no cycle).

## Home and files

Sauron persists its state under a single home directory: `$SAURON_HOME` when set,
the platform default `~/.sauron` otherwise. The home holds three files, each a
multi-document YAML stream:

| File | Documents (`kind`) |
|---|---|
| `registries.yaml` | `Registry` |
| `track.yaml` | `Skill`, `Agent`, `Persona` |
| `settings.yaml` | `Provider`, `Schedule` |

The **catalogue** is not persisted — what a registry offers is fetched live at
command time — so it has no file and no schema here.

## Manifest envelope

Every document is a Kubernetes-style manifest with four top-level keys:

```yaml
apiVersion: sauron.raitonbl.com/v1
kind: <Registry|Skill|Agent|Persona|Provider|Schedule>
metadata:
  name: <identity>
  labels: {}            # optional, free-form, on every kind
spec:
  ...                   # kind-specific (Provider has none beyond metadata)
```

- `apiVersion` is `sauron.raitonbl.com/v1`. Schema evolution is expressed by
  advancing the version in this string; documents of an unknown `apiVersion` are
  rejected.
- `kind` selects the document type and thereby the `spec` schema below.
- `metadata.name` is the document's identity. `metadata.labels` is an optional
  string map available on every kind.

The JSON Schema for each kind lives under [schemas/](schemas/) and is the
machine-checkable form of the rules below:
[Registry](schemas/Registry.schema.json),
[Skill](schemas/Skill.schema.json),
[Agent](schemas/Agent.schema.json),
[Persona](schemas/Persona.schema.json),
[Provider](schemas/Provider.schema.json),
[Schedule](schemas/Schedule.schema.json).

## Write semantics

- **Atomic writes.** A file is written to a temporary file and renamed into place,
  so a reader never observes a half-written file.
- **Lockfile.** Writes are guarded by a lockfile under the home, so a scheduled
  run and a manual command cannot corrupt a file when they overlap.
- **No secrets at rest.** Credentials are stored only as environment references
  (`${env:VAR}`); resolved secret values are never written to any file. The track
  file holds no credentials at all.

## `Registry` (registries.yaml)

```yaml
apiVersion: sauron.raitonbl.com/v1
kind: Registry
metadata:
  name: acme
spec:
  transport: git                 # git | http | filesystem
  uri: git@github.com:acme/artifacts.git
  auth:                          # optional
    username: ${env:ACME_USER}
    password: ${env:ACME_TOKEN}
  tls:                           # optional
    skipVerify: false
    caCert: /path/ca.pem
    clientCert: /path/client.pem
    clientKey: /path/client.key
  sshKey: /path/id_ed25519       # optional
  timeout: 30s
```

- `metadata.name` is unique and path-safe — it is the namespacing segment in
  `sauron-<registry>-<name>` and the value each tracked artifact references. There
  is no rename; a name change is a delete plus a re-add.
- `spec.transport` is the registry's transport. At the CLI it is selected by
  `--kind`; this `--kind` → `spec.transport` mapping is intentional.
- `spec.auth`, `spec.tls`, and `spec.sshKey` apply per transport; secret-bearing
  fields carry environment references only.

## `Skill` / `Agent` (track.yaml)

```yaml
apiVersion: sauron.raitonbl.com/v1
kind: Skill                      # or Agent — identical spec shape
metadata:
  name: go-style
  labels:
    team: backend
spec:
  registry: acme
  version: v1.4.0                # optional
  digest: <content-identity>     # always present
  path: skills/sauron-acme-go-style
  provenance:
    direct: true
    personas: [backend-dev]
  installedAt: 2026-06-15T10:00:00Z
  updatedAt:   2026-06-15T10:00:00Z
```

- The unique key for any artifact is the triple `(kind, registry, name)`; the same
  name may appear across registries and across kinds.
- `spec.digest` is the content identity that `sync`/`upgrade` compare to detect
  upstream change and local drift; it is always present.
- `spec.version` is the optional human-meaningful label (derived for git,
  declared-only for http/filesystem).
- `spec.path` is the exact installed location, so removal and provider migration
  are precise and independent of recomputing the naming scheme.
- `spec.provenance` is the authoritative record of why the artifact is installed:
  `direct` (installed explicitly) and `personas` (personas that bring it in). An
  artifact is removed only when `direct` is false and `personas` is empty.

## `Persona` (track.yaml)

```yaml
apiVersion: sauron.raitonbl.com/v1
kind: Persona
metadata:
  name: backend-dev
spec:
  registry: acme
  version: 9f4d2a1               # optional
  digest: <definition-identity>
  members:
    skills: [go-style, sql-review]
    agents: [code-reviewer]
  installedAt: 2026-06-15T10:00:00Z
  updatedAt:   2026-06-15T10:00:00Z
```

- `spec.members` is the snapshot of the last-resolved membership; `sync`/`upgrade`
  diff the freshly-resolved definition against it. It is distinct from each
  member's `provenance`, which records why that member is installed.
- `spec.digest` is the content identity of the persona definition itself.

## `Provider` / `Schedule` (settings.yaml)

```yaml
apiVersion: sauron.raitonbl.com/v1
kind: Provider
metadata:
  name: claude                   # claude | zencoder
---
apiVersion: sauron.raitonbl.com/v1
kind: Schedule
metadata:
  name: sync                     # sync | upgrade
spec:
  cron: "0 */6 * * *"
```

- There is exactly one `Provider` document; its identity is its `metadata.name`.
- There is at most one `Schedule` document per operation (`sync`, `upgrade`);
  `spec.cron` is the expression registered in the OS crontab.

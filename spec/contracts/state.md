# State Data Contract

The contract for the state Sauron persists under its home. It owns the
**semantics** of that state — the file layout, the manifest envelope, write
guarantees, and the per-kind rules a schema cannot express. Each feature's
`data/state.md` links here and never restates it; this contract never links back
to feature requirements (the relationship is one-directional, no cycle).

## Source of truth

The **JSON Schemas under [schemas/](schemas/) are normative** for the *structure*
of every document — fields, types, required keys, enums, and patterns. They are
machine-checked: the [`storage`](architecture.md#state-storage) capability
validates every document **it reads** against the matching schema, and an invalid
document is a runtime error.

This contract therefore does **not** restate field-level constraints — that would
duplicate the schemas and invite drift. It documents only what the schemas cannot:
which file holds which kind, the shared envelope, the write guarantees, identity
and uniqueness, and the removal rule. For a field's exact shape, read its schema.

## Home and files

Sauron persists its state under a single home directory: `$SAURON_HOME` when set,
the platform default `~/.sauron` otherwise. The home holds two files, each a
multi-document YAML stream:

| File | Documents (`kind`) | Schema |
|---|---|---|
| `track.yaml` | `Skill`, `Agent` | [Skill](schemas/Skill.schema.json) · [Agent](schemas/Agent.schema.json) |
| `settings.yaml` | `Registry`, `Provider` | [Registry](schemas/Registry.schema.json) · [Provider](schemas/Provider.schema.json) |

The **catalogue** is not persisted — what the registry offers is fetched live at
command time — so it has no file and no schema.

The [`Persona`](schemas/Persona.schema.json) and
[`Schedule`](schemas/Schedule.schema.json) schemas are retained under `schemas/`
as reference for their deferred features (see
[ADR-0003](../architecture/ADR-0003-persona-deferred.md) and
[ADR-0004](../architecture/ADR-0004-schedule-deferred.md)); no v1 document uses
them, and neither kind is written to any file.

## Manifest envelope

Every document is a Kubernetes-style manifest: `apiVersion`, `kind`, `metadata`,
and a kind-specific `spec`.

```yaml
apiVersion: sauron.raitonbl.com/v1
kind: <Registry|Skill|Agent|Provider>
metadata:
  name: <identity>
  labels: {}                                # optional, free-form, on every kind
  createdAt: 2026-06-21T07:30:00Z   # writer-stamped, RFC3339 UTC
  lastUpdatedAt: 2026-06-21T07:30:00Z
spec:
  ...                   # kind-specific; see the kind's schema (Provider has none)
```

- `apiVersion` is `sauron.raitonbl.com/v1`. Schema evolution advances the version
  in this string; a document of an unknown `apiVersion` is rejected.
- `kind` selects the document type and thereby its `spec` schema.
- `metadata.name` is the document's identity; `metadata.labels` is an optional
  free-form string map available on every kind.
- `metadata.createdAt` and `metadata.lastUpdatedAt` are audit
  timestamps available on every kind: RFC3339 UTC instants stamped by the use case
  that writes the document, never hand-edited. They are equal when the document is
  first created; a later write advances `lastUpdatedAt` only. Both are
  optional on read, so documents written before they existed still load.

## Write semantics

- **Atomic writes.** A file is written to a temporary file and renamed into place,
  so a reader never observes a half-written file.
- **Lockfile.** Writes are guarded by a lockfile under the home, so a scheduled
  run and a manual command cannot corrupt a file when they overlap.
- **No secrets at rest.** Credential fields hold environment references
  (`${env:VAR}`) only; resolved secret values are never written to any file. The
  track file holds no credentials at all.
- **Owner-only permissions.** The home directory (`$SAURON_HOME`, or `~/.sauron/`)
  is created with mode `0700` and each state file (`track.yaml`, `settings.yaml`)
  is written with mode `0600` — owner read/write only. Although no secret is stored
  at rest, the files reveal which registry and artifacts a developer uses, so they
  are not world- or group-readable.
- **Audit timestamps are writer-stamped.** Whenever Sauron writes a document it
  stamps `metadata.createdAt` (on first create) and
  `metadata.lastUpdatedAt` (on every write) from an injected clock as
  RFC3339 UTC; these fields are never hand-edited and are tolerated absent on read.
- **Validation is on read, not on app-authored write.** Documents are validated
  against their schema when loaded (the home files are hand-editable); documents
  Sauron itself authors are constructed from typed values and written without
  re-validation.

## Per-kind semantics

Structure is in each kind's schema (linked above); the rules below are the
meaning layered on top.

### `Registry` — `settings.yaml`

- There is exactly one `Registry` document — Sauron has a single registry
  (supporting more is deferred, see
  [ADR-0002](../architecture/ADR-0002-single-registry.md)). It carries no
  user-given name; `metadata.name` is unused, and `spec.source` is its identity.
  Setting a registry replaces the one already present.
- `spec.transport` is the registry's transport; at the CLI it is selected by
  `--transport` (the `--transport` → `spec.transport` mapping is intentional).
- `spec.revision` is the optional git revision (branch, tag, or commit) the registry is
  pinned to; it applies to the `git` transport only and, when absent, resolution
  uses the repository's default branch.
- `spec.credentials`, `spec.tls`, and `spec.sshKey` apply per transport; secret-bearing
  fields carry environment references only (see **No secrets at rest**).

### `Skill` / `Agent` — `track.yaml`

`Skill` and `Agent` share one spec shape; they differ only by `kind`.

- The unique key for any artifact is the pair `(kind, name)`; the same name may
  appear across kinds. There is one registry, so the source is implicit and is not
  recorded per artifact.
- `spec.digest` is the content identity `sync`/`upgrade` compare to detect upstream
  change and local drift; it is always present.
- `spec.version` is the optional human-meaningful label — derived for `git`,
  declared-only for `http`/`filesystem`.
- `spec.path` is the exact installed location (`sauron-<name>` under the provider's
  directory for the kind), so removal and provider migration are precise and
  independent of recomputing the naming scheme.

### `Provider` — `settings.yaml`

- There is exactly one `Provider` document; its identity is its `metadata.name`
  (`claude` | `zencoder`).

### `Preferences` — `settings.yaml`

- There is at most one `Preferences` document; `metadata.name` is unused.
- `spec.theme` is the active terminal UI color theme (`sauron`, the default, or
  `light`), set in the TUI by `m` or on either surface by `--theme`; when the
  document or the field is absent, the default theme applies.

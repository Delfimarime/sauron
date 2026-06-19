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
and uniqueness, and the cross-document rules (`provenance`, `members`, removal).
For a field's exact shape, read its schema.

## Home and files

Sauron persists its state under a single home directory: `$SAURON_HOME` when set,
the platform default `~/.sauron` otherwise. The home holds three files, each a
multi-document YAML stream:

| File | Documents (`kind`) | Schema |
|---|---|---|
| `registries.yaml` | `Registry` | [Registry](schemas/Registry.schema.json) |
| `track.yaml` | `Skill`, `Agent`, `Persona` | [Skill](schemas/Skill.schema.json) · [Agent](schemas/Agent.schema.json) · [Persona](schemas/Persona.schema.json) |
| `settings.yaml` | `Provider`, `Schedule` | [Provider](schemas/Provider.schema.json) · [Schedule](schemas/Schedule.schema.json) |

The **catalogue** is not persisted — what a registry offers is fetched live at
command time — so it has no file and no schema.

## Manifest envelope

Every document is a Kubernetes-style manifest: `apiVersion`, `kind`, `metadata`,
and a kind-specific `spec`.

```yaml
apiVersion: sauron.raitonbl.com/v1
kind: <Registry|Skill|Agent|Persona|Provider|Schedule>
metadata:
  name: <identity>
  labels: {}            # optional, free-form, on every kind
spec:
  ...                   # kind-specific; see the kind's schema (Provider has none)
```

- `apiVersion` is `sauron.raitonbl.com/v1`. Schema evolution advances the version
  in this string; a document of an unknown `apiVersion` is rejected.
- `kind` selects the document type and thereby its `spec` schema.
- `metadata.name` is the document's identity; `metadata.labels` is an optional
  free-form string map available on every kind.

## Write semantics

- **Atomic writes.** A file is written to a temporary file and renamed into place,
  so a reader never observes a half-written file.
- **Lockfile.** Writes are guarded by a lockfile under the home, so a scheduled
  run and a manual command cannot corrupt a file when they overlap.
- **No secrets at rest.** Credential fields hold environment references
  (`${env:VAR}`) only; resolved secret values are never written to any file. The
  track file holds no credentials at all.
- **Validation is on read, not on app-authored write.** Documents are validated
  against their schema when loaded (the home files are hand-editable); documents
  Sauron itself authors are constructed from typed values and written without
  re-validation.

## Per-kind semantics

Structure is in each kind's schema (linked above); the rules below are the
meaning layered on top.

### `Registry` — `registries.yaml`

- `metadata.name` is unique and path-safe — it is the namespacing segment in
  `sauron-<registry>-<name>` and the value each tracked artifact references. There
  is no rename; a name change is a delete plus a re-add.
- `spec.transport` is the registry's transport; at the CLI it is selected by
  `--kind` (the `--kind` → `spec.transport` mapping is intentional).
- `spec.ref` is the optional git ref (branch, tag, or commit) the registry is
  pinned to; it applies to the `git` transport only and, when absent, resolution
  uses the repository's default branch.
- `spec.auth`, `spec.tls`, and `spec.sshKey` apply per transport; secret-bearing
  fields carry environment references only (see **No secrets at rest**).

### `Skill` / `Agent` — `track.yaml`

`Skill` and `Agent` share one spec shape; they differ only by `kind`.

- The unique key for any artifact is the triple `(kind, registry, name)`; the same
  name may appear across registries and across kinds.
- `spec.digest` is the content identity `sync`/`upgrade` compare to detect upstream
  change and local drift; it is always present.
- `spec.version` is the optional human-meaningful label — derived for `git`,
  declared-only for `http`/`filesystem`.
- `spec.path` is the exact installed location, so removal and provider migration
  are precise and independent of recomputing the naming scheme.
- `spec.provenance` is the authoritative record of why the artifact is installed:
  `direct` (installed explicitly) and `personas` (the personas that bring it in).
  An artifact is removed only when `direct` is false and `personas` is empty.

### `Persona` — `track.yaml`

- `spec.members` is the snapshot of the last-resolved membership; `sync`/`upgrade`
  diff the freshly-resolved definition against it. It is distinct from each
  member's `provenance`, which records why that member is installed.
- `spec.digest` is the content identity of the persona definition itself.

### `Provider` / `Schedule` — `settings.yaml`

- There is exactly one `Provider` document; its identity is its `metadata.name`
  (`claude` | `zencoder`).
- There is at most one `Schedule` document per operation (`sync`, `upgrade`);
  `spec.cron` is the expression registered in the OS crontab.

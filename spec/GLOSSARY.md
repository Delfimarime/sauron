# Glossary

The canonical domain vocabulary for sauron. One term per concept; specs,
contracts, and code use these words and no synonyms. This is the single source of
the project's vocabulary — [AUTHORING.md](AUTHORING.md) and the contracts link
here rather than restating a definition.

**Audience:** everyone — developers, architects, security analysts, and
engineers reading any spec or contract.

| Term | Meaning |
|---|---|
| artifact | A unit Sauron distributes: a skill or an agent. A third kind, persona, is deferred past v1 — see [ADR-0003](architecture/ADR-0003-persona-deferred.md) |
| skill | An artifact hosted under a registry's `.skills/` directory |
| agent | An artifact hosted under a registry's `.agents/` directory |
| persona | *(deferred — not implemented in v1; see [ADR-0003](architecture/ADR-0003-persona-deferred.md))* A first-class artifact that references a set of skills and agents within the registry; installed, listed, and described like any artifact |
| membership | *(deferred — see persona)* The set of skills and agents a persona references; resolved at install and re-resolved by `sync`/`upgrade` |
| registry | The single registered source of artifacts; Sauron has exactly one. Supporting more is deferred — see [ADR-0002](architecture/ADR-0002-single-registry.md) |
| transport | A registry's type — `git`, `http`, or `filesystem` — selecting how the source is reached, validated, and fetched from; persisted as `spec.transport` and selected at the CLI by `--kind` |
| kind | In a manifest, the document type (`Registry`, `Skill`, `Agent`, `Provider`; `Persona` and `Schedule` are deferred). At the CLI, the `--kind` flag selects the registry's `transport` |
| ref | A git revision — a branch, tag, or commit — a `git`-transport registry is pinned to; persisted as `spec.ref` and selected at the CLI by `--ref`; when absent, the repository's default branch is used |
| catalogue | The live, paginated view of what the registry offers, fetched fresh at command time; it is never persisted and has no offline form |
| provider | The destination environment where artifacts are installed (`claude`, `zencoder`); a single global setting recorded as the `Provider` document in `settings` |
| namespacing | The installed-target naming `sauron-<name>`; the `sauron-` prefix marks Sauron ownership, so Sauron only ever touches artifacts it installed |
| install | Fetching named artifacts from the registry and placing them under the provider |
| uninstall | Removing named installed artifacts from the provider and the track file |
| sync | The full reconcile of the installed set against the source: refresh, drift repair, and removal of what vanished upstream |
| upgrade | The non-destructive refresh of the installed set: refresh changed artifacts; never removes |
| plan | The printed list of pending changes — `+` additions, `~` updates, `-` removals |
| digest | The content identity recorded per artifact, used to detect change and local drift; always present |
| version | An optional artifact label or revision identifier; for `git` it is the commit that last touched the artifact (that commit's SHA), and is declared by the source otherwise |
| track file | `track.yaml`, the multi-document stream of `Skill`/`Agent` manifests recording installed artifacts |
| state | The set of files Sauron persists under `~/.sauron/` — `track.yaml` and `settings.yaml` — whose schema is owned by the [state data contract](contracts/state.md). Distinct from the `Configuration` DI struct, which is app configuration (resolved home), not persisted state |
| settings | `settings.yaml`, holding the `Registry` and `Provider` documents |
| manifest | A persisted document carrying `apiVersion` (`sauron.raitonbl.com/v1`) and `kind`, with `metadata` and `spec`, in the spirit of a Kubernetes object |

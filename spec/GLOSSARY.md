# Glossary

The canonical domain vocabulary for sauron. One term per concept; specs,
contracts, and code use these words and no synonyms. This is the single source of
the project's vocabulary — [AUTHORING.md](AUTHORING.md) and the contracts link
here rather than restating a definition.

**Audience:** everyone — developers, architects, security analysts, and
engineers reading any spec or contract.

| Term | Meaning |
|---|---|
| artifact | A unit Sauron distributes: a skill, an agent, or a persona |
| skill | An artifact hosted under a registry's `.skills/` directory |
| agent | An artifact hosted under a registry's `.agents/` directory |
| persona | A first-class artifact that references a set of skills and agents within the same registry; installed, listed, and described like any artifact |
| membership | The set of skills and agents a persona references; resolved at install and re-resolved by `sync`/`upgrade` |
| registry | A registered source of artifacts |
| transport | A registry's type — `git`, `http`, or `filesystem` — selecting how the source is reached, validated, and fetched from; persisted as `spec.transport` and selected at the CLI by `--kind` |
| kind | In a manifest, the document type (`Registry`, `Skill`, `Agent`, `Persona`, `Provider`, `Schedule`). At the CLI, the `--kind` flag selects a registry's `transport` |
| ref | A git revision — a branch, tag, or commit — a `git`-transport registry is pinned to; persisted as `spec.ref` and selected at the CLI by `--ref`; when absent, the repository's default branch is used |
| catalogue | The live, paginated view of what a registry offers, fetched fresh at command time; it is never persisted and has no offline form |
| provider | The destination environment where artifacts are installed (`claude`, `zencoder`); a single global setting recorded as the `Provider` document in `settings` |
| namespacing | The installed-target naming `sauron-<registry>-<name>`, which lets two registries offer the same artifact name without conflict; the `sauron-` prefix marks Sauron ownership |
| install | Fetching named artifacts from a registry and placing them under the provider; installing a persona installs its members |
| uninstall | Removing named installed artifacts; uninstalling a persona removes the members it brought in |
| sync | The full reconcile of the installed set against its sources: refresh, drift repair, removal of what vanished upstream, and persona membership re-resolution (additions and removals) |
| upgrade | The non-destructive refresh of the installed set: refresh changed artifacts and add newly-added persona members; never removes |
| plan | The printed list of pending changes — `+` additions, `~` updates, `-` removals |
| digest | The content identity recorded per artifact, used to detect change and local drift; always present |
| version | An optional artifact label or revision identifier; for `git` it is the commit that last touched the artifact (that commit's SHA), and is declared by the source otherwise |
| provenance | The origin recorded for each installed artifact in the `track file`: whether it was installed directly and which personas brought it in |
| track file | `track.yaml`, the multi-document stream of `Skill`/`Agent`/`Persona` manifests recording installed artifacts and provenance |
| state | The set of files Sauron persists under `~/.sauron/` — `registries.yaml`, `track.yaml`, and `settings.yaml` — whose schema is owned by the [state data contract](contracts/state.md). Distinct from the `Configuration` DI struct, which is app configuration (resolved home), not persisted state |
| settings | `settings.yaml`, holding the `Provider` document and the `Schedule` documents |
| manifest | A persisted document carrying `apiVersion` (`sauron.raitonbl.com/v1`) and `kind`, with `metadata` and `spec`, in the spirit of a Kubernetes object |

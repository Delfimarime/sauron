# Sauron

A provider-agnostic command-line manager for AI coding artifacts — skills and
agents.

**Audience:** this documentation is written for developers, architects, security
analysts, and engineers. It is the canonical domain model; the feature specs,
contracts, and the [glossary](GLOSSARY.md) build on the vocabulary defined here.
Unless a spec declares a narrower `**Audience:**`, this is the audience it
assumes.

## The Problem

Skills and agents are important artifacts when coding with agentic AI, and they
matter even more for a team expected to follow the same principles. How do you
distribute these artifacts and keep them up to date across the team?

Each AI provider ships its own marketplace, bound to that provider. Sauron ignores
those boundaries: it delivers skills and agents — and keeps them current — in any
environment, regardless of provider.

## How it works

Sauron is imperative, in the spirit of a package manager (`apt`, `brew`). The
developer **installs** named artifacts from a registry and **uninstalls** them;
**sync** and **upgrade** keep what is installed current with its source. Nothing
runs in the background unless explicitly scheduled.

```
   ┌──────────────────────┐                          ┌──────────────────────┐
   │      Registries      │                          │         User         │
   │   artifact sources   │                          │  a developer using   │
   │   git · http · fs    │                          │   an AI assistant    │
   └──────────┬───────────┘                          └──────────┬───────────┘
              ▲                                                  │
              │ fetch artifacts                                  │ runs `sauron`
              │ (install · sync · upgrade)                       │ commands
┌─ User env ──┼──────────────────────────────────────────────────┼───────────┐
│             │                                                  │            │
│  ┌────────────┐    ┌──────────────────┐                        │            │
│  │ OS crontab ├──▶ │    SAURON CLI    │◀───────────────────────┘            │
│  │ (optional) │    │                  │                                     │
│  └────────────┘    └─────┬────────────┘                                     │
│                          │ installs / removes artifacts                     │
│                          ▼                                                  │
│            ┌─────────────┬──────────────┐                                   │
│            │          Provider          │                                   │
│            │     claude | zencoder      │                                   │
│            │   (artifact directories)   │                                   │
│            └────────────────────────────┘                                   │
└─────────────────────────────────────────────────────────────────────────────┘
```

Everything Sauron touches at delivery time lives in the user's environment: the
CLI itself, the optional crontab entry that schedules it, and the provider's
artifact directories. Registries are external sources — a `filesystem` registry
may happen to be on the same machine, but Sauron treats it as a source like any
other.

## Concepts

- **Artifact** — a unit Sauron distributes. Three kinds: **skill**, **agent**, and
  **persona**.
- **Skill / Agent** — content hosted in a registry under its `.skills/` or
  `.agents/` directory.
- **Persona** — a named grouping that references a set of skills and agents
  **within the same registry**. A persona is first-class: it is installed, listed,
  and described like any artifact, and its realized content is its resolved
  **membership**.
- **Registry** — a registered source of artifacts. Its **transport** — `git`,
  `http`, or `filesystem` — determines how it is reached, validated, and fetched
  from.
- **Provider** — the destination environment where artifacts are installed
  (`claude` or `zencoder`). There is one global provider; changing it migrates the
  installed artifacts to the new provider's directories.
- **Catalogue** — the live, paginated view of what a registry offers. It is always
  fetched fresh; it is never persisted, and there is no offline catalogue.
- **Track** — the recorded set of installed artifacts and their provenance: the
  source of truth for `uninstall`, `sync`, and `upgrade`.

### Namespacing

Every installed artifact lands at a target named `sauron-<registry>-<name>` in the
provider's directory for its kind. The registry segment namespaces artifacts, so
two registries may offer the same name without conflict. The `sauron-` prefix
marks ownership: Sauron only ever touches artifacts it installed.

## Domain model

```
  ┌──────────────────────┐         ┌──────────────────────┐
  │       REGISTRY       │         │       PROVIDER       │
  │  name · uri · kind   │         │  claude | zencoder   │
  │   (transport) · …    │         │  one global setting  │
  └──────────┬───────────┘         └──────────┬───────────┘
             │ hosts                          │
             ▼                                │
   ┌──────────────────────────────────┐       │
   │             ARTIFACT             │       │
   │  skill (.skills/) ·              │       │
   │  agent (.agents/) ·              │       │
   │  persona → members within        │       │
   │            the same registry     │       │
   └──────────────────┬───────────────┘       │
                      │ install / uninstall   │  installs into
                      │ sync / upgrade        │  sauron-<registry>-<name>
                      ▼                       │
            ┌───────────────────────┐         │
            │      track.yaml       │◀────────┘
            │ installed + provenance│
            └───────────────────────┘
```

## State

Sauron keeps its state under `~/.sauron/` (or `$SAURON_HOME`) in three YAML files,
each a multi-document stream of Kubernetes-style manifests
(`apiVersion: sauron.raitonbl.com/v1`, with `kind`):

```
~/.sauron/
├── registries.yaml   Registry documents — the registered artifact sources
├── track.yaml        Skill / Agent / Persona documents — what is installed and where it came from
└── settings.yaml     Provider and Schedule documents — the active provider and the sync/upgrade schedules
```

The schema of every document is owned by the
[state data contract](contracts/state.md). The **catalogue** is
not persisted: what a registry *offers* is computed live from a fetch against the
registry, so there is no offline catalogue.

The track file is what makes maintenance safe: `uninstall`, `sync`, and `upgrade`
touch only artifacts Sauron recorded, and provenance distinguishes an artifact
installed directly from one brought in by a persona, so cascades remove exactly
what they should.

## The flow

1. **Register sources** — `add registry` validates and persists each source.
2. **Browse** — `list catalogue` shows what a registry offers, live and paginated.
3. **Install** — `install (skill|agent|persona) <registry> <name>...` installs
   named artifacts; installing a persona installs its members.
4. **Maintain** — `upgrade` refreshes installed artifacts to the latest
   non-destructively; `sync` fully reconciles them against their sources
   (refresh, drift repair, removal of what vanished, persona membership
   re-resolution). Review either with `--dry-run`.
5. **Schedule** — `schedule (sync|upgrade) <expression>` registers an OS-crontab
   entry; `unschedule (sync|upgrade)` removes it.

## Reconciling artifacts

`sync` and `upgrade` operate only on the installed set; neither installs anything
new on its own.

```
   installed set  (track.yaml)
        │
        │ for each artifact: compare digest against its source
        ▼
   ┌─────────────────────────┐
   │          PLAN           │   --dry-run ──▶ print plan, change nothing, exit 0
   │  + add  ~ update  - remove
   └────────────┬────────────┘
                │ apply
                ▼
   install / update / remove on provider ──▶ record in track.yaml
```

- **`upgrade`** is non-destructive: it refreshes changed artifacts and adds
  newly-added persona members, but never removes anything.
- **`sync`** is a full reconcile: everything `upgrade` does, plus repairing local
  drift, removing artifacts that vanished upstream, and removing persona members
  dropped upstream.

## Specifications

Every feature is specified before it is built. Status is owned by each feature's
`spec.md` `**Status:**` field; this table is the aggregated view.

| # | Feature | Status |
|---|---|---|
| [0001](0001-add-registry/spec.md) | Add registry | **Built** |
| [0002](0002-list-registries/spec.md) | List registries | **Built** |
| [0003](0003-describe-registry/spec.md) | Describe registry | **Built** |
| [0004](0004-delete-registry/spec.md) | Delete registry | **Partial** — registry removal ships; artifact cascade deferred to [0007](0007-uninstall-artifacts/spec.md) |
| [0005](0005-list-catalogue/spec.md) | List catalogue | Specified |
| [0006](0006-install-artifacts/spec.md) | Install artifacts | Specified |
| [0007](0007-uninstall-artifacts/spec.md) | Uninstall artifacts | Specified |
| [0008](0008-sync/spec.md) | Sync | Specified |
| [0009](0009-upgrade/spec.md) | Upgrade | Specified |
| [0010](0010-list-artifacts/spec.md) | List artifacts | Specified |
| [0011](0011-describe-artifact/spec.md) | Describe artifact | Specified |
| [0012](0012-set-provider/spec.md) | Set provider | Specified |
| [0013](0013-describe-provider/spec.md) | Describe provider | Specified |
| [0014](0014-schedule/spec.md) | Schedule | Specified |

Authoring conventions and the requirement taxonomy live in
[AUTHORING.md](AUTHORING.md); the domain vocabulary in [GLOSSARY.md](GLOSSARY.md);
the binding cross-cutting contracts under [contracts/](contracts/).

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
runs in the background; to reconcile periodically, a developer wires `sync` or
`upgrade` into their own OS scheduler.

```
   ┌──────────────────────┐                          ┌──────────────────────┐
   │       Registry       │                          │         User         │
   │   artifact source    │                          │  a developer using   │
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
CLI itself, an optional OS-scheduler entry the developer wires themselves, and the
provider's artifact directories. The registry is an external source — a
`filesystem` registry may happen to be on the same machine, but Sauron treats it
as a source like any other.

## Concepts

- **Artifact** — a unit Sauron distributes. Two kinds in v1: **skill** and
  **agent**. A third kind, **persona**, is a defined concept deferred past v1 — see
  [ADR-0003](architecture/ADR-0003-persona-deferred.md).
- **Skill / Agent** — content hosted in a registry under its `.skills/` or
  `.agents/` directory.
- **Persona** *(deferred — not implemented in v1)* — a named grouping that
  references a set of skills and agents within the registry; first-class
  (installed, listed, and described like any artifact), its realized content being
  its resolved **membership**. The full design is recorded in
  [ADR-0003](architecture/ADR-0003-persona-deferred.md).
- **Registry** — the single registered source of artifacts. Its **transport** —
  `git`, `http`, or `filesystem` — determines how it is reached, validated, and
  fetched from. Sauron has exactly one registry; supporting more is deferred — see
  [ADR-0002](architecture/ADR-0002-single-registry.md).
- **Provider** — the destination environment where artifacts are installed
  (`claude` or `zencoder`). There is one global provider; changing it migrates the
  installed artifacts to the new provider's directories.
- **Catalogue** — the live, paginated view of what the registry offers. It is
  always fetched fresh; it is never persisted, and there is no offline catalogue.
- **Track** — the recorded set of installed artifacts: the source of truth for
  `uninstall`, `sync`, and `upgrade`.

### Namespacing

Every installed artifact lands at a target named `sauron-<name>` in the provider's
directory for its kind. The `sauron-` prefix marks ownership: Sauron only ever
touches artifacts it installed.

## Domain model

```
  ┌──────────────────────┐         ┌──────────────────────┐
  │       REGISTRY       │         │       PROVIDER       │
  │  uri · transport     │         │  claude | zencoder   │
  │  one global setting  │         │  one global setting  │
  └──────────┬───────────┘         └──────────┬───────────┘
             │ hosts                          │
             ▼                                │
   ┌──────────────────────────────────┐       │
   │             ARTIFACT             │       │
   │  skill (.skills/) ·              │       │
   │  agent (.agents/)                │       │
   │  (persona — deferred, ADR-0003)  │       │
   └──────────────────┬───────────────┘       │
                      │ install / uninstall   │  installs into
                      │ sync / upgrade        │  sauron-<name>
                      ▼                       │
            ┌───────────────────────┐         │
            │      track.yaml       │◀────────┘
            │   installed artifacts │
            └───────────────────────┘
```

## State

Sauron keeps its state under `~/.sauron/` (or `$SAURON_HOME`) in two YAML files,
each a multi-document stream of Kubernetes-style manifests
(`apiVersion: sauron.raitonbl.com/v1`, with `kind`):

```
~/.sauron/
├── track.yaml      Skill / Agent documents — what is installed
└── settings.yaml   Registry and Provider documents — the configured source and the active provider
```

The schema of every document is owned by the
[state data contract](contracts/state.md). The **catalogue** is not persisted:
what the registry *offers* is computed live from a fetch against it, so there is no
offline catalogue.

The track file is what makes maintenance safe: `uninstall`, `sync`, and `upgrade`
touch only artifacts Sauron recorded, so they never disturb a developer's own
skills and agents.

## The flow

1. **Set the source** — `set registry` validates and persists the single source.
2. **Browse** — `list catalogue` shows what the registry offers, live and
   paginated.
3. **Install** — `install (skill|agent) <name>...` installs named artifacts.
4. **Maintain** — `upgrade` refreshes installed artifacts to the latest
   non-destructively; `sync` fully reconciles them against the source (refresh,
   drift repair, removal of what vanished). Review either with `--dry-run`.

To reconcile on a schedule, wire `sauron sync` or `sauron upgrade` into the OS
scheduler. A built-in `schedule` command is deferred — see
[ADR-0004](architecture/ADR-0004-schedule-deferred.md).

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

- **`upgrade`** is non-destructive: it refreshes changed artifacts but never
  removes anything.
- **`sync`** is a full reconcile: everything `upgrade` does, plus repairing local
  drift and removing artifacts that vanished upstream.

## Specifications

Every feature is specified before it is built. Status is owned by each feature's
`spec.md` `**Status:**` field; this table is the aggregated view. Decisions that
shape the whole project — single registry, deferred persona, deferred schedule —
are recorded as project-level ADRs under [architecture/](architecture/).

| # | Feature | Status |
|---|---|---|
| [0001](0001-set-registry/spec.md) | Set registry | **Built** |
| [0002](0002-describe-registry/spec.md) | Describe registry | **Built** |
| [0003](0003-unset-registry/spec.md) | Unset registry | **Built** |
| [0004](0004-list-catalogue/spec.md) | List catalogue | Specified |
| [0005](0005-set-provider/spec.md) | Set provider | Specified |
| [0006](0006-describe-provider/spec.md) | Describe provider | Specified |
| [0007](0007-install-artifacts/spec.md) | Install artifacts | Specified |
| [0008](0008-list-artifacts/spec.md) | List artifacts | Specified |
| [0009](0009-describe-artifact/spec.md) | Describe artifact | Specified |
| [0010](0010-uninstall-artifacts/spec.md) | Uninstall artifacts | Specified |
| [0011](0011-sync/spec.md) | Sync | Specified |
| [0012](0012-upgrade/spec.md) | Upgrade | Specified |

Authoring conventions and the requirement taxonomy live in
[AUTHORING.md](AUTHORING.md); the domain vocabulary in [GLOSSARY.md](GLOSSARY.md);
the binding cross-cutting contracts under [contracts/](contracts/).

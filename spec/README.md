# Sauron

A simple orchestrator for delivering skills and agents.

## The Problem

Skills and agents are important artifacts when coding with agentic AI, and they
matter even more for a team expected to follow the same principles. How do you
distribute these artifacts and keep them up to date across the team?

Claude has a marketplace, and other providers have their own, but each is
restricted to its specific provider. Sauron ignores those boundaries: it
delivers skills and agents — and keeps them current — in any environment,
regardless of provider.

## System Context

```
   ┌──────────────────────┐  ┌──────────────────────┐  ┌──────────────────────┐
   │      Registries      │  │       Backend        │  │         User         │
   │   artifact sources   │  │  persona-definition  │  │  a developer using   │
   │   git · http · fs    │  │     source (one)     │  │   an AI assistant    │
   └──────────┬───────────┘  └──────────┬───────────┘  └──────────┬───────────┘
              ▲                         ▲                         │
              │ fetch artifacts         │ fetch persona           │ runs `sauron`
              │ (sync artifacts)        │ defs (sync personas)    │ commands
┌─ User env ──┼─────────────────────────┼─────────────────────────┼────────────┐
│             │                         │                         │            │
│             └────────────┬────────────┘                         │            │
│  ┌────────────┐          │                                      │            │
│  │ OS crontab ├──▶ ┌─────┴────────────┐                         │            │
│  │ (optional) │    │    SAURON CLI    │◀────────────────────────┘            │
│  └────────────┘    │                  │                                      │
│                    └─────┬────────────┘                                      │
│                          │ installs / removes artifacts                      │
│                          ▼                                                   │
│            ┌─────────────┬──────────────┐                                    │
│            │          Provider          │                                    │
│            │     claude | zencoder      │                                    │
│            │   (artifact directories)   │                                    │
│            └────────────────────────────┘                                    │
└──────────────────────────────────────────────────────────────────────────────┘
```

Everything Sauron touches at delivery time lives in the user's environment:
the CLI itself, the optional crontab entry that schedules it, and the
provider's artifact directories. Registries are external sources — a
`filesystem` registry may happen to be on the same machine, but Sauron
treats it as a source like any other. The backend that supplies persona
definitions is external in the same way.

## Concepts

- **Artifact** — a skill or an agent distributed by Sauron. Skills live under
  a registry's `.skills/` directory, agents under `.agents/`.
- **Registry** — a registered source of artifacts. Its **kind** — `http`,
  `filesystem`, or `git` — determines how the source is validated and how
  artifacts are fetched from it. A registry must host at least one skill or
  agent.
- **Persona** — a named set of artifacts shared by a group of people, e.g.
  *Backend Dev*. Personas can carry tags and are optional: when none are
  defined, Sauron delivers everything the registries provide.
- **Backend** — the singleton external source that owns **persona** definitions,
  the persona analog of a registry. Sauron fetches the definitions from it during
  `sync personas`, and it may be `http`, `filesystem`, or `git`.
- **Provider** — the destination environment where artifacts are installed
  (`claude` or `zencoder`). There is one global provider; changing it migrates
  the installed artifacts to the new provider's directories.
- **Priority** — integer precedence, lower value wins. When two registries
  offer the same artifact name, registry priority resolves the conflict;
  persona priority orders personas the same way. A **pin** binds a specific
  artifact to a chosen registry, overriding priority for that artifact.
- **Sync & Plan** — sync computes the desired artifact set from registries
  and personas, prints a **plan** (`+` additions, `-` removals), and
  reconciles the provider to it. A dry run prints the plan and stops, changing
  nothing.

## Domain model

```
  ┌──────────────────────┐
  │       BACKEND        │
  │      uri · kind      │
  │     (singleton)      │
  └──────────┬───────────┘
             │ defines
             ▼
  ┌──────────────────────┐    ┌──────────────────────┐    ┌──────────────────────┐
  │       PERSONA        │    │       REGISTRY       │    │       PROVIDER       │
  │   name · priority    │    │      name · uri      │    │  claude | zencoder   │
  │         tags         │    │   kind · priority    │    │  one global setting  │
  └──────────┬───────────┘    └──────────┬───────────┘    └──────────┬───────────┘
             │ groups a set of           │ hosts                     │
             ▼                           ▼                           │
        ┌────┬───────────────────────────┬────┐                      │
        │              ARTIFACT               │                      │
        │ skill (.skills/) · agent (.agents/) │                      │
        └──────────────────┬──────────────────┘                      │
                           │ name clash? lower                       │
                           ▼ registry priority wins                  │
                 ┌─────────┬─────────┐                               │
                 │       SYNC        │                               │
                 │   prints a PLAN   │         installs into         │
                 │ + add / - remove  ├───────────────────────────────┘
                 └─────────┬─────────┘
                           │ records installed artifacts + provenance
                           ▼
               ┌───────────────────────┐
               │      track.yaml       │
               └───────────────────────┘
```

## State

Sauron keeps its state in files under `~/.sauron/`, split by concern:

```
~/.sauron/
├── registries.yaml   the registered artifact sources
├── backend.yaml      the singleton backend connection (persona definitions)
├── personas.yaml     the installed personas, with their definitions
├── track.yaml        what is installed and where it came from (provenance)
└── settings.yaml     global settings: the active provider and the sync schedule
```

The schema of every file is owned by the
[configuration data contract](contracts/configuration.md). There is no persisted
catalog: the *available* personas are computed live from the installed personas
plus a live fetch from the backend, so Sauron still lists and describes installed
personas when the backend is unreachable.

The track file is what makes maintenance safe: pruning removes artifacts
orphaned by unregistered registries, deleting artifacts removes everything Sauron
installed (and nothing else), and deleting a registry or persona never
touches already-installed artifacts.

## How it works

Sauron is a command-line application; nothing runs in the background unless
scheduled.

1. **Register sources** — add registries and configure a backend for personas;
   each is validated before it is persisted to its configuration file.
2. **Sync** — reconcile the provider with the desired set; review the plan with
   a dry run first.
3. **Maintain** — re-run sync for updates; prune or delete artifacts to clean
   up; adjust priorities to resolve conflicts.
4. **Schedule** — `schedule sync artifacts` and `schedule sync personas` register
   OS crontab entries that run the syncs automatically; `unschedule sync` removes
   them.

## Further reading

- [Spec authoring rules](AUTHORING.md) — spec types, numbering, required
  sections, EARS templates, glossary, cross-link form, and the CLI conventions
  (command grammar, shared flags, exit status, output discipline).
- [Command line interface reference](contracts/cli.md) — the compiled list of
  every command, with its synopsis, intent, key flags, and a link to the
  feature contract that owns it.
- [Constitution](../CONSTITUTION.md) — project and implementation principles.

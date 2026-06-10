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
       ┌───────────────────────────┐       ┌───────────────────────────┐
       │       Repositories        │       │           User            │
       │ external artifact sources │       │  a developer using an AI  │
       │  git · http · filesystem  │       │      coding assistant     │
       └─────────────┬─────────────┘       └─────────────┬─────────────┘
                     ▲                                   │
                     │ fetches artifacts                 │ runs `sauron` commands
                     │ during sync                       │
┌─ User environment ─┼───────────────────────────────────┼───────────────┐
│                    │                                   │               │
│                    │       ┌──────────────────┐        │               │
│                    └───────┤                  │◀───────┘               │
│  ┌────────────────┐invokes │    SAURON CLI    │                        │
│  │   OS crontab   ├───────▶│                  │                        │
│  │   (optional    │ `sync` └─────────┬────────┘                        │
│  │   scheduler)   │                  │ installs / removes              │
│  └────────────────┘                  │ artifacts                       │
│                                      ▼                                 │
│                       ┌─────────────────────────────┐                  │
│                       │       Target provider       │                  │
│                       │      claude | zencoder      │                  │
│                       │    (artifact directories)   │                  │
│                       └─────────────────────────────┘                  │
└────────────────────────────────────────────────────────────────────────┘
```

Everything Sauron touches at delivery time lives in the user's environment:
the CLI itself, the optional crontab entry that schedules it, and the target
provider's artifact directories. Repositories are external sources — a
`filesystem` repository may happen to be on the same machine, but Sauron
treats it as a source like any other.

## Concepts

- **Artifact** — a skill or an agent distributed by Sauron. Skills live under
  a repository's `.skills/` directory, agents under `.agents/`.
- **Repository** — a registered source of artifacts. Its **kind** — `http`,
  `filesystem`, or `git` — determines how the source is validated and how
  artifacts are fetched from it. A repository must host at least one skill or
  agent.
- **Persona** — a named set of artifacts shared by a group of people, e.g.
  *Backend Dev*. Personas can carry tags and are optional: when none are
  defined, Sauron delivers everything the repositories provide.
- **Target** — the provider destination where artifacts are persisted
  (`claude` or `zencoder`). There is one global target; changing it migrates
  the installed artifacts to the new provider's directories.
- **Priority** — integer precedence, lower value wins. When two repositories
  offer the same artifact name, repository priority resolves the conflict;
  persona priority orders personas the same way.
- **Sync & Plan** — sync computes the desired artifact set from repositories
  and personas, prints a **plan** (`+` additions, `-` removals), and
  reconciles the target to it. A dry run prints the plan and stops, changing
  nothing.

## Domain model

```
  ┌────────────────────┐      ┌────────────────────┐      ┌────────────────────┐
  │     REPOSITORY     │      │      PERSONA       │      │       TARGET       │
  │  name · location   │      │  name · priority   │      │ claude | zencoder  │
  │  kind · priority   │      │        tags        │      │ one global setting │
  └─────────┬──────────┘      └─────────┬──────────┘      └─────────┬──────────┘
            │ hosts                     │ groups a set of           ▲
            ▼                           ▼                           │
  ┌────────────────────────────────────────────────┐                │
  │                    ARTIFACT                    │                │
  │      skill (.skills/) · agent (.agents/)       │                │
  └────────────────────────┬───────────────────────┘                │
                           │ same artifact name from two sources?   │
                           │ → the lower repository priority wins   │
                           ▼                                        │
                ┌─────────────────────┐                             │
                │        SYNC         │                             │
                │    prints a PLAN    │              installs into  │
                │  + add / - remove   ├─────────────────────────────┘
                └──────────┬──────────┘
                           │ records installed artifacts + provenance
                           ▼
                ┌─────────────────────┐
                │     track.yaml      │
                └─────────────────────┘
```

## State

Sauron keeps its state in two files:

```
~/.sauron/
├── settings.yaml   the configuration: repositories, personas, target, priorities
└── track.yaml      what is installed and where it came from (provenance)
```

The track file is what makes maintenance safe: pruning removes artifacts
orphaned by unregistered repositories, clearing removes everything Sauron
installed (and nothing else), and deleting a repository or persona never
touches already-installed artifacts.

## How it works

Sauron is a command-line application; nothing runs in the background unless
scheduled.

1. **Register sources** — add repositories and import personas; each is
   validated before it is persisted to the settings.
2. **Sync** — reconcile the target with the desired set; review the plan with
   a dry run first.
3. **Maintain** — re-run sync for updates; prune or clear to clean up; adjust
   priorities to resolve conflicts.
4. **Schedule** — register an OS crontab entry that runs the sync
   automatically.

## Further reading

- [Spec authoring rules](AUTHORING.md) — spec types, numbering, required
  sections, EARS templates, glossary, cross-link form, and the CLI conventions
  (command grammar, shared flags, exit status, output discipline).
- [Command line interface reference](contracts/cli.md) — the compiled list of
  every command, with its synopsis, intent, key flags, and a link to the
  feature contract that owns it.
- [Constitution](../CONSTITUTION.md) — project and implementation principles.

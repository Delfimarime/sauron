# sauron

Distribute skills and agents to any AI coding assistant — across provider
boundaries — and keep them in sync.

sauron is a spec-driven CLI that delivers skills and agents from remote
registries (git, HTTP, or filesystem) to the assistant you use (Claude,
Zencoder, …). It groups them into shareable personas and reconciles your
provider to a desired set with a reviewable plan, manually or on a schedule.
(Currently in its specification phase — see [`spec/`](spec/).)

## How it works

sauron is a command-line app — nothing runs in the background unless you
schedule it.

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

Registries and the backend are external sources; the CLI, the optional crontab
entry, and the provider's artifact directories all live in your environment.

1. **Register sources** — add registries (where skills and agents come from)
   and, optionally, a backend that defines personas. Each source is validated
   before it is saved.
2. **Sync** — sauron computes the *desired set* of artifacts from your
   registries and personas, prints a **plan** (`+` to add, `-` to remove), and
   reconciles your provider to match. Run with `--dry-run` to preview the plan
   and change nothing.
3. **Maintain** — re-run sync for updates; prune or delete artifacts to clean
   up; set priorities or pin an artifact to a registry to resolve name clashes.
4. **Schedule** — register OS crontab entries so the syncs run automatically.

State lives in files under `~/.sauron/`, and a track file records what sauron
installed and where it came from — so maintenance only ever touches artifacts
sauron itself manages. The full model is in the
[domain documentation](spec/README.md).

## Documentation

- [Domain model & concepts](spec/README.md) — registries, personas, providers, sync/plan, and state.
- [Command-line reference](spec/contracts/cli.md) — every command, its synopsis, flags, and owning feature.
- [Constitution](CONSTITUTION.md) — project principles and the verification gate.
- [Contributing](CONTRIBUTING.md) — Trunk-based flow, Conventional Commits, proposals, and issue templates.
- [Agent & contributor guide](AGENTS.md) — how humans and AI agents work in this repo; routes to all the rules.
- [Architecture contract](spec/contracts/architecture.md) — layout, fx wiring, Use Case/Action, storage, CI/CD, gates, dependencies.
- [Spec authoring rules](spec/AUTHORING.md) — spec types, EARS, numbering, glossary, and CLI conventions.
- License — [Apache-2.0](LICENSE).

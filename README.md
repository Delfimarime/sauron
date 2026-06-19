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

Registries are external artifact sources; the CLI, the optional crontab entry,
and the provider's artifact directories all live in your environment. You
register sources, sauron validates them, and then it installs the skills and
agents you ask for and keeps them current — touching only artifacts it installed.
State lives in files under `~/.sauron/`.

The full model — registries, personas, providers, the sync plan, and state — is
in the **[domain documentation](spec/README.md)**.

## Documentation

- **[Domain model & concepts](spec/README.md)** — the canonical guide to what
  sauron is and how it works.
- **[Agent & contributor guide](AGENTS.md)** — how to work in this repo; routes
  to every rule and contract.
- License — [Apache-2.0](LICENSE).

# sauron

Distribute skills and agents to any AI coding assistant — across provider
boundaries — and keep them in sync.

sauron is a spec-driven CLI that delivers skills and agents from remote
registries (git, HTTP, or filesystem) to the assistant you use (Claude,
Zencoder, …). It groups them into shareable personas and reconciles your
provider to a desired set with a reviewable plan, manually or on a schedule.

Registry management (`add` / `list` / `describe` / `delete registry`) ships today;
the remaining commands are specified and being built — see the
[feature status index](spec/README.md#specifications).

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
State lives in files under `~/.sauron/`, including the **track file** that records
exactly what sauron installed and where it came from — the source of truth that
keeps `uninstall`, `sync`, and `upgrade` safe.

The full model — registries, personas, providers, the sync plan, and state — is
in the **[domain documentation](spec/README.md)**.

## Who this is for

This is a spec-driven project: behavior is specified before it is built, and the
documentation is organized so each reader can go straight to what they need.

- **New here?** Start with the [domain model](spec/README.md), then the
  [lifecycle walkthrough](spec/WORKFLOW.md) to see how a change flows from
  requirement to shipped code.
- **Developers / engineers** — the [architecture contract](spec/contracts/architecture.md)
  (layout, wiring, Use Case/Action), the [CLI contract](spec/contracts/cli.md),
  and [how to author specs](spec/AUTHORING.md). Day-to-day workflow is in
  [CONTRIBUTING](CONTRIBUTING.md).
- **Architects** — the [Constitution](CONSTITUTION.md) (governing principles) and
  the [architecture](spec/contracts/architecture.md) / [state](spec/contracts/state.md)
  contracts.
- **Security analysts** — the [security overview](spec/SECURITY.md): secret
  handling, TLS, write integrity, file permissions, and the HTTP registry API's
  auth model, with links to the rules that own them.

## Documentation

| Doc | What it covers |
|---|---|
| [Domain model & concepts](spec/README.md) | what sauron is, the model, and the feature status index |
| [Glossary](spec/GLOSSARY.md) | the canonical domain vocabulary |
| [Lifecycle walkthrough](spec/WORKFLOW.md) | one slice from requirement to shipped, verified code |
| [Constitution](CONSTITUTION.md) | governing principles (spec-driven, contracts, implementation, traceability) |
| [Spec authoring](spec/AUTHORING.md) | how specs and ADRs are written |
| [Contracts](spec/contracts/) | the binding [CLI](spec/contracts/cli.md) · [state](spec/contracts/state.md) · [architecture](spec/contracts/architecture.md) · [delivery](spec/contracts/delivery.md) contracts |
| [Security overview](spec/SECURITY.md) | the security posture, for analysts |
| [Contributing](CONTRIBUTING.md) | branching, commits, proposals, the verification gate |
| [Agent guide](AGENTS.md) | how AI coding agents work in this repo; routes to every rule |

License — [Apache-2.0](LICENSE).

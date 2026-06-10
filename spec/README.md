# Sauron

A simple orchestrator for delivering skills and agents.

## The Problem

Skills and agents are important artifacts when coding with agentic AI, and they matter even more for a team of developers who are expected to follow the same principles. This raises a practical question: how do you distribute these artifacts and keep them up to date across the team?

Claude has a marketplace, and other providers have their own, but each is restricted to its specific provider. Sauron ignores those boundaries. It makes it possible to share skills and agents — and keep them current — in any environment that needs them, regardless of provider.

## Concepts

**Repository** — A source that hosts skills and/or agents. It can be a remote Git repository, an HTTP server, or a filesystem directory. A repository must host at least one skill or agent and follow this structure:

```
.agents/[agent name]
.skills/[skill name]
```

**Persona** — It describes a group of people who share the same set of agents and skills. 

**Target** — The destination agentic AI provider. The target determines where skills and agents are persisted, since each provider stores them differently.

<br/>
<br/>

```
       ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
       │ Repository A │  │ Repository B │  │ Repository C │
       │ .agents/     │  │ .agents/     │  │ .agents/     │
       │ .skills/     │  │ .skills/     │  │ .skills/     │
       └──────┬───────┘  └──────┬───────┘  └──────┬───────┘
              │                 │                 │
              └─────────────────┼─────────────────┘
                                │  watch & pull latest
                                ▼
                         ┌─────────────┐
                         │   SAURON    │
                         └──────┬──────┘
                                │
              ┌─────────────────┴──────────────────┐
       with persona(s)                      without persona
              │                                    │
              ▼                                    │
   ┌────────────────────────┐                      │
   │ Persona(s)             │                      │
   │   optional; 1+ roles   │                      │  all agents
   │   delivers the union   │                      │  & skills
   │   of role-based subsets│                      │
   │   (e.g. Backend Dev)   │                      │
   └────────────┬───────────┘                      │
                │                                  │
                └─────────────────┬────────────────┘
                                  │  deliver to
                                  ▼
                  ┌───────────────────────────────┐
                  │ Target                        │
                  │   provider destination where  │
                  │   artifacts are persisted     │
                  │   (e.g. Claude)               │
                  └───────────────────────────────┘
```

## How It Works

Sauron is a command-line application that orchestrates delivery. It watches the configured repositories and keeps each target's skills and agents in sync with the latest versions. It can also register a cron job, so the whole process runs automatically on a schedule.

## Further reading

- [Spec authoring rules](AUTHORING.md) — spec types, numbering, required
  sections, EARS templates, glossary, and cross-link form.
- [Command line interface contract](contracts/cli.md) — command grammar,
  shared flags, exit status, output discipline, and the command index.
- [Constitution](../CONSTITUTION.md) — project and implementation principles.

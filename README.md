# sauron

A spec-driven command line interface that distributes skills and agents from
remote repositories to AI coding targets.

Sauron watches registered **repositories** (git, HTTP, or filesystem sources
hosting `.skills/` and `.agents/` directories), groups artifacts into
**personas**, and keeps the configured **target** (e.g. Claude, Zencoder) in
sync — manually via `sauron sync` or on a schedule via `sauron cron sync`.

## Documentation

- [Domain model and concepts](spec/README.md)
- [Command line interface contract](spec/contracts/cli.md)
- [Spec authoring rules](spec/AUTHORING.md)
- [Constitution](CONSTITUTION.md)

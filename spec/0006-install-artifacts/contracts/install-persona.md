# `install persona` — command line

```
sauron install persona <registry> <name>...
```

Install named personas from a registry into the active provider. Installing a
persona resolves its membership and installs each member.

## Arguments

| Argument | Required | Meaning |
|---|---|---|
| `<registry>` | yes | The source registry |
| `<name>...` | yes | One or more persona names to install |

## Output

The plan grouped under `personas:` and the `skills:`/`agents:` headings for the
members it brings in, `+` for additions and `~` for updates, followed by a summary
count. A persona member the registry does not offer is reported without stopping
the run.

## Exit codes

| Code | Condition |
|---|---|
| `0` | The named personas (and their members) were installed or already current |
| `2` | Missing/invalid arguments or flags |
| `1` | No provider is set, the registry is unreachable, or the install could not be persisted |

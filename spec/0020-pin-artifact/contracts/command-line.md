# Contract: Command Line — Pin Artifact

Conventions: [CLI conventions](../../AUTHORING.md).

**Spec**: [Pin Artifact](../spec.md)

Defines the command-line interface for pinning and unpinning a skill or agent to
a registry. This is the user-facing contract only.

## Synopsis

```
sauron pin skill <name> <registry> [--reconcile]
sauron pin agent <name> <registry> [--reconcile]
sauron unpin skill <name> [--reconcile] [--dry-run]
sauron unpin agent <name> [--reconcile] [--dry-run]
```

Command hierarchy: `sauron` (root) → `pin`/`unpin` (verb) → `skill`/`agent`
(noun) → `<name> [<registry>]`.

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `<name>` | Yes | The artifact name. Realizes [spec](../spec.md) FR-002, FR-003, FR-007. |
| `<registry>` | Yes, for `pin` | The registry to pin the artifact to. Realizes [spec](../spec.md) FR-002, FR-007, FR-008, FR-009. |

## Flags

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--reconcile` | No | false | Apply the change now by reconciling the affected artifact (a scoped sync), installing it when the pin targets a not-yet-installed artifact. Realizes [spec](../spec.md) FR-012. |
| `--dry-run` | No | false | `unpin` only — print the current pinned registry and the registry priority would pick next, without changing anything. Realizes [spec](../spec.md) FR-013. |

## Output

- **Success**: a single confirmation line naming the artifact and its source registry; for `unpin --dry-run`, the current pinned registry and the next priority registry.
- **Failure**: a single human-readable message on stderr.

## Exit codes

Exit-status meanings are owned by the [CLI conventions](../../AUTHORING.md);
this table refines which conditions map to each code.

| Code | Meaning | Realizes |
|------|---------|----------|
| `0` | Pinned, unpinned, an already-pinned/already-unpinned no-op, or a `--dry-run` | [spec](../spec.md) FR-002, FR-003, FR-004, FR-013 |
| `2` | Usage error — missing name, or missing registry for `pin` | [spec](../spec.md) FR-007 |
| `1` | Runtime error — unknown registry, artifact not offered there, artifact not installed (without `--reconcile`), or the track file is unreadable/unwritable | [spec](../spec.md) FR-008, FR-009, FR-010, FR-011 |

## Examples

```
# Pin a skill to a specific registry (recorded; reconciled on next sync)
$ sauron pin skill code-review team-internal
Pinned skill 'code-review' to registry 'team-internal'.

# Pin and apply immediately
$ sauron pin agent triager team-internal --reconcile
Pinned agent 'triager' to registry 'team-internal' and reconciled.

# Unpin (re-resolves by priority on next sync)
$ sauron unpin skill code-review
Unpinned skill 'code-review'.

# Preview an unpin
$ sauron unpin skill code-review --dry-run
skill 'code-review' is pinned to 'team-internal'; priority would pick 'team-deploy'.

# Missing registry for pin (usage error, exit 2)
$ sauron pin skill code-review
Error: pin requires a registry

# Registry does not offer the artifact (runtime error, exit 1)
$ sauron pin skill code-review team-empty
Error: registry 'team-empty' does not offer skill 'code-review'
```

# Data Model: Configuration — Sauron Settings (Pruning)

**Spec**: `../spec.md` (Prune Orphaned Skills and Agents)
**Status**: Draft

Describes the data the Prune feature reads and updates. Prune compares installed artifacts (`track.yaml`) against registered repositories (`settings.yaml`) and removes the orphans.

## Inputs

### Registered repositories — `~/.sauron/settings.yaml`

Read-only for prune. The set of names in `repositories[]` is the registered set; an artifact is orphaned when its source repository name is not in this set.

### Tracking record — `~/.sauron/track.yaml`

The record of installed artifacts and their provenance. Created and maintained by the sync feature (`0011-sync`, which owns the full schema); prune reads it and removes entries for pruned artifacts.

- **Path**: `~/.sauron/track.yaml` (home directory resolved per platform).
- **Format**: a single YAML document.

Top-level document:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `installed` | array of Installed Artifact | Yes | Delivered artifacts. Empty array when none. |

Installed Artifact entry:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type` | string | Yes | `skill` or `agent`. |
| `name` | string | Yes | Artifact name, as installed. |
| `target` | string | Yes | Provider the artifact was delivered to (e.g. `claude`, `zencoder`). |
| `path` | string | Yes | Where it was installed. |
| `repository` | string | Yes | Name of the source repository (provenance). |
| `persona` | string | No | Persona that brought the artifact into scope; absent when synced without personas. Not used by prune's orphan test. |

## Operation

- For each installed artifact of the requested type(s), if its `repository` is not in the registered set, the artifact is orphaned. Realizes FR-004.
- Orphaned artifacts are deleted from their `path` and their entries removed from `installed[]`. Realizes FR-005. With `--dry-run`, nothing is deleted or written. Realizes FR-007.
- Artifacts whose `repository` is still registered are left untouched. Realizes FR-008.

## Write semantics

- Prune writes only `track.yaml`, never `settings.yaml`.
- Updates are atomic: serialize to a temporary file in `~/.sauron/`, then rename over `track.yaml`. The file is left untouched on `--dry-run` or when nothing is orphaned.

## Example (`track.yaml`)

```yaml
installed:
  - type: skill
    name: code-review
    target: claude
    path: ~/.claude/skills/code-review
    repository: team-deploy
    persona: backend-developer
  - type: agent
    name: triager
    target: claude
    path: ~/.claude/agents/triager.md
    repository: old-http
```

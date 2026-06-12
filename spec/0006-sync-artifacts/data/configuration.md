# Data Model: Configuration — Sauron Settings (Sync)

**Spec**: [Sync](../spec.md)

Describes the data sync artifacts reads and the tracking record it owns. Sync artifacts reads `settings.yaml` (registries and installed personas), delivers artifacts to the provider's locations, and maintains `track.yaml`.

## Inputs — `~/.sauron/settings.yaml`

Read-only for sync artifacts:

- `registries[]` — the sources of artifacts; the `priority` field resolves same-named artifacts ([ADR-0001](../architecture/ADR-0001-conflict-resolution-by-registry-priority.md)).
- `personas[]` — the installed persona definitions that scope the desired set; persona ordering follows [priority model](../../AUTHORING.md#priority-model).
- `provider` — the active provider to deliver to (`claude` by default; managed by [set provider](../../0009-set-provider/spec.md)). Realizes [spec](../spec.md) FR-007.

## Tracking record — `~/.sauron/track.yaml`

The record of installed artifacts and their provenance. **Sync artifacts creates and maintains this file**; other features (e.g. [prune](../../0004-prune/spec.md)) read it.

- **Path**: `~/.sauron/track.yaml` (home directory resolved per platform).
- **Format**: a single YAML document.
- **Lifecycle**: created on the first successful sync artifacts run if absent.

Top-level document:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `installed` | array of Installed Artifact | Yes | Delivered artifacts. Empty array when none. |

Installed Artifact entry:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type` | string | Yes | `skill` or `agent`. |
| `name` | string | Yes | Artifact name, as installed. |
| `provider` | string | Yes | Provider the artifact was delivered to (`claude` or `zencoder`). |
| `path` | string | Yes | Where it was installed (the provider's location for this artifact). |
| `registry` | string | Yes | Name of the source registry (provenance; the conflict winner per [ADR-0001](../architecture/ADR-0001-conflict-resolution-by-registry-priority.md)). |
| `persona` | string | No | Installed persona that brought the artifact into the desired set; when several do, the highest-precedence installed persona; absent when synced without personas. Realizes [spec](../spec.md) FR-006. |

An entry is identified by (`provider`, `type`, `name`) — the same artifact delivered to two providers yields two entries.

## Operation

- The desired set is computed per [spec](../spec.md) FR-002–FR-004 and FR-016, and compared against the entries whose `provider` matches the active global provider; the difference is the plan. Realizes [spec](../spec.md) FR-005.
- Applying the plan installs/updates artifacts at their `path`, removes tracked artifacts no longer desired, and rewrites their entries. Only tracked artifacts are ever removed. Realizes [spec](../spec.md) FR-006, FR-009.
- With `--dry-run`, neither the environment nor `track.yaml` is touched. Realizes [spec](../spec.md) FR-017.

## Write semantics

- Sync artifacts writes only `track.yaml`, never `settings.yaml`.
- Updates are atomic: serialize to a temporary file in `~/.sauron/`, then rename over `track.yaml`.

## Example (`track.yaml`)

```yaml
installed:
  - type: skill
    name: design-oas3
    provider: zencoder
    path: /home/user/.zencoder/skills/design-oas3
    registry: team-deploy
    persona: backend-developer
  - type: agent
    name: software-engineer
    provider: zencoder
    path: /home/user/.zencoder/agents/software-engineer
    registry: team-deploy
    persona: backend-developer
  - type: skill
    name: code-review
    provider: claude
    path: /home/user/.claude/skills/code-review
    registry: team-deploy
    persona: backend-developer
```

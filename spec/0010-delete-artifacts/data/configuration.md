# Data Model: Configuration — Delete Artifacts (track.yaml)

**Spec**: [Delete Artifacts](../spec.md)

This feature removes installed-artifact entries from `track.yaml`; it operates on
the tracking record only. The schema is owned by the
[configuration data contract](../../contracts/configuration.md#trackyaml); this
document does not restate it.

## Reads

- `track.yaml` `items` — the artifacts in scope. The noun narrows by `type`:
  `artifacts` = both, `skills` = `type: skill`, `agents` = `type: agent`.
  Without `--persona`, every entry of that type is in scope (all providers); with
  `--persona <name>`, only entries whose `persona` field equals `<name>`, whether
  or not that persona is still installed in `personas.yaml`.

## Owns

- Nothing. `track.yaml` is owned by
  [sync artifacts](../../0006-sync-artifacts/spec.md); this feature only removes
  entries it has recorded.

## Writes

- `track.yaml` `items` — each in-scope artifact is deleted from its `path`
  and its entry removed. A missing file is treated as nothing to delete. With
  `--dry-run`, nothing is deleted or written. The file is left untouched on
  `--dry-run` or when nothing is in scope. No other configuration file is read or
  written.

## Notes

Configuration is now split across files per the
[configuration data contract](../../contracts/configuration.md); file references
updated accordingly.

@no-sandbox
Feature: Describe provider
  As an operator
  I want to see the active provider's full detail
  So that I can review where sauron installs artifacts

  # FR-001 — describe shows the active provider, with the directory derived from
  # the provider name (claude -> ~/.claude).
  Scenario: shows the active provider with its derived directory
    Given the settings file contains:
      """
      apiVersion: sauron.raitonbl.com/v1
      kind: Provider
      metadata:
        name: claude
      """
    When the user runs sauron describe provider
    Then the command succeeds
    And the descriptor shows name as claude
    And the descriptor shows directory as ~/.claude

  # FR-002 — --fields selects and orders the displayed fields, name always first.
  Scenario: shows only the selected fields with --fields
    Given the settings file contains:
      """
      apiVersion: sauron.raitonbl.com/v1
      kind: Provider
      metadata:
        name: claude
        createdAt: "2026-06-21T07:30:00Z"
      """
    When the user runs sauron describe provider --fields directory,name
    Then the command succeeds
    And the descriptor shows name as claude
    And the descriptor shows directory as ~/.claude
    And the output does not contain createdAt

  # FR-003 — no provider set reports so and exits successfully (exit 0, not an error).
  Scenario: reports when no provider is set
    When the user runs sauron describe provider
    Then the command succeeds
    And the output reports no provider is set

  # FR-005 — an invalid --fields value is a usage error (exit 2).
  Scenario: rejects an invalid --fields value
    Given the settings file contains:
      """
      apiVersion: sauron.raitonbl.com/v1
      kind: Provider
      metadata:
        name: claude
      """
    When the user runs sauron describe provider --fields bogus
    Then the command exits with status 2

  # FR-001 — the full descriptor of a synced provider renders verbatim: name first,
  # the derived directory, the key-sorted labels section, the audit timestamps, and
  # the sync timestamps, through the shared aligned label: value renderer.
  Scenario: renders the full descriptor of a synced provider
    Given the settings file contains:
      """
      apiVersion: sauron.raitonbl.com/v1
      kind: Provider
      metadata:
        name: claude
        createdAt: "2026-06-21T07:30:00Z"
        lastUpdatedAt: "2026-06-22T08:00:00Z"
        labels:
          team: backend
      spec:
        lastSyncedAt: "2026-06-25T09:15:00Z"
        lastSyncAttemptAt: "2026-06-26T06:00:00Z"
      """
    When the user runs sauron describe provider
    Then the command succeeds
    And the descriptor shows name as claude
    And the descriptor shows directory as ~/.claude
    And the descriptor shows createdAt as 2026-06-21T07:30:00Z
    And the descriptor shows lastUpdatedAt as 2026-06-22T08:00:00Z
    And the descriptor shows lastSyncedAt as 2026-06-25T09:15:00Z
    And the descriptor shows lastSyncAttemptAt as 2026-06-26T06:00:00Z
    And the descriptor reads:
      """
      name:               claude
      directory:          ~/.claude
      labels:
        team:             backend
      createdAt:          2026-06-21T07:30:00Z
      lastUpdatedAt:      2026-06-22T08:00:00Z
      lastSyncedAt:       2026-06-25T09:15:00Z
      lastSyncAttemptAt:  2026-06-26T06:00:00Z
      """

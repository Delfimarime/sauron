Feature: List agents
  As a developer
  I want to see what agents are installed
  So that I know what I already have without browsing the registry

  # FR-001 — lists every installed agent of the chosen kind, reading track.yaml.
  Scenario: lists every installed agent
    Given the track file contains:
      """
      apiVersion: sauron.raitonbl.com/v1
      kind: Agent
      metadata:
        name: code-reviewer
      spec:
        version: 3af1c2e
        path: agents/code-reviewer
        installedAt: "2026-06-14T00:00:00Z"
        updatedAt: "2026-06-14T00:00:00Z"
      ---
      apiVersion: sauron.raitonbl.com/v1
      kind: Agent
      metadata:
        name: release-bot
      spec:
        version: 91fbe20
        path: agents/release-bot
        installedAt: "2026-06-16T00:00:00Z"
        updatedAt: "2026-06-16T00:00:00Z"
      """
    When the user runs sauron list agents
    Then the command succeeds
    And the artifacts list code-reviewer
    And the artifacts list release-bot

  # FR-002 — the default view shows only name; --fields selects and reorders the
  # displayed columns, name always present and first regardless of the requested
  # order.
  Scenario: shows name by default, and --fields selects and reorders columns
    Given the track file contains:
      """
      apiVersion: sauron.raitonbl.com/v1
      kind: Agent
      metadata:
        name: code-reviewer
      spec:
        version: 3af1c2e
        path: agents/code-reviewer
        installedAt: "2026-06-14T00:00:00Z"
        updatedAt: "2026-06-14T00:00:00Z"
      """
    When the user runs sauron list agents
    Then the command succeeds
    And the artifacts list code-reviewer
    And the output does not contain 3af1c2e
    When the user runs sauron list agents --fields lastUpdatedAt,version
    Then the command succeeds
    And the artifacts list code-reviewer 2026-06-14T00:00:00Z 3af1c2e

  # FR-003 — --page/--limit page the results; the paging line reports the applied
  # window with no total.
  Scenario: pages results with --page and --limit
    Given the track file contains:
      """
      apiVersion: sauron.raitonbl.com/v1
      kind: Agent
      metadata:
        name: alpha
      spec:
        version: v1.0.0
        path: agents/alpha
        installedAt: "2026-06-10T00:00:00Z"
        updatedAt: "2026-06-10T00:00:00Z"
      ---
      apiVersion: sauron.raitonbl.com/v1
      kind: Agent
      metadata:
        name: bravo
      spec:
        version: v1.0.0
        path: agents/bravo
        installedAt: "2026-06-11T00:00:00Z"
        updatedAt: "2026-06-11T00:00:00Z"
      """
    When the user runs sauron list agents --page 2 --limit 1
    Then the command succeeds
    And the artifacts list bravo
    And the paging line reads showing 2–2 (page 2, limit 1)

  # FR-003 — paging past the end yields no rows and the empty-page paging line.
  Scenario: reports an empty page when paged past the end
    Given the track file contains:
      """
      apiVersion: sauron.raitonbl.com/v1
      kind: Agent
      metadata:
        name: alpha
      spec:
        version: v1.0.0
        path: agents/alpha
        installedAt: "2026-06-10T00:00:00Z"
        updatedAt: "2026-06-10T00:00:00Z"
      """
    When the user runs sauron list agents --page 9 --limit 20
    Then the command succeeds
    And the paging line reads showing 0 results (page 9, limit 20)

  # FR-004 — --search includes only entries whose name contains the term
  # (case-insensitive).
  Scenario: filters entries with --search
    Given the track file contains:
      """
      apiVersion: sauron.raitonbl.com/v1
      kind: Agent
      metadata:
        name: code-reviewer
      spec:
        version: 3af1c2e
        path: agents/code-reviewer
        installedAt: "2026-06-10T00:00:00Z"
        updatedAt: "2026-06-10T00:00:00Z"
      ---
      apiVersion: sauron.raitonbl.com/v1
      kind: Agent
      metadata:
        name: release-bot
      spec:
        version: 91fbe20
        path: agents/release-bot
        installedAt: "2026-06-11T00:00:00Z"
        updatedAt: "2026-06-11T00:00:00Z"
      ---
      apiVersion: sauron.raitonbl.com/v1
      kind: Agent
      metadata:
        name: sql-reviewer
      spec:
        version: 5be9a10
        path: agents/sql-reviewer
        installedAt: "2026-06-12T00:00:00Z"
        updatedAt: "2026-06-12T00:00:00Z"
      """
    When the user runs sauron list agents --search review
    Then the command succeeds
    And the artifacts list code-reviewer
    And the artifacts list sql-reviewer
    And the output does not contain release-bot

  # FR-005 — --sort lastUpdatedAt --order desc orders the entries before paging.
  Scenario: orders entries with --sort lastUpdatedAt --order desc
    Given the track file contains:
      """
      apiVersion: sauron.raitonbl.com/v1
      kind: Agent
      metadata:
        name: alpha
      spec:
        version: v1.0.0
        path: agents/alpha
        installedAt: "2026-06-10T00:00:00Z"
        updatedAt: "2026-06-10T00:00:00Z"
      ---
      apiVersion: sauron.raitonbl.com/v1
      kind: Agent
      metadata:
        name: bravo
      spec:
        version: v1.0.0
        path: agents/bravo
        installedAt: "2026-06-14T00:00:00Z"
        updatedAt: "2026-06-14T00:00:00Z"
      ---
      apiVersion: sauron.raitonbl.com/v1
      kind: Agent
      metadata:
        name: charlie
      spec:
        version: v1.0.0
        path: agents/charlie
        installedAt: "2026-06-18T00:00:00Z"
        updatedAt: "2026-06-18T00:00:00Z"
      """
    When the user runs sauron list agents --sort lastUpdatedAt --order desc --limit 1
    Then the command succeeds
    And the artifacts list charlie
    And the paging line reads showing 1–1 (page 1, limit 1)
    And the output does not contain alpha

  # FR-006 — nothing installed prints an empty result and the empty-page paging
  # line, and exits successfully.
  Scenario: shows an empty result when nothing is installed
    When the user runs sauron list agents
    Then the command succeeds
    And the paging line reads showing 0 results (page 1, limit 20)

  # FR-007 — an unreadable/malformed track.yaml fails with a runtime error
  # (exit 1).
  Scenario: fails with a runtime error when track.yaml is malformed
    Given the track file contains:
      """
      not: [valid
      """
    When the user runs sauron list agents
    Then the command exits with status 1

  # FR-008 — an invalid --page/--limit/--sort/--order value is a usage error
  # (exit 2), resolved before track.yaml is ever read.
  Scenario: rejects invalid flags
    When the user runs sauron list agents --page 0
    Then the command exits with status 2
    When the user runs sauron list agents --limit 0
    Then the command exits with status 2
    When the user runs sauron list agents --sort bogus
    Then the command exits with status 2
    When the user runs sauron list agents --order sideways
    Then the command exits with status 2
    When the user runs sauron list agents --fields bogus
    Then the command exits with status 2

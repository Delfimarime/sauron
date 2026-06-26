@no-sandbox
Feature: List catalogue
  As a developer
  I want to browse the registry's offered skills and agents
  So that I can see what is available before installing

  # FR-001 — list the agents the registry offers, each as a NAME KIND row.
  Scenario: lists every agent the registry offers
    Given the registry offers the following agents:
      """
      # file: .agents/code-reviewer.yaml
      apiVersion: sauron.raitonbl.com/v1
      kind: Agent
      metadata:
        name: code-reviewer
      # file: .agents/release-bot.yaml
      apiVersion: sauron.raitonbl.com/v1
      kind: Agent
      metadata:
        name: release-bot
      """
    When the user runs sauron list catalogue agent
    Then the command succeeds
    And the catalogue lists code-reviewer agent
    And the catalogue lists release-bot agent

  # FR-001 — skills list as NAME KIND.
  Scenario: lists the skills the registry offers as NAME KIND
    Given the registry offers the following skills:
      """
      # file: .skills/go-style.yaml
      apiVersion: sauron.raitonbl.com/v1
      kind: Skill
      metadata:
        name: go-style
      # file: .skills/sql-review.yaml
      apiVersion: sauron.raitonbl.com/v1
      kind: Skill
      metadata:
        name: sql-review
      """
    When the user runs sauron list catalogue skill
    Then the command succeeds
    And the catalogue lists go-style skill
    And the catalogue lists sql-review skill

  # FR-002 — --page/--limit page the results; the paging line reports the applied
  # window with no total.
  Scenario: pages results with --page and --limit
    Given the registry offers the following agents:
      """
      # file: .agents/alpha.yaml
      apiVersion: sauron.raitonbl.com/v1
      kind: Agent
      metadata:
        name: alpha
      # file: .agents/bravo.yaml
      apiVersion: sauron.raitonbl.com/v1
      kind: Agent
      metadata:
        name: bravo
      """
    When the user runs sauron list catalogue agent --page 2 --limit 1
    Then the command succeeds
    And the catalogue lists bravo agent
    And the paging line reads showing 2–2 (page 2, limit 1)

  # FR-002 — paging past the end yields no rows and the empty-page paging line.
  Scenario: reports an empty page when paged past the end
    Given the registry offers the following agents:
      """
      # file: .agents/alpha.yaml
      apiVersion: sauron.raitonbl.com/v1
      kind: Agent
      metadata:
        name: alpha
      """
    When the user runs sauron list catalogue agent --page 9 --limit 20
    Then the command succeeds
    And the paging line reads showing 0 results (page 9, limit 20)

  # FR-003 — --search includes only entries whose name contains the term
  # (case-insensitive).
  Scenario: filters entries with --search
    Given the registry offers the following skills:
      """
      # file: .skills/code-review.yaml
      apiVersion: sauron.raitonbl.com/v1
      kind: Skill
      metadata:
        name: code-review
      # file: .skills/go-style.yaml
      apiVersion: sauron.raitonbl.com/v1
      kind: Skill
      metadata:
        name: go-style
      # file: .skills/sql-review.yaml
      apiVersion: sauron.raitonbl.com/v1
      kind: Skill
      metadata:
        name: sql-review
      """
    When the user runs sauron list catalogue skill --search rev
    Then the command succeeds
    And the catalogue lists code-review skill
    And the catalogue lists sql-review skill
    And the output does not contain go-style

  # FR-004 — --sort name --order desc orders the entries before paging.
  Scenario: orders entries with --sort name --order desc
    Given the registry offers the following skills:
      """
      # file: .skills/alpha.yaml
      apiVersion: sauron.raitonbl.com/v1
      kind: Skill
      metadata:
        name: alpha
      # file: .skills/bravo.yaml
      apiVersion: sauron.raitonbl.com/v1
      kind: Skill
      metadata:
        name: bravo
      # file: .skills/charlie.yaml
      apiVersion: sauron.raitonbl.com/v1
      kind: Skill
      metadata:
        name: charlie
      """
    When the user runs sauron list catalogue skill --sort name --order desc --limit 1
    Then the command succeeds
    And the catalogue lists charlie skill
    And the paging line reads showing 1–1 (page 1, limit 1)
    And the output does not contain alpha

  # FR-006 — when no registry is set, list catalogue fails with a runtime error
  # (exit 1); there is no offline catalogue.
  Scenario: fails when no registry is set
    When the user runs sauron list catalogue agent
    Then the command exits with status 1

  # FR-005 — a registry pointing at an unreachable/absent source fails with a runtime
  # error (exit 1).
  Scenario: fails when the registry source is unreachable
    Given the registry is configured:
      | transport  | source                 |
      | filesystem | /nonexistent/acme/repo |
    When the user runs sauron list catalogue agent
    Then the command exits with status 1

  # A bare `list catalogue` (no kind noun) shows help and exits 0, as any command
  # group does. FR-007 — an invalid --page/--order value — is a usage error (exit 2).
  Scenario: shows help for a missing kind noun, and rejects invalid flags
    Given the registry offers the following agents:
      """
      # file: .agents/code-reviewer.yaml
      apiVersion: sauron.raitonbl.com/v1
      kind: Agent
      metadata:
        name: code-reviewer
      """
    When the user runs sauron list catalogue
    Then the command exits with status 0
    When the user runs sauron list catalogue agent --page 0
    Then the command exits with status 2
    When the user runs sauron list catalogue agent --order sideways
    Then the command exits with status 2

@no-sandbox
Feature: List catalogue
  As a developer
  I want to browse a registry's offered skills, agents, and personas
  So that I can see what is available before installing

  # FR-001 — list the agents a registry offers, each as a NAME KIND row.
  Scenario: lists every agent a registry offers
    Given the registry acme offers the following agents:
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
    When the user runs sauron list catalogue agent acme
    Then the command succeeds
    And the catalogue lists code-reviewer agent
    And the catalogue lists release-bot agent

  # FR-001 — skills list as NAME KIND; personas list as NAME MEMBERS, summarizing
  # the skills/agents membership each persona manifest declares.
  Scenario: lists skills as NAME KIND and personas as NAME MEMBERS
    Given the registry acme offers the following skills:
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
    And the registry acme offers the following personas:
      """
      # file: .personas/backend-dev.yaml
      apiVersion: sauron.raitonbl.com/v1
      kind: Persona
      metadata:
        name: backend-dev
      spec:
        members:
          skills:
            - go-style
            - sql-review
          agents:
            - code-reviewer
      """
    When the user runs sauron list catalogue skill acme
    Then the command succeeds
    And the catalogue lists go-style skill
    And the catalogue lists sql-review skill
    When the user runs sauron list catalogue persona acme
    Then the command succeeds
    And the output contains backend-dev
    And the output contains skills: go-style, sql-review; agents: code-reviewer

  # FR-002 — --page/--limit page the results; the paging line reports the applied
  # window with no total.
  Scenario: pages results with --page and --limit
    Given the registry acme offers the following agents:
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
    When the user runs sauron list catalogue agent acme --page 2 --limit 1
    Then the command succeeds
    And the catalogue lists bravo agent
    And the paging line reads showing 2–2 (page 2, limit 1)

  # FR-002 — paging past the end yields no rows and the empty-page paging line.
  Scenario: reports an empty page when paged past the end
    Given the registry acme offers the following agents:
      """
      # file: .agents/alpha.yaml
      apiVersion: sauron.raitonbl.com/v1
      kind: Agent
      metadata:
        name: alpha
      """
    When the user runs sauron list catalogue agent acme --page 9 --limit 20
    Then the command succeeds
    And the paging line reads showing 0 results (page 9, limit 20)

  # FR-003 — --search includes only entries whose name contains the term
  # (case-insensitive).
  Scenario: filters entries with --search
    Given the registry acme offers the following skills:
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
    When the user runs sauron list catalogue skill acme --search rev
    Then the command succeeds
    And the catalogue lists code-review skill
    And the catalogue lists sql-review skill
    And the output does not contain go-style

  # FR-004 — --sort name --order desc orders the entries before paging.
  Scenario: orders entries with --sort name --order desc
    Given the registry acme offers the following skills:
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
    When the user runs sauron list catalogue skill acme --sort name --order desc --limit 1
    Then the command succeeds
    And the catalogue lists charlie skill
    And the paging line reads showing 1–1 (page 1, limit 1)
    And the output does not contain alpha

  # FR-006 — a name matching no registry fails with a runtime error (exit 1).
  Scenario: fails when no registry of that name exists
    Given the registry acme offers the following agents:
      """
      # file: .agents/code-reviewer.yaml
      apiVersion: sauron.raitonbl.com/v1
      kind: Agent
      metadata:
        name: code-reviewer
      """
    When the user runs sauron list catalogue agent ghost
    Then the command exits with status 1
    And the output contains ghost

  # FR-005 — a registry pointing at an unreachable/absent source fails with a
  # runtime error (exit 1); there is no offline catalogue.
  Scenario: fails when the registry source is unreachable
    Given the following registries are configured:
      | name | transport  | uri                    |
      | acme | filesystem | /nonexistent/acme/repo |
    When the user runs sauron list catalogue agent acme
    Then the command exits with status 1

  # FR-007 — a missing <registry> arg, or an invalid --page/--order, is a usage
  # error (exit 2), reported before the command executes.
  Scenario: rejects missing or invalid arguments and flags
    Given the registry acme offers the following agents:
      """
      # file: .agents/code-reviewer.yaml
      apiVersion: sauron.raitonbl.com/v1
      kind: Agent
      metadata:
        name: code-reviewer
      """
    When the user runs sauron list catalogue agent
    Then the command exits with status 2
    When the user runs sauron list catalogue agent acme --page 0
    Then the command exits with status 2
    When the user runs sauron list catalogue agent acme --order sideways
    Then the command exits with status 2

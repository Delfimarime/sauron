@no-sandbox
Feature: List registries
  As an operator
  I want to list the registered registries
  So that I can review which sources sauron resolves artifacts from

  # FR-001 / FR-002 — every registry is listed, one row each, with the default
  # name, transport, and uri columns.
  Scenario: lists the registered registries in the default columns
    Given the following registries are configured:
      | name     | transport  | uri                               |
      | acme     | git        | git@github.com:acme/artifacts.git |
      | internal | http       | https://reg.example.com/          |
    When the user runs sauron list registries
    Then the command succeeds
    And the output contains NAME
    And the output contains TRANSPORT
    And the output contains URI
    And the output contains acme
    And the output contains internal

  # FR-002 — --fields selects and reorders the displayed columns.
  Scenario: shows only the selected columns with --fields
    Given the following registries are configured:
      | name     | transport  | uri                               |
      | acme     | git        | git@github.com:acme/artifacts.git |
      | internal | http       | https://reg.example.com/          |
    When the user runs sauron list registries --fields name,uri
    Then the command succeeds
    And the output contains NAME
    And the output contains URI
    And the registries are listed in order: acme, internal

  # FR-003 — --search keeps only the registries whose name matches the term
  # (case-insensitive).
  Scenario: filters by a case-insensitive search term
    Given the following registries are configured:
      | name     | transport  | uri                               |
      | acme     | git        | git@github.com:acme/artifacts.git |
      | internal | http       | https://reg.example.com/          |
    When the user runs sauron list registries --search ACME
    Then the command succeeds
    And the output contains acme
    And the registries are listed in order: acme

  # FR-004 — --sort transport --order desc orders the rows by transport descending.
  Scenario: sorts by transport in descending order
    Given the following registries are configured:
      | name     | transport  | uri                               |
      | acme     | git        | git@github.com:acme/artifacts.git |
      | internal | http       | https://reg.example.com/          |
    When the user runs sauron list registries --sort transport --order desc
    Then the command succeeds
    And the registries are listed in order: internal, acme

  # FR-005 — no registry configured prints no output and exits 0.
  Scenario: prints nothing when no registry is configured
    When the user runs sauron list registries
    Then the command succeeds
    And the output is empty

  # FR-006 — a corrupt registries.yaml fails with a runtime error (exit 1). The
  # malformed document is designed here explicitly (spec.transport opens a flow
  # sequence that is never closed) so the file under test is visible in the scenario.
  Scenario: fails with a runtime error when the state file is corrupt
    Given the registries file contains:
      """
      apiVersion: sauron.raitonbl.com/v1
      kind: Registry
      metadata:
        name: broken
      spec:
        transport: [this is not closed
        uri:
      """
    When the user runs sauron list registries
    Then the command exits with status 1

  # FR-007 — an invalid --sort value is a usage error (exit 2).
  Scenario: rejects an invalid --sort value
    Given the following registries are configured:
      | name     | transport  | uri                               |
      | acme     | git        | git@github.com:acme/artifacts.git |
    When the user runs sauron list registries --sort uri
    Then the command exits with status 2

  # write-then-read — the mandatory black-box arrange: add a registry, then list it.
  Scenario: lists a registry that was added through the command
    Given a filesystem registry
    And the filesystem registry hosts a skill named go-style
    When the user adds the filesystem registry acme from #{.folder.default.path}
    And the user runs sauron list registries
    Then the command succeeds
    And the output contains acme
    And the registries are listed in order: acme

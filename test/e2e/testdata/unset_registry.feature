Feature: Unset registry
  As an operator
  I want to disconnect the configured source
  So that sauron stops resolving artifacts from it, while what it installed stays

  # FR-001 / FR-002 — unset removes the Registry document and confirms; every
  # already-installed artifact is left in place. The registry is produced black-box
  # (set registry) and the installed skill is seeded into track.yaml, which unset
  # only preserves and never writes.
  Scenario: removes the configured registry and preserves installed artifacts
    Given an http server hosting a registry
    And the http server hosts a skill named go-style
    And a tracked skill named sauron-acme-go-style
    When the user sets the http registry from #{.webserver.default.url}
    And the user runs sauron unset registry
    Then the command succeeds
    And the output contains registry unset; installed artifacts preserved
    And there is no registry
    And the skill sauron-acme-go-style is still tracked

  # FR-005 — unsetting when no registry is configured exits 0 and reports that
  # nothing was unset.
  Scenario: reports nothing was unset when no registry is configured
    When the user runs sauron unset registry
    Then the command succeeds
    And the output reports nothing was unset

  # FR-004 — --dry-run previews without changing state: the registry survives.
  Scenario: dry-run previews without changing state
    Given an http server hosting a registry
    And the http server hosts a skill named go-style
    When the user sets the http registry from #{.webserver.default.url}
    And the user runs sauron unset registry --dry-run
    Then the command succeeds
    And there is exactly one registry

  # FR-006 — an unexpected positional argument is a usage error (exit 2).
  Scenario: rejects an unexpected argument
    When the user runs sauron unset registry acme
    Then the command exits with status 2

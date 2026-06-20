@git
Feature: Add a git registry
  As an operator
  I want to register a git registry over ssh
  So that sauron can resolve artifacts from a git remote

  Scenario: adds a git registry over ssh
    Given a git server hosting a registry
    And the git server hosts the directory testdata/registries/acme
    When the user adds the git registry acme from #{.git.default.url}
    Then the command succeeds
    And there is exactly one registry
    And a registry named acme exists
    And the registry acme has transport git

  Scenario: adds a git registry pinned to a ref
    Given a git server hosting a registry
    And the git server hosts the directory testdata/registries/acme
    When the user adds the git registry acme from #{.git.default.url} pinned to v1.0.0
    Then the command succeeds
    And the registry acme is described by:
      | field    | value    |
      | spec.ref | v1.0.0   |

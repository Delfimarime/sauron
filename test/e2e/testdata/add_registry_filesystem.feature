@no-sandbox
Feature: Add a filesystem registry
  As an operator
  I want to register a local filesystem registry
  So that sauron can resolve artifacts from a directory

  Scenario: adds a filesystem registry from a local folder
    Given a filesystem registry
    And the filesystem registry hosts a skill named go-style
    When the user adds the filesystem registry acme from #{.folder.default.path}
    Then the command succeeds
    And there is exactly one registry
    And a registry named acme exists
    And the registry acme has transport filesystem
    And the registry acme is described by:
      | field          | value                  |
      | kind           | Registry               |
      | apiVersion     | sauron.raitonbl.com/v1 |
      | spec.transport | filesystem             |

  Scenario: adds a filesystem registry from an authored content directory
    Given a filesystem registry
    And the filesystem registry hosts the directory testdata/registries/acme
    When the user adds the filesystem registry acme from #{.folder.default.path}
    Then the command succeeds
    And a registry named acme exists
    And the registry acme has transport filesystem

  Scenario: fails when the registry hosts no artifacts
    Given a filesystem registry
    When the user adds the filesystem registry empty from #{.folder.default.path}
    Then the command fails because the registry hosts no artifacts

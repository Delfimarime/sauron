@no-sandbox
Feature: Set a filesystem registry
  As an operator
  I want to configure a local filesystem registry
  So that sauron can resolve artifacts from a directory

  Scenario: sets a filesystem registry from a local folder
    Given a filesystem registry
    And the filesystem registry hosts a skill named go-style
    When the user sets the filesystem registry from #{.folder.default.path}
    Then the command succeeds
    And the output contains registry set to
    And the output contains (filesystem)
    And there is exactly one registry
    And the registry has transport filesystem
    And the registry has a creation timestamp
    And the registry is described by:
      | field          | value                  |
      | kind           | Registry               |
      | apiVersion     | sauron.raitonbl.com/v1 |
      | spec.transport | filesystem             |

  Scenario: sets a filesystem registry from an authored content directory
    Given a filesystem registry
    And the filesystem registry hosts the directory testdata/registries/acme
    When the user sets the filesystem registry from #{.folder.default.path}
    Then the command succeeds
    And there is exactly one registry
    And the registry has transport filesystem

  # FR-007 — setting a registry replaces the one already configured; there is still
  # exactly one.
  Scenario: setting again replaces the configured registry
    Given a filesystem registry
    And the filesystem registry hosts a skill named go-style
    When the user sets the filesystem registry from #{.folder.default.path}
    And the user sets the filesystem registry from #{.folder.default.path}
    Then the command succeeds
    And there is exactly one registry

  Scenario: fails when the registry hosts no artifacts
    Given a filesystem registry
    When the user sets the filesystem registry from #{.folder.default.path}
    Then the command fails because the registry hosts no artifacts

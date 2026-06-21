@no-sandbox
Feature: Delete a registry
  As an operator
  I want to delete a registry I no longer use
  So that sauron stops resolving artifacts from it

  # FR-001 / FR-003 — the named registry is removed from registries.yaml, the other
  # registries remain, and an applied removal reports the summary count. The cascade
  # is a no-op here, so no artifacts are removed.
  Scenario: removes the named registry and leaves the others
    Given the following registries are configured:
      | name     | transport  | uri                               |
      | acme     | git        | git@github.com:acme/artifacts.git |
      | internal | http       | https://reg.example.com/          |
    When the user runs sauron delete registry acme
    Then the command succeeds
    And the removal summary reads registry "acme" removed; 0 artifacts removed
    And the stored state does not contain acme
    And a registry named internal exists
    And there are 1 registries

  # FR-005 — deleting a registry that does not exist exits 0 and reports that
  # nothing was deleted; the stored state is unchanged.
  Scenario: deleting an unknown registry reports nothing was deleted
    Given the following registries are configured:
      | name     | transport  | uri                               |
      | acme     | git        | git@github.com:acme/artifacts.git |
    When the user runs sauron delete registry ghost
    Then the command succeeds
    And the output reports nothing was deleted
    And a registry named acme exists
    And there are 1 registries

  # FR-004 — --dry-run prints the plan but writes nothing: the registry survives.
  Scenario: dry-run previews without changing state
    Given the following registries are configured:
      | name     | transport  | uri                               |
      | acme     | git        | git@github.com:acme/artifacts.git |
    When the user runs sauron delete registry acme --dry-run
    Then the command succeeds
    And a registry named acme exists
    And there are 1 registries

  # FR-006 — a missing <name> is a usage error (exit 2).
  Scenario: rejects a missing name
    Given the following registries are configured:
      | name     | transport  | uri                               |
      | acme     | git        | git@github.com:acme/artifacts.git |
    When the user runs sauron delete registry
    Then the command exits with status 2

  # write-then-read — the mandatory black-box arrange: add a registry, then delete it.
  Scenario: deletes a registry that was added through the command
    Given a filesystem registry
    And the filesystem registry hosts a skill named go-style
    When the user adds the filesystem registry acme from #{.folder.default.path}
    And the user runs sauron delete registry acme
    Then the command succeeds
    And the stored state does not contain acme
    And there are 0 registries

  # Deferred to 0007: requires install + the Action body; uncomment when artifact
  # removal lands. This scenario asserts the cascade actually removes a registry's
  # installed artifacts (FR-002/FR-003/FR-007); it cannot be arranged until install
  # exists and the shared cleaning step has a real body.
  # Scenario: cascade uninstalls every artifact the registry delivered
  #   Given the following registries are configured:
  #     | name | transport | uri                               |
  #     | acme | git       | git@github.com:acme/artifacts.git |
  #   And the registry acme has installed the skill sauron-acme-go-style
  #   And the registry acme has installed the agent sauron-acme-code-reviewer
  #   When the user runs sauron delete registry acme
  #   Then the command succeeds
  #   And the output contains skills:
  #   And the output contains   - sauron-acme-go-style
  #   And the output contains agents:
  #   And the output contains   - sauron-acme-code-reviewer
  #   And the removal summary reads registry "acme" removed; 2 artifacts removed
  #   And the skill sauron-acme-go-style is not installed
  #   And the agent sauron-acme-code-reviewer is not installed

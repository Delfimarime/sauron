@no-sandbox
Feature: Set provider
  As an operator
  I want to choose where artifacts are installed
  So that sauron records the single global provider destination

  # FR-001 — the first set records the single Provider document in settings.yaml
  # and confirms. Nothing is installed (install lands with 0007), so the migration
  # plan is empty and the summary carries no count.
  Scenario: records the provider on first set
    When the user runs sauron set provider claude
    Then the command succeeds
    And the output reports provider claude was set
    And the provider is set to claude

  # FR-003 — re-setting the already-active provider exits 0, reports no change, and
  # leaves the recorded provider as it was.
  Scenario: re-setting the active provider changes nothing
    When the user runs sauron set provider claude
    And the user runs sauron set provider claude
    Then the command succeeds
    And the output reports provider claude was already set
    And the provider is set to claude

  # FR-001 / FR-002 — switching to a different provider records the new one. Nothing
  # is installed, so the migration count is zero and the summary carries no count;
  # the structure (skills:/agents: groups) is present but empty.
  Scenario: switching to a different provider records the new provider
    When the user runs sauron set provider claude
    And the user runs sauron set provider zencoder
    Then the command succeeds
    And the output reports provider zencoder was set
    And the provider is set to zencoder

  # FR-004 — an unsupported provider name is a usage error (exit 2); the recorded
  # provider is left unchanged.
  Scenario: rejects an unsupported provider name
    When the user runs sauron set provider claude
    And the user runs sauron set provider bogus
    Then the command exits with status 2
    And the provider is set to claude

  # FR-006 — a missing argument exits with code 2 without executing the command.
  Scenario: rejects a missing argument
    When the user runs sauron set provider
    Then the command exits with status 2

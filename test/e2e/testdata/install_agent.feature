Feature: Install agent
  As a developer
  I want to install named agents from the registry into the active provider
  So that sauron fetches each agent and records it in the track file

  # FR-001/FR-002/FR-004 — install places the agent under the provider at
  # "agents/sauron-<name>" and records it in track.yaml with the source version
  # and exact path. The plan prints an "agents:" heading, a "+" entry, and a
  # summary count of added artifacts.
  Scenario: installs an agent and records it in the track file
    Given an http server hosting a registry
    And the http server hosts an agent named code-reviewer
    When the user sets the http registry from #{.webserver.default.url}
    And the user runs sauron set provider claude
    And the user runs sauron install agent code-reviewer
    Then the command succeeds
    And the output contains agents:
    And the output contains + sauron-code-reviewer
    And the output contains 1 added
    And the agent code-reviewer is tracked with a non-empty version
    And the agent code-reviewer has spec.path agents/sauron-code-reviewer

  # FR-003/FR-004 — if the tracked version matches the source version the agent
  # is already current: the plan shows no "+" entry for it and the run is a
  # no-op.
  Scenario: re-installing an unchanged agent is a no-op
    Given an http server hosting a registry
    And the http server hosts an agent named code-reviewer
    When the user sets the http registry from #{.webserver.default.url}
    And the user runs sauron set provider claude
    And the user runs sauron install agent code-reviewer
    And the user runs sauron install agent code-reviewer
    Then the command succeeds
    And the output does not contain + sauron-code-reviewer
    And the agent code-reviewer is tracked with a non-empty version

  # FR-006 — a name the registry does not offer is reported in the output without
  # stopping the run; offered sibling names are still installed and exit 0.
  Scenario: reports unknown agent names while installing the others
    Given an http server hosting a registry
    And the http server hosts an agent named code-reviewer
    When the user sets the http registry from #{.webserver.default.url}
    And the user runs sauron set provider claude
    And the user runs sauron install agent code-reviewer unknown-agent
    Then the command succeeds
    And the output contains unknown-agent
    And the output contains + sauron-code-reviewer
    And the agent code-reviewer is tracked with a non-empty version

  # FR-005 — when no provider is set, install fails with a runtime error (exit 1)
  # and installs nothing.
  Scenario: fails with a runtime error when no provider is set
    Given an http server hosting a registry
    And the http server hosts an agent named code-reviewer
    When the user sets the http registry from #{.webserver.default.url}
    And the user runs sauron install agent code-reviewer
    Then the command exits with status 1

  # FR-008 — a missing agent name is a usage error; the command exits 2 without
  # contacting the registry.
  Scenario: rejects a missing agent name
    When the user runs sauron install agent
    Then the command exits with status 2

Feature: Install skill
  As a developer
  I want to install named skills from the registry into the active provider
  So that sauron fetches each skill and records it in the track file

  # FR-001/FR-002/FR-004 — install places the skill under the provider at
  # "skills/sauron-<name>" and records it in track.yaml with the source version
  # and exact path. The plan prints a "skills:" heading, a "+" entry, and a
  # summary count of added artifacts.
  Scenario: installs a skill and records it in the track file
    Given an http server hosting a registry
    And the http server hosts a skill named go-style
    When the user sets the http registry from #{.webserver.default.url}
    And the user runs sauron set provider claude
    And the user runs sauron install skill go-style
    Then the command succeeds
    And the output contains skills:
    And the output contains + sauron-go-style
    And the output contains 1 added
    And the skill go-style is tracked with a non-empty version
    And the skill go-style has spec.path skills/sauron-go-style

  # FR-003/FR-004 — if the tracked version matches the source version the skill
  # is already current: the plan shows no "+" entry for it and the run is a
  # no-op.
  Scenario: re-installing an unchanged skill is a no-op
    Given an http server hosting a registry
    And the http server hosts a skill named go-style
    When the user sets the http registry from #{.webserver.default.url}
    And the user runs sauron set provider claude
    And the user runs sauron install skill go-style
    And the user runs sauron install skill go-style
    Then the command succeeds
    And the output does not contain + sauron-go-style
    And the skill go-style is tracked with a non-empty version

  # FR-003/FR-004 — if the tracked version differs from the source version the
  # skill is updated: the plan shows a "~" entry and the summary counts it as
  # updated. (A seeded track entry with version "seed" triggers the update when
  # the registry returns its fixed version "1.0.0".)
  Scenario: updates a skill when its version changed
    Given an http server hosting a registry
    And the http server hosts a skill named go-style
    And a tracked skill named go-style
    When the user sets the http registry from #{.webserver.default.url}
    And the user runs sauron set provider claude
    And the user runs sauron install skill go-style
    Then the command succeeds
    And the output contains skills:
    And the output contains ~ sauron-go-style
    And the output contains 1 updated
    And the skill go-style is tracked with a non-empty version
    And the skill go-style has spec.path skills/sauron-go-style

  # FR-006 — a name the registry does not offer is reported in the output without
  # stopping the run; offered sibling names are still installed and exit 0.
  Scenario: reports unknown skill names while installing the others
    Given an http server hosting a registry
    And the http server hosts a skill named go-style
    When the user sets the http registry from #{.webserver.default.url}
    And the user runs sauron set provider claude
    And the user runs sauron install skill go-style unknown-skill
    Then the command succeeds
    And the output contains unknown-skill
    And the output contains + sauron-go-style
    And the skill go-style is tracked with a non-empty version

  # FR-005 — when no provider is set, install fails with a runtime error (exit 1)
  # and installs nothing.
  Scenario: fails with a runtime error when no provider is set
    Given an http server hosting a registry
    And the http server hosts a skill named go-style
    When the user sets the http registry from #{.webserver.default.url}
    And the user runs sauron install skill go-style
    Then the command exits with status 1

  # FR-008 — a missing skill name is a usage error; the command exits 2 without
  # contacting the registry.
  Scenario: rejects a missing skill name
    When the user runs sauron install skill
    Then the command exits with status 2

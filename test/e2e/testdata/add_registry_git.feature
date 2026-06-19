@git
Feature: Add a git registry
  As an operator
  I want to register a git registry over ssh
  So that sauron can resolve artifacts from a git remote

  # @git is filtered out of the gate (godog --tags "~@git") until the ssh fixture
  # lands: the git source is deferred, so #{.git.default.url} errors.
  Scenario: adds a git registry over ssh
    Given a git server hosting a registry
    When the user adds the git registry acme from #{.git.default.url}
    Then the command succeeds

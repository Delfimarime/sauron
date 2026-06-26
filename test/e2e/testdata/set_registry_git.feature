@git
Feature: Set a git registry
  As an operator
  I want to configure a git registry over ssh
  So that sauron can resolve artifacts from a git remote

  Scenario: sets a git registry over ssh
    Given a git server hosting a registry
    And the git server hosts the directory testdata/registries/acme
    When the user sets the git registry from #{.git.default.url} using ssh key #{.git.default.sshKey}
    Then the command succeeds
    And the output contains (git)
    And there is exactly one registry
    And the registry has transport git

  Scenario: sets a git registry pinned to a ref
    Given a git server hosting a registry
    And the git server hosts the directory testdata/registries/acme
    When the user sets the git registry from #{.git.default.url} pinned to v1.0.0 using ssh key #{.git.default.sshKey}
    Then the command succeeds
    And the registry is described by:
      | field         | value  |
      | spec.revision | v1.0.0 |

  # Regression: a ref that is a commit SHA — neither a branch nor a tag — must
  # resolve via a full clone and checkout. Commit-addressed refs used to fail.
  Scenario: sets a git registry pinned to a commit
    Given a git server hosting a registry
    And the git server hosts the directory testdata/registries/acme
    When the user sets the git registry from #{.git.default.url} pinned to #{.git.default.revision} using ssh key #{.git.default.sshKey}
    Then the command succeeds
    And there is exactly one registry
    And the registry has transport git

  # Regression: the git transport cannot apply a client certificate, so it must
  # reject --client-cert/--client-key as a usage error (exit 2) before contacting
  # the remote, rather than silently accepting and persisting them.
  Scenario: rejects a client certificate, which the git transport cannot apply
    When the user sets the git registry from ssh://git@registry-git-default:22/home/git/acme.git with client certificate /tmp/client.crt and key /tmp/client.key
    Then the command exits with status 2

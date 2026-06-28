Feature: Set an http registry
  As an operator
  I want to configure a registry served over http
  So that sauron can resolve artifacts from a web endpoint

  Scenario: sets an http registry served over http
    Given an http server hosting a registry
    And the http server hosts the directory testdata/registries/acme
    When the user sets the http registry from #{.webserver.default.url}
    Then the command succeeds
    And the output contains registry set to
    And the output contains (http)
    And there is exactly one registry
    And the registry has transport http
    And the registry has a creation timestamp

  Scenario: sets an http registry behind basic auth, storing the secret as a reference
    Given an http server hosting a registry
    And the http server hosts a skill named go-style
    And the http server requires basic auth acme / ${env:ACME_TOKEN}
    When the user sets the http registry from #{.webserver.default.url} with username acme and password ${env:ACME_TOKEN}
    Then the command succeeds
    And the registry stores password as the reference ${env:ACME_TOKEN}
    And the stored state does not contain s3cr3t

  # FR-007 — setting a registry replaces the one already configured; there is still
  # exactly one.
  Scenario: setting again replaces the configured registry
    Given an http server hosting a registry
    And the http server hosts a skill named go-style
    When the user sets the http registry from #{.webserver.default.url}
    And the user sets the http registry from #{.webserver.default.url}
    Then the command succeeds
    And there is exactly one registry

  Scenario: fails when the registry hosts no artifacts
    Given an http server hosting a registry
    When the user sets the http registry from #{.webserver.default.url}
    Then the command fails because the registry hosts no artifacts

Feature: Add an http registry
  As an operator
  I want to register a registry served over http
  So that sauron can resolve artifacts from a web endpoint

  Scenario: adds an http registry served over http
    Given an http server hosting a registry
    And the http server hosts the directory testdata/registries/acme
    When the user adds the http registry acme from #{.webserver.default.url}
    Then the command succeeds
    And there is exactly one registry
    And a registry named acme exists
    And the registry acme has transport http

  Scenario: adds an http registry behind basic auth, storing the secret as a reference
    Given an http server hosting a registry
    And the http server hosts a skill named go-style
    And the http server requires basic auth acme / ${env:ACME_TOKEN}
    When the user adds the http registry acme from #{.webserver.default.url} with username acme and password ${env:ACME_TOKEN}
    Then the command succeeds
    And the registry acme stores password as the reference ${env:ACME_TOKEN}
    And the stored configuration does not contain s3cr3t

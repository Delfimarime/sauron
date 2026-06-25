@no-sandbox
Feature: Describe registry
  As an operator
  I want to see the configured registry's full detail
  So that I can review how sauron reaches its source

  # FR-001 — describe shows every field of the configured registry.
  Scenario: shows the full detail of the configured registry
    Given the registry is configured:
      | transport | uri                               | ref    | timeout |
      | git       | git@github.com:acme/artifacts.git | v1.2.0 | 45s     |
    When the user runs sauron describe registry
    Then the command succeeds
    And the descriptor shows transport as git
    And the descriptor shows uri as git@github.com:acme/artifacts.git
    And the descriptor shows ref as v1.2.0
    And the descriptor shows timeout as 45s

  # FR-003 — --fields selects and orders the displayed fields, uri first.
  Scenario: shows only the selected fields with --fields
    Given the registry is configured:
      | transport | uri                               |
      | git       | git@github.com:acme/artifacts.git |
    When the user runs sauron describe registry --fields transport,uri
    Then the command succeeds
    And the descriptor shows uri as git@github.com:acme/artifacts.git
    And the descriptor shows transport as git
    And the output does not contain ref

  # FR-002 — credential fields render as the stored ${env:…} reference, never a
  # resolved secret.
  Scenario: shows the auth block as environment references, never a secret
    Given the registry is configured:
      | transport | uri                               | username         | password          |
      | git       | git@github.com:acme/artifacts.git | ${env:ACME_USER} | ${env:ACME_TOKEN} |
    When the user runs sauron describe registry
    Then the command succeeds
    And the output contains auth:
    And the descriptor shows username as ${env:ACME_USER}
    And the descriptor shows password as ${env:ACME_TOKEN}
    And the output does not contain s3cr3t

  # FR-004 — no registry configured fails with a runtime error (exit 1).
  Scenario: fails with a runtime error when no registry is configured
    When the user runs sauron describe registry
    Then the command exits with status 1

  # FR-006 — describe surfaces the audit timestamps, and the full descriptor renders
  # verbatim: uri first (no name line), the aligned label: value column, and the
  # indented nested tls block exactly as the contract shows.
  Scenario: renders the full descriptor structure with timestamps, tls, and sshKey
    Given the settings file contains:
      """
      apiVersion: sauron.raitonbl.com/v1
      kind: Registry
      metadata:
        creationTimestamp: "2026-06-21T07:30:00Z"
        lastUpdatedTimestamp: "2026-06-22T08:00:00Z"
      spec:
        transport: git
        uri: git@github.com:acme/artifacts.git
        sshKey: /home/dev/.ssh/id_ed25519
        tls:
          skipVerify: true
          caCert: /etc/ssl/ca.pem
      """
    When the user runs sauron describe registry
    Then the command succeeds
    And the descriptor shows creationTimestamp as 2026-06-21T07:30:00Z
    And the descriptor shows lastUpdatedTimestamp as 2026-06-22T08:00:00Z
    And the descriptor shows sshKey as /home/dev/.ssh/id_ed25519
    And the descriptor shows caCert as /etc/ssl/ca.pem
    And the descriptor reads:
      """
      uri:                   git@github.com:acme/artifacts.git
      transport:             git
      tls:
        skipVerify:          true
        caCert:              /etc/ssl/ca.pem
      sshKey:                /home/dev/.ssh/id_ed25519
      creationTimestamp:     2026-06-21T07:30:00Z
      lastUpdatedTimestamp:  2026-06-22T08:00:00Z
      """

  # FR-005 — an invalid --fields value is a usage error (exit 2).
  Scenario: rejects an invalid --fields value
    Given the registry is configured:
      | transport | uri                               |
      | git       | git@github.com:acme/artifacts.git |
    When the user runs sauron describe registry --fields bogus
    Then the command exits with status 2

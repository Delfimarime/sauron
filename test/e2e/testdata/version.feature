Feature: Version banner
  As an operator
  I want sauron to report its version
  So that I can confirm which build I am running

  # Smoke scenario (graybox): execs the built binary located via SAURON_BIN and
  # asserts the --version banner fixed by the architecture contract's Root
  # command section:
  #
  #   <AppName> v<AppVersion>
  #   Hash <AppHash>
  #   Home: <home>
  Scenario: --version prints the banner
    When I run the binary with "--version"
    Then the command exits successfully
    And the output is a version banner

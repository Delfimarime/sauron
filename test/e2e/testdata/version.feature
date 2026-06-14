@no-sandbox
Feature: Version banner
  As an operator
  I want sauron to report its version
  So that I can confirm which build I am running
  # Graybox: hostRuntime execs the binary located via SAURON_BIN. The expected
  # banner is a Go template rendered against the world's Environment (build
  # identity, from env vars) and Variables (runtime values) maps.

  Scenario: --version is 0.0.0
    When sauron --version
    Then the output should be:
      """
      sauron v0.0.0-{{.Environment.App.Tag}}
      Hash {{.Environment.GitHash}}
      Home: {{.Variables.HomeDirectory}}
      """
    And sauron version is {{.Environment.App.Version}}-{{.Environment.App.Tag}}

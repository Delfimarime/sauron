@no-sandbox
Feature: Version
  As an operator
  I want sauron to report its build identity
  So that I can confirm which build I am running

  Scenario: reports its build identity
    Then sauron version is {{.App.FullVersion}}

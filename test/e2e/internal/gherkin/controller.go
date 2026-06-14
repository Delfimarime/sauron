package gherkin

import "github.com/cucumber/godog"

type Controller interface {
	Init(sc *godog.ScenarioContext)
}

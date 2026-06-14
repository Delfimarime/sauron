package gherkin

import (
	"github.com/cucumber/godog"
	"github.com/delfimarime/sauron/test/e2e/internal/runtime"
)

func Init(sc *godog.ScenarioContext, rt runtime.Runtime) {
	for _, each := range []Controller{
		&basicController{rt: rt},
	} {
		each.Init(sc)
	}
}

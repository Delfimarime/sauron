package gherkin

import (
	"github.com/cucumber/godog"
	"github.com/delfimarime/sauron/test/e2e/internal/runtime"
)

// Init registers every controller against the scenario context. Controllers hold
// only the runtime handle (rt); the runtime is the per-scenario shared state, so no
// "world" is threaded between them.
func Init(sc *godog.ScenarioContext, rt runtime.Runtime) {
	for _, each := range []Controller{
		&basicController{rt: rt},
		&commandController{rt: rt},
		&stateController{rt: rt},
		&listController{rt: rt},
		newRegistryFsController(rt),
		newRegistryHTTPController(rt),
		newRegistryGitController(rt),
	} {
		each.Init(sc)
	}
}

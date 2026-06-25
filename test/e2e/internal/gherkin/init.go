package gherkin

import (
	"github.com/cucumber/godog"
	"github.com/delfimarime/sauron/test/e2e/internal/runtime"
)

// Init registers every controller against the scenario context. Controllers hold
// only the runtime handle (rt); the runtime is the per-scenario shared state, so no
// "world" is threaded between them.
func Init(sc *godog.ScenarioContext, rt runtime.Runtime) {
	commands := &commandController{rt: rt}
	for _, each := range []Controller{
		&basicController{rt: rt},
		commands,
		&stateController{rt: rt},
		&listController{rt: rt},
		&describeController{command: commands},
		&deleteController{commands: commands},
		&catalogueController{rt: rt, command: commands},
		newRegistryFsController(rt),
		newRegistryHTTPController(rt),
		newRegistryGitController(rt),
	} {
		each.Init(sc)
	}
}

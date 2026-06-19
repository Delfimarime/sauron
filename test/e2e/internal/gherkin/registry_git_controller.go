package gherkin

import (
	"github.com/cucumber/godog"
	"github.com/delfimarime/sauron/test/e2e/internal/runtime"
)

// registryGitController declares the git-registry fixture. The git source is
// deferred (git remotes are ssh-only and the ssh fixture is not built yet):
// declaring it is harmless, but resolving #{.git.default.url} errors, so every git
// scenario carries @git and is filtered out of the gate until the ssh fixture
// lands. It reuses the shared declare step; it exposes no content (the source
// never materializes).
type registryGitController struct {
	sourceFixture
}

func newRegistryGitController(rt runtime.Runtime) *registryGitController {
	return &registryGitController{sourceFixture{
		source: func() runtime.Source { return rt.Git(defaultAlias) },
	}}
}

func (c *registryGitController) Init(sc *godog.ScenarioContext) {
	sc.Step(`^a git server hosting a registry$`, c.declare)
}

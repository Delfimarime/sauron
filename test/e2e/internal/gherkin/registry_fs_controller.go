package gherkin

import (
	"github.com/cucumber/godog"
	"github.com/delfimarime/sauron/test/e2e/internal/runtime"
)

// registryFsController owns the filesystem-registry fixture steps, translating them
// into a Folder source declaration. The registry uri is then #{.folder.default.path}.
type registryFsController struct {
	sourceFixture
}

func newRegistryFsController(rt runtime.Runtime) *registryFsController {
	return &registryFsController{sourceFixture{
		source: func() runtime.Source { return rt.Folder(defaultAlias) },
	}}
}

func (c *registryFsController) Init(sc *godog.ScenarioContext) {
	sc.Step(`^a filesystem registry$`, c.declare)
	sc.Step(`^the filesystem registry hosts a skill named (\S+)$`, c.hostsSkill)
	sc.Step(`^the filesystem registry hosts the directory (\S+)$`, c.hostsDirectory)
	sc.Step(`^the filesystem registry hosts the file (\S+) as (\S+)$`, c.hostsFile)
}

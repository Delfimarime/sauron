package gherkin

import (
	"github.com/cucumber/godog"
	"github.com/delfimarime/sauron/test/e2e/internal/runtime"
)

// registryGitController owns the git-registry fixture steps, translating them into a
// Git source declaration (an sshd sidecar serving a seeded bare repo under docker).
// The registry uri is then #{.git.default.url}. It reuses the shared content steps so
// the same provider content set is exposed over git as over folder and http.
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
	sc.Step(`^the git server hosts a skill named (\S+)$`, c.hostsSkill)
	sc.Step(`^the git server hosts the directory (\S+)$`, c.hostsDirectory)
	sc.Step(`^the git server hosts the file (\S+) as (\S+)$`, c.hostsFile)
}

package gherkin

import (
	"context"

	"github.com/cucumber/godog"
	"github.com/delfimarime/sauron/test/e2e/internal/runtime"
)

// registryHTTPController owns the http-registry fixture steps, translating them into
// a Webserver source declaration (an nginx sidecar under docker). The registry uri is
// then #{.webserver.default.url}. It adds the basic-auth step on top of the shared
// content steps.
type registryHTTPController struct {
	sourceFixture
}

func newRegistryHTTPController(rt runtime.Runtime) *registryHTTPController {
	return &registryHTTPController{sourceFixture{
		source: func() runtime.Source { return rt.Webserver(defaultAlias) },
	}}
}

func (c *registryHTTPController) Init(sc *godog.ScenarioContext) {
	sc.Step(`^an http server hosting a registry$`, c.declare)
	sc.Step(`^the http server hosts a skill named (\S+)$`, c.hostsSkill)
	sc.Step(`^the http server hosts the directory (\S+)$`, c.hostsDirectory)
	sc.Step(`^the http server hosts the file (\S+) as (\S+)$`, c.hostsFile)
	sc.Step(`^the http server requires basic auth (\S+) / (\S+)$`, c.requiresBasicAuth)
}

func (c *registryHTTPController) requiresBasicAuth(_ context.Context, username, password string) error {
	c.source().Expose(runtime.Resource{Username: username, Password: password})
	return nil
}

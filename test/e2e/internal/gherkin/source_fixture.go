package gherkin

import (
	"context"

	"github.com/delfimarime/sauron/test/e2e/internal/runtime"
)

// sourceFixture is the shared behaviour of the registry source fixtures: declare a
// runtime source and expose provider content on it. The http and git controllers
// differ only in which source they select (and their Gherkin wording), so the
// translation logic lives here once. The source selector closes over the
// runtime and the capability, so a fixture never threads rt itself.
type sourceFixture struct {
	source func() runtime.Source
}

// declare names the source; materialization is deferred to the first need.
func (f *sourceFixture) declare(context.Context) error {
	f.source()
	return nil
}

func (f *sourceFixture) hostsSkill(_ context.Context, name string) error {
	f.source().Expose(skillResource(name))
	return nil
}

// hostsDirectory exposes an authored testdata directory (a skill/agent/persona
// content set), preserving its layout.
func (f *sourceFixture) hostsDirectory(_ context.Context, path string) error {
	return exposeDirectory(f.source(), path)
}

// hostsFile exposes a single authored testdata file at served.
func (f *sourceFixture) hostsFile(_ context.Context, path, served string) error {
	return exposeFile(f.source(), path, served)
}

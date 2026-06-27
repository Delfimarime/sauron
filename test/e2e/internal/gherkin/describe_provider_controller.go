package gherkin

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// noProviderSetMessage is the FR-003 none-set report: describing a provider when
// none is configured exits 0 and says so.
const noProviderSetMessage = "no provider is set"

// describeProviderController owns the one describe-provider assertion no generic
// step covers: the FR-003 none-set message. The descriptor structure, field
// selection, and audit/sync timestamps are asserted by the shared describeController
// steps. It reads the recorded command output through the commandController that
// captured it, so the result stays consumed in one place rather than re-run here.
type describeProviderController struct {
	commands *commandController
}

func (c *describeProviderController) Init(sc *godog.ScenarioContext) {
	sc.Step(`^the output reports no provider is set$`, c.reportsNoProviderSet)
}

// reportsNoProviderSet asserts the FR-003 message is present in the command output.
func (c *describeProviderController) reportsNoProviderSet(_ context.Context) error {
	out, err := c.commands.lastOutput()
	if err != nil {
		return err
	}
	if !strings.Contains(out, noProviderSetMessage) {
		return fmt.Errorf("output does not report %q; got: %s", noProviderSetMessage, out)
	}
	return nil
}

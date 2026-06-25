package gherkin

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// nothingUnsetMessage is the FR-005 idempotent-unset report: unsetting when no
// registry is configured exits 0 and says so.
const nothingUnsetMessage = "nothing was unset"

// unsetController owns the unset-registry output assertion no generic step covers:
// the FR-005 nothing-was-unset message. It reads the recorded command output through
// the commandController that captured it, so the result stays consumed in one place
// rather than re-run here.
type unsetController struct {
	commands *commandController
}

func (c *unsetController) Init(sc *godog.ScenarioContext) {
	sc.Step(`^the output reports nothing was unset$`, c.reportsNothingUnset)
}

// reportsNothingUnset asserts the FR-005 message is present in the command output.
func (c *unsetController) reportsNothingUnset(_ context.Context) error {
	out, err := c.commands.lastOutput()
	if err != nil {
		return err
	}
	if !strings.Contains(out, nothingUnsetMessage) {
		return fmt.Errorf("output does not report %q; got: %s", nothingUnsetMessage, out)
	}
	return nil
}

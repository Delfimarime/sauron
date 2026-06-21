package gherkin

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// nothingDeletedMessage is the FR-005 idempotent-delete report: deleting an absent
// registry exits 0 and says so.
const nothingDeletedMessage = "nothing was deleted"

// deleteController owns the delete-registry output assertions no existing step
// covers: the FR-005 nothing-was-deleted message and the removal summary line. It
// reads the recorded command output through the commandController that captured it,
// so the result stays consumed in one place rather than re-run here.
type deleteController struct {
	commands *commandController
}

func (c *deleteController) Init(sc *godog.ScenarioContext) {
	sc.Step(`^the output reports nothing was deleted$`, c.reportsNothingDeleted)
	sc.Step(`^the removal summary reads (.+)$`, c.removalSummaryReads)
}

// reportsNothingDeleted asserts the FR-005 message is present in the command output.
func (c *deleteController) reportsNothingDeleted(_ context.Context) error {
	out, err := c.commands.lastOutput()
	if err != nil {
		return err
	}
	if !strings.Contains(out, nothingDeletedMessage) {
		return fmt.Errorf("output does not report %q; got: %s", nothingDeletedMessage, out)
	}
	return nil
}

// removalSummaryReads asserts the command output contains the exact summary line.
func (c *deleteController) removalSummaryReads(_ context.Context, summary string) error {
	out, err := c.commands.lastOutput()
	if err != nil {
		return err
	}
	if !strings.Contains(out, summary) {
		return fmt.Errorf("output does not contain the removal summary %q; got: %s", summary, out)
	}
	return nil
}

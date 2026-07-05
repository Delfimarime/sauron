package gherkin

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
)

// listArtifactsController owns the `list skills`/`list agents` row assertion. The
// track.yaml fixture is seeded by the shared `the track file contains:` step
// (seed_controller.go); paging is asserted through the existing generic
// `the paging line reads (.+)$` step (registered once by catalogueController and
// reused here verbatim, so both list families report paging identically). It reads
// the recorded output of the last command from the commandController rather than
// re-running anything, so the rendered rows are asserted exactly as the user saw
// them.
type listArtifactsController struct {
	command *commandController
}

func (c *listArtifactsController) Init(sc *godog.ScenarioContext) {
	sc.Step(`^the artifacts list (.+)$`, c.artifactsList)
}

// artifactsList asserts a row is present in the rendered `list skills`/`list
// agents` table. The argument is the whitespace-separated row the table renders,
// e.g. "go-style" (the default, name-only view) or
// "go-style v1.4.0 2026-06-15T00:00:00Z" (name/version/lastUpdatedAt, or any
// --fields selection/order).
func (c *listArtifactsController) artifactsList(_ context.Context, row string) error {
	output, err := c.command.lastOutput()
	if err != nil {
		return err
	}
	if !hasRow(output, row) {
		return fmt.Errorf("artifact list does not list %q; got: %s", row, output)
	}
	return nil
}

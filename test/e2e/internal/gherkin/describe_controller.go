package gherkin

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// describeController owns the describe-detail Then assertions. It reads the
// recorded output of the last command from the commandController rather than
// re-running anything, so the descriptor is asserted exactly as the user saw it.
type describeController struct {
	command *commandController
}

func (c *describeController) Init(sc *godog.ScenarioContext) {
	sc.Step(`^the descriptor shows (.+) as (.+)$`, c.descriptorShows)
	sc.Step(`^the output does not contain (.+)$`, c.outputDoesNotContain)
	sc.Step(`^the descriptor reads:$`, c.descriptorReads)
}

// descriptorReads asserts the descriptor output matches the expected block
// verbatim — proving field ordering, the aligned "label: value" column, and the
// indented nested auth/tls blocks render exactly as the contract shows. Both sides
// are normalized only by trimming a trailing newline so the doc-string need not be
// byte-perfect on its final line.
func (c *describeController) descriptorReads(_ context.Context, expected *godog.DocString) error {
	if err := c.command.requireRun(); err != nil {
		return err
	}
	want := strings.TrimRight(expected.Content, "\n")
	got := strings.TrimRight(c.command.last.output, "\n")
	if want != got {
		return fmt.Errorf("descriptor output mismatch.\n--- expected ---\n%s\n--- got ---\n%s", want, got)
	}
	return nil
}

// descriptorShows reads a "label: value" line from the descriptor and asserts the
// pair. Nested fields (e.g. the auth block's username/password) are indented, so
// the label is matched on its trimmed leading token.
func (c *describeController) descriptorShows(_ context.Context, label, value string) error {
	if err := c.command.requireRun(); err != nil {
		return err
	}
	got, ok := descriptorValue(c.command.last.output, label)
	if !ok {
		return fmt.Errorf("descriptor has no %q field; got: %s", label, c.command.last.output)
	}
	return assertExpected(fmt.Sprintf("descriptor field %q", label), value, got)
}

// outputDoesNotContain asserts a value never appears in the last command's output;
// it drives FR-002 (a resolved secret is never shown) and field projection.
func (c *describeController) outputDoesNotContain(_ context.Context, text string) error {
	if err := c.command.requireRun(); err != nil {
		return err
	}
	if strings.Contains(c.command.last.output, text) {
		return fmt.Errorf("output unexpectedly contains %q; got: %s", text, c.command.last.output)
	}
	return nil
}

// descriptorValue finds the descriptor line labelled label and returns its value
// (everything after the first colon, trimmed). Pure (text in, value out) so it is
// unit-tested without a process.
func descriptorValue(output, label string) (string, bool) {
	want := label + ":"
	for raw := range strings.SplitSeq(output, "\n") {
		line := strings.TrimSpace(raw)
		if !strings.HasPrefix(line, want) {
			continue
		}
		return strings.TrimSpace(strings.TrimPrefix(line, want)), true
	}
	return "", false
}

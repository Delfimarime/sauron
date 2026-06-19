package gherkin

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
	"github.com/delfimarime/sauron/test/e2e/internal/runtime"
)

// commandController owns the action under test (When) and the command-outcome
// assertions (Then). It holds the last command result itself: that result is
// consumed only here, so keeping it controller-local does not reconstitute a shared
// "world" — cross-controller state (source addresses) stays on the runtime.
type commandController struct {
	rt   runtime.Runtime
	last *commandResult
}

type commandResult struct {
	code   int
	output string
}

func (c *commandController) Init(sc *godog.ScenarioContext) {
	sc.Step(`^the user runs (.+)$`, c.userRuns)
	sc.Step(`^the user adds the registry:$`, c.addRegistryFromTable)
	sc.Step(`^the user adds the (filesystem|http|git) registry (\S+) from (\S+)$`, c.addRegistry)
	sc.Step(`^the user adds the (filesystem|http|git) registry (\S+) from (\S+) with username (\S+) and password (\S+)$`, c.addRegistryWithAuth)

	sc.Step(`^the command succeeds$`, c.succeeds)
	sc.Step(`^the command exits with status (\d+)$`, c.exitsWith)
	sc.Step(`^the output contains (.+)$`, c.outputContains)
	sc.Step(`^the command fails because the registry hosts no artifacts$`, c.failsNoArtifacts)
}

// userRuns is the escape hatch: it runs a raw command line through the runtime.
func (c *commandController) userRuns(ctx context.Context, line string) error {
	return c.run(ctx, strings.Fields(line))
}

// addRegistry runs `sauron add registry` for one transport, resolving the source
// reference (#{…}) to a concrete uri.
func (c *commandController) addRegistry(ctx context.Context, transport, name, uriRef string) error {
	return c.add(ctx, transport, name, uriRef, "", "")
}

// addRegistryWithAuth is addRegistry with basic-auth credentials forwarded to the
// command (the binary stores them as ${env:VAR} references).
func (c *commandController) addRegistryWithAuth(ctx context.Context, transport, name, uriRef, username, password string) error {
	return c.add(ctx, transport, name, uriRef, username, password)
}

// addRegistryFromTable is the canonical table-driven form; the uri cell is resolved
// through valueOf like the sugar steps.
func (c *commandController) addRegistryFromTable(ctx context.Context, table *godog.Table) error {
	fields, err := tableFields(table)
	if err != nil {
		return err
	}
	return c.add(ctx, fields["transport"], fields["name"], fields["uri"], fields["username"], fields["password"])
}

// add is the shared body of every "adds the registry" step: resolve the uri
// reference, then run `sauron add registry` with the optional basic-auth flags.
func (c *commandController) add(ctx context.Context, transport, name, uriRef, username, password string) error {
	uri, err := valueOf[string](ctx, c.rt, uriRef)
	if err != nil {
		return err
	}
	return c.run(ctx, addRegistryArgs(name, transport, uri, username, password))
}

func (c *commandController) succeeds(context.Context) error {
	if err := c.requireRun(); err != nil {
		return err
	}
	if c.last.code != 0 {
		return fmt.Errorf("expected success but command exited %d: %s", c.last.code, c.last.output)
	}
	return nil
}

func (c *commandController) exitsWith(_ context.Context, code int) error {
	if err := c.requireRun(); err != nil {
		return err
	}
	return assertExpected("exit status", code, c.last.code)
}

func (c *commandController) outputContains(_ context.Context, text string) error {
	if err := c.requireRun(); err != nil {
		return err
	}
	if !strings.Contains(c.last.output, text) {
		return fmt.Errorf("output does not contain %q; got: %s", text, c.last.output)
	}
	return nil
}

func (c *commandController) failsNoArtifacts(context.Context) error {
	if err := c.requireRun(); err != nil {
		return err
	}
	if c.last.code == 0 {
		return fmt.Errorf("expected a failure because the registry hosts no artifacts, but the command succeeded")
	}
	return nil
}

// run executes args, recording the result for the outcome assertions. A non-zero
// exit is recorded, not raised; only a harness-level failure returns an error.
func (c *commandController) run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no command to run")
	}
	code, out, err := c.rt.Execute(ctx, args...)
	if err != nil {
		return fmt.Errorf("execute %q: %w", strings.Join(args, " "), err)
	}
	c.last = &commandResult{code: code, output: out}
	return nil
}

func (c *commandController) requireRun() error {
	if c.last == nil {
		return fmt.Errorf("no command has run yet")
	}
	return nil
}

// addRegistryArgs assembles the `sauron add registry` invocation shared by every
// When step, appending the optional basic-auth flags only when set.
func addRegistryArgs(name, transport, uri, username, password string) []string {
	args := []string{"sauron", "add", "registry", "--kind", transport, "--name", name, "--uri", uri}
	if username != "" {
		args = append(args, "--username", username)
	}
	if password != "" {
		args = append(args, "--password", password)
	}
	return args
}

// tableFields flattens a two-column |field|value| table into a map, skipping the
// header row.
func tableFields(table *godog.Table) (map[string]string, error) {
	out := map[string]string{}
	for i, row := range table.Rows {
		if i == 0 {
			continue // header: | field | value |
		}
		if len(row.Cells) != 2 {
			return nil, fmt.Errorf("row %d: expected 2 cells (field, value), got %d", i, len(row.Cells))
		}
		out[row.Cells[0].Value] = row.Cells[1].Value
	}
	return out, nil
}

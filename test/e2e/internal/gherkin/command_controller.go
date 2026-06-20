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
	sc.Step(`^the user adds the (filesystem|http|git) registry (\S+) from (\S+) pinned to (\S+)$`, c.addRegistryPinned)
	sc.Step(`^the user adds the (filesystem|http|git) registry (\S+) from (\S+) with username (\S+) and password (\S+)$`, c.addRegistryWithAuth)
	sc.Step(`^the user adds the (filesystem|http|git) registry (\S+) from (\S+) using ssh key (\S+)$`, c.addRegistryWithSSHKey)
	sc.Step(`^the user adds the (filesystem|http|git) registry (\S+) from (\S+) pinned to (\S+) using ssh key (\S+)$`, c.addRegistryPinnedWithSSHKey)

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
	return c.add(ctx, addOptions{transport: transport, name: name, uriRef: uriRef})
}

// addRegistryPinned is addRegistry with the source pinned to a git ref, forwarded
// as --ref (the binary records it as spec.ref).
func (c *commandController) addRegistryPinned(ctx context.Context, transport, name, uriRef, ref string) error {
	return c.add(ctx, addOptions{transport: transport, name: name, uriRef: uriRef, ref: ref})
}

// addRegistryWithAuth is addRegistry with basic-auth credentials forwarded to the
// command (the binary stores the password as a ${env:VAR} reference).
func (c *commandController) addRegistryWithAuth(ctx context.Context, transport, name, uriRef, username, password string) error {
	return c.add(ctx, addOptions{transport: transport, name: name, uriRef: uriRef, username: username, password: password})
}

// addRegistryWithSSHKey is addRegistry with an ssh key forwarded as --ssh-key, so
// the git transport authenticates with the key rather than the ssh agent. The key
// argument is resolved through valueOf (a #{.git.<alias>.sshKey} reference yields
// the in-runtime key path).
func (c *commandController) addRegistryWithSSHKey(ctx context.Context, transport, name, uriRef, sshKeyRef string) error {
	return c.add(ctx, addOptions{transport: transport, name: name, uriRef: uriRef, sshKeyRef: sshKeyRef})
}

// addRegistryPinnedWithSSHKey is the pinned git form authenticated with an ssh key.
func (c *commandController) addRegistryPinnedWithSSHKey(ctx context.Context, transport, name, uriRef, ref, sshKeyRef string) error {
	return c.add(ctx, addOptions{transport: transport, name: name, uriRef: uriRef, ref: ref, sshKeyRef: sshKeyRef})
}

// addRegistryFromTable is the canonical table-driven form; the uri cell is resolved
// through valueOf like the sugar steps.
func (c *commandController) addRegistryFromTable(ctx context.Context, table *godog.Table) error {
	fields, err := tableFields(table)
	if err != nil {
		return err
	}
	return c.add(ctx, addOptions{
		transport: fields["transport"],
		name:      fields["name"],
		uriRef:    fields["uri"],
		ref:       fields["ref"],
		username:  fields["username"],
		password:  fields["password"],
		sshKeyRef: fields["sshKey"],
	})
}

// addOptions carries the fields every "adds the registry" step contributes; the
// optional ones default to empty and are forwarded only when set.
type addOptions struct {
	transport string
	name      string
	uriRef    string
	ref       string
	username  string
	password  string
	sshKeyRef string
}

// add is the shared body of every "adds the registry" step: resolve the uri (and
// the ssh-key path, when given) references, then run `sauron add registry` with the
// optional ref, basic-auth, and ssh-key flags.
func (c *commandController) add(ctx context.Context, o addOptions) error {
	uri, err := valueOf[string](ctx, c.rt, o.uriRef)
	if err != nil {
		return err
	}
	sshKey := o.sshKeyRef
	if sshKey != "" {
		sshKey, err = valueOf[string](ctx, c.rt, o.sshKeyRef)
		if err != nil {
			return err
		}
	}
	return c.run(ctx, addRegistryArgs(o.name, o.transport, uri, o.ref, o.username, o.password, sshKey))
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
// When step. The command takes the transport as --kind and the name and uri as
// positional arguments; the ref, basic-auth, and ssh-key flags are appended only
// when set, before the positionals.
func addRegistryArgs(name, transport, uri, ref, username, password, sshKey string) []string {
	args := []string{"sauron", "add", "registry", "--kind", transport}
	if ref != "" {
		args = append(args, "--ref", ref)
	}
	if username != "" {
		args = append(args, "--username", username)
	}
	if password != "" {
		args = append(args, "--password", password)
	}
	if sshKey != "" {
		args = append(args, "--ssh-key", sshKey)
	}
	return append(args, name, uri)
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

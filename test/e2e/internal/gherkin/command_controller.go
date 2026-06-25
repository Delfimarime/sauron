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
	sc.Step(`^the user sets the (filesystem|http|git) registry from (\S+)$`, c.setRegistry)
	sc.Step(`^the user sets the (filesystem|http|git) registry from (\S+) pinned to (\S+)$`, c.setRegistryPinned)
	sc.Step(`^the user sets the (filesystem|http|git) registry from (\S+) with username (\S+) and password (\S+)$`, c.setRegistryWithAuth)
	sc.Step(`^the user sets the (filesystem|http|git) registry from (\S+) using ssh key (\S+)$`, c.setRegistryWithSSHKey)
	sc.Step(`^the user sets the (filesystem|http|git) registry from (\S+) pinned to (\S+) using ssh key (\S+)$`, c.setRegistryPinnedWithSSHKey)
	sc.Step(`^the user sets the (filesystem|http|git) registry from (\S+) with client certificate (\S+) and key (\S+)$`, c.setRegistryWithClientCert)

	sc.Step(`^the command succeeds$`, c.succeeds)
	sc.Step(`^the command exits with status (\d+)$`, c.exitsWith)
	sc.Step(`^the output contains (.+)$`, c.outputContains)
	sc.Step(`^the output is empty$`, c.outputEmpty)
	sc.Step(`^the command fails because the registry hosts no artifacts$`, c.failsNoArtifacts)
}

// userRuns is the escape hatch: it runs a raw command line through the runtime.
func (c *commandController) userRuns(ctx context.Context, line string) error {
	return c.run(ctx, strings.Fields(line))
}

// setRegistry runs `sauron set registry` for one transport, resolving the source
// reference (#{…}) to a concrete uri. The single registry carries no name: it is
// identified by its uri, and setting one replaces any registry already configured.
func (c *commandController) setRegistry(ctx context.Context, transport, uriRef string) error {
	return c.set(ctx, setOptions{transport: transport, uriRef: uriRef})
}

// setRegistryPinned is setRegistry with the source pinned to a git ref, forwarded
// as --ref (the binary records it as spec.ref).
func (c *commandController) setRegistryPinned(ctx context.Context, transport, uriRef, ref string) error {
	return c.set(ctx, setOptions{transport: transport, uriRef: uriRef, ref: ref})
}

// setRegistryWithAuth is setRegistry with basic-auth credentials forwarded to the
// command (the binary stores the password as a ${env:VAR} reference).
func (c *commandController) setRegistryWithAuth(ctx context.Context, transport, uriRef, username, password string) error {
	return c.set(ctx, setOptions{transport: transport, uriRef: uriRef, username: username, password: password})
}

// setRegistryWithSSHKey is setRegistry with an ssh key forwarded as --ssh-key, so
// the git transport authenticates with the key rather than the ssh agent. The key
// argument is resolved through valueOf (a #{.git.<alias>.sshKey} reference yields
// the in-runtime key path).
func (c *commandController) setRegistryWithSSHKey(ctx context.Context, transport, uriRef, sshKeyRef string) error {
	return c.set(ctx, setOptions{transport: transport, uriRef: uriRef, sshKeyRef: sshKeyRef})
}

// setRegistryPinnedWithSSHKey is the pinned git form authenticated with an ssh key.
func (c *commandController) setRegistryPinnedWithSSHKey(ctx context.Context, transport, uriRef, ref, sshKeyRef string) error {
	return c.set(ctx, setOptions{transport: transport, uriRef: uriRef, ref: ref, sshKeyRef: sshKeyRef})
}

// setRegistryWithClientCert is setRegistry forwarding --client-cert/--client-key,
// the mutual-TLS flags the git transport rejects (it cannot apply a client
// certificate); the cert and key arguments are paths the command never reads
// before rejecting them.
func (c *commandController) setRegistryWithClientCert(ctx context.Context, transport, uriRef, clientCert, clientKey string) error {
	return c.set(ctx, setOptions{transport: transport, uriRef: uriRef, clientCert: clientCert, clientKey: clientKey})
}

// setOptions carries the fields every "sets the registry" step contributes; the
// optional ones default to empty and are forwarded only when set.
type setOptions struct {
	transport  string
	uriRef     string
	ref        string
	username   string
	password   string
	sshKeyRef  string
	clientCert string
	clientKey  string
}

// set is the shared body of every "sets the registry" step: resolve the uri, and
// the ssh-key and ref references when given (a ref may be a #{.git.<alias>.revision}
// placeholder), then run `sauron set registry` with the optional ref, basic-auth,
// and ssh-key flags.
func (c *commandController) set(ctx context.Context, o setOptions) error {
	uri, err := valueOf(ctx, c.rt, o.uriRef)
	if err != nil {
		return err
	}
	sshKey := o.sshKeyRef
	if sshKey != "" {
		sshKey, err = valueOf(ctx, c.rt, o.sshKeyRef)
		if err != nil {
			return err
		}
	}
	ref := o.ref
	if ref != "" {
		ref, err = valueOf(ctx, c.rt, o.ref)
		if err != nil {
			return err
		}
	}
	o.uriRef = uri
	o.sshKeyRef = sshKey
	o.ref = ref
	return c.run(ctx, setRegistryArgs(o))
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

// outputEmpty asserts the last command produced no non-whitespace output.
func (c *commandController) outputEmpty(context.Context) error {
	if err := c.requireRun(); err != nil {
		return err
	}
	if strings.TrimSpace(c.last.output) != "" {
		return fmt.Errorf("expected empty output but got: %s", c.last.output)
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

// lastOutput returns the output of the most recent command, for sibling
// controllers that assert on it. It errors when no command has run.
func (c *commandController) lastOutput() (string, error) {
	if err := c.requireRun(); err != nil {
		return "", err
	}
	return c.last.output, nil
}

// setRegistryArgs assembles the `sauron set registry` invocation shared by every
// When step. It takes the option struct with its uri and ssh-key references already
// resolved to concrete values. The command takes the transport as --kind and the
// uri as the sole positional argument (the single registry carries no name); the
// ref, basic-auth, ssh-key, and client-certificate flags are appended only when
// set, before the positional.
func setRegistryArgs(o setOptions) []string {
	args := []string{"sauron", "set", "registry", "--kind", o.transport}
	if o.ref != "" {
		args = append(args, "--ref", o.ref)
	}
	if o.username != "" {
		args = append(args, "--username", o.username)
	}
	if o.password != "" {
		args = append(args, "--password", o.password)
	}
	if o.sshKeyRef != "" {
		args = append(args, "--ssh-key", o.sshKeyRef)
	}
	if o.clientCert != "" {
		args = append(args, "--client-cert", o.clientCert)
	}
	if o.clientKey != "" {
		args = append(args, "--client-key", o.clientKey)
	}
	return append(args, o.uriRef)
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

package e2e

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/delfimarime/sauron/test/e2e/internal/runtime"
)

type compositionBasedRuntime struct {
	isStartup *atomic.Bool
	backedBy  runtime.Runtime
}

// IsReadOnly reports whether the runtime must not be mutated by a scenario.
// The host OS is read-only (true); ephemeral sandboxes are not (false).
func (c *compositionBasedRuntime) IsReadOnly() bool {
	if c.backedBy == nil {
		return true
	}
	if !c.isStartup.Load() {
		return true
	}
	return c.backedBy.IsReadOnly()
}

// Execute runs the binary under test with command as its args, returning the
// exit code, the relevant output stream (stdout on success, stderr on a
// non-zero exit), and an error ONLY for harness-level failures (the process or
// container exec could not run). A non-zero exit is not an error.
func (c *compositionBasedRuntime) Execute(ctx context.Context, args ...string) (int, string, error) {
	if c.backedBy == nil {
		return -1, "", fmt.Errorf("runtime has not be defined")
	}
	if !c.isStartup.Load() {
		return -1, "", fmt.Errorf("Cannot execute command on runtime that hasn't been started up")
	}
	return c.backedBy.Execute(ctx, args...)
}

func (c *compositionBasedRuntime) Start(ctx context.Context) error {
	if c.backedBy == nil {
		return fmt.Errorf("runtime has not be defined")
	}
	if c.isStartup.Load() {
		return nil
	}
	return c.backedBy.Start(ctx)
}

func (c *compositionBasedRuntime) Stop(ctx context.Context) error {
	if c.backedBy == nil {
		return fmt.Errorf("runtime has not be defined")
	}
	if !c.isStartup.Load() {
		return nil
	}
	return c.backedBy.Stop(ctx)
}

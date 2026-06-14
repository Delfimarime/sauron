package e2e

import (
	"context"
	"fmt"

	"github.com/delfimarime/sauron/test/e2e/internal/runtime"
)

// compositionBasedRuntime is a guarded facade over a runtime.Runtime. Step
// definitions hold one stable handle for the whole scenario; the backend is
// attached per scenario (in the suite's Before hook) and started lazily on the
// first Execute, so a scenario that never runs a command never starts a sandbox.
type compositionBasedRuntime struct {
	started  bool
	backedBy runtime.Runtime
}

// IsReadOnly reports the backend's read-only flag, defaulting to read-only (the
// safe assumption) until a backend is attached.
func (c *compositionBasedRuntime) IsReadOnly() bool {
	if c.backedBy == nil {
		return true
	}
	return c.backedBy.IsReadOnly()
}

// Start starts the backend at most once; further calls are no-ops.
func (c *compositionBasedRuntime) Start(ctx context.Context) error {
	if c.backedBy == nil {
		return fmt.Errorf("runtime backend is not set")
	}
	if c.started {
		return nil
	}
	if err := c.backedBy.Start(ctx); err != nil {
		return err
	}
	c.started = true
	return nil
}

// Execute starts the backend on first use, then runs the command through it.
func (c *compositionBasedRuntime) Execute(ctx context.Context, args ...string) (int, string, error) {
	if c.backedBy == nil {
		return -1, "", fmt.Errorf("runtime backend is not set")
	}
	if !c.started {
		if err := c.Start(ctx); err != nil {
			return -1, "", err
		}
	}
	return c.backedBy.Execute(ctx, args...)
}

// Stop tears down the backend only if it was started.
func (c *compositionBasedRuntime) Stop(ctx context.Context) error {
	if c.backedBy == nil || !c.started {
		return nil
	}
	return c.backedBy.Stop(ctx)
}

package e2e

import (
	"context"
	"fmt"

	"github.com/delfimarime/sauron/test/e2e/internal/runtime"
)

// compositionBasedRuntime is a guarded facade over a runtime.Runtime. Step
// definitions hold one stable handle for the whole scenario; the backend is
// attached per scenario (in the suite's Before hook) and started lazily on the
// first Execute OR the first source-attribute access, so a scenario that never
// needs the sandbox never starts one.
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

func (c *compositionBasedRuntime) ReadFile(ctx context.Context, path string) ([]byte, error) {
	if c.backedBy == nil {
		return nil, fmt.Errorf("runtime backend is not set")
	}
	return c.backedBy.ReadFile(ctx, path)
}

func (c *compositionBasedRuntime) CopyTo(ctx context.Context, locationURI string, content []byte) error {
	if c.backedBy == nil {
		return fmt.Errorf("runtime backend is not set")
	}
	return c.backedBy.CopyTo(ctx, locationURI, content)
}

// Folder/Webserver/Git return a lazy source: Expose accumulates against the stable
// backend source without starting anything, while Path/URL force the lazy Start
// first (resolving #{…} is what materializes every accumulated sidecar).
func (c *compositionBasedRuntime) Folder(alias string) runtime.Source {
	return c.lazy(func() runtime.Source { return c.backedBy.Folder(alias) })
}

func (c *compositionBasedRuntime) Webserver(alias string) runtime.Source {
	return c.lazy(func() runtime.Source { return c.backedBy.Webserver(alias) })
}

func (c *compositionBasedRuntime) Git(alias string) runtime.Source {
	return c.lazy(func() runtime.Source { return c.backedBy.Git(alias) })
}

func (c *compositionBasedRuntime) lazy(fetch func() runtime.Source) runtime.Source {
	return &lazySource{rt: c, fetch: func() runtime.Source {
		if c.backedBy == nil {
			return runtime.NewErroringSource(fmt.Errorf("runtime backend is not set"))
		}
		return fetch()
	}}
}

// Stop tears down the backend only if it was started.
func (c *compositionBasedRuntime) Stop(ctx context.Context) error {
	if c.backedBy == nil || !c.started {
		return nil
	}
	return c.backedBy.Stop(ctx)
}

// lazySource defers materialization: it forwards Expose to the backend source as a
// pure accumulation, but Path/URL force the runtime's single lazy Start before
// reading the now-live address.
type lazySource struct {
	rt    *compositionBasedRuntime
	fetch func() runtime.Source
}

func (s *lazySource) Expose(resources ...runtime.Resource) { s.fetch().Expose(resources...) }

func (s *lazySource) Path(ctx context.Context) (string, error) {
	if err := s.rt.Start(ctx); err != nil {
		return "", err
	}
	return s.fetch().Path(ctx)
}

func (s *lazySource) URL(ctx context.Context) (string, error) {
	if err := s.rt.Start(ctx); err != nil {
		return "", err
	}
	return s.fetch().URL(ctx)
}

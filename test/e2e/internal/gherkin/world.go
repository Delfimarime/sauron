// Package gherkin holds the e2e harness's godog step definitions, the
// per-scenario World, and the expected-output templating. It is graybox: it runs
// the binary under test through a docker.Runtime and asserts on its output via the
// public pkg/ surface, never importing the production internal/ packages.
package gherkin

import (
	"context"
	"errors"
	"os"

	"github.com/delfimarime/sauron/test/e2e/internal/docker"
)

const envSauronBin = "SAURON_BIN"

// result captures one Execute.
type result struct {
	exitCode int
	output   string
}

// World is the per-scenario state shared across step definitions. Environment is
// build identity (from env vars); Variables is runtime-provided (e.g. home). Both
// are the data templates render against.
type World struct {
	bin         string
	Environment map[string]any
	Variables   map[string]any
	specs       []docker.ContainerSpec
	runtime     docker.Runtime
	last        *result
}

// NewWorld resolves the binary and build identity from the process environment.
func NewWorld() (*World, error) {
	return newWorld(os.Getenv(envSauronBin), os.Getenv)
}

// newWorld builds a world from an explicit bin and getenv, so tests need not
// touch the real process environment.
func newWorld(bin string, getenv func(string) string) (*World, error) {
	if bin == "" {
		return nil, errors.New(envSauronBin + " is not set; the gate-integration task must point it at the built binary")
	}

	env, err := newEnvironment(getenv)
	if err != nil {
		return nil, err
	}

	return &World{
		bin:         bin,
		Environment: env,
		Variables:   newVariables(getenv),
	}, nil
}

// Declare appends a container spec a Given step needs; nothing starts yet.
func (w *World) Declare(spec docker.ContainerSpec) {
	w.specs = append(w.specs, spec)
}

// Execute runs the binary under test through the lazily-built runtime (chosen by
// the declared spec count), recording the result as the last invocation.
func (w *World) Execute(ctx context.Context, args ...string) error {
	if w.runtime == nil {
		rt := docker.SelectRuntime(w.bin, w.specs)
		if err := rt.Start(ctx); err != nil {
			return err
		}
		w.runtime = rt
	}

	code, out, err := w.runtime.Execute(ctx, args...)
	if err != nil {
		return err
	}

	w.last = &result{exitCode: code, output: out}
	return nil
}

// Last returns the most recent invocation, or nil when none has run.
func (w *World) Last() *result { return w.last }

// Reset stops any runtime and clears per-scenario state.
func (w *World) Reset(ctx context.Context) error {
	var err error
	if w.runtime != nil {
		err = w.runtime.Stop(ctx)
		w.runtime = nil
	}
	w.specs = nil
	w.last = nil
	return err
}

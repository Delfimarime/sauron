// Package gherkin holds the e2e harness's godog step definitions, the
// per-scenario World, and the expected-output templating. It is graybox: it runs
// the binary under test through a runtime.Runtime and asserts on its output via
// the public pkg/ surface, never importing the production internal/ packages.
package gherkin

import (
	"context"
	"errors"
	"os"

	"github.com/delfimarime/sauron/test/e2e/internal/runtime"
	"github.com/delfimarime/sauron/test/e2e/internal/runtime/docker"
	"github.com/delfimarime/sauron/test/e2e/internal/runtime/host"
)

const envSauronBin = "SAURON_BIN"

// noSandboxTag marks scenarios that run against the host OS rather than a
// compose sandbox.
const noSandboxTag = "@no-sandbox"

// result captures one Execute.
type result struct {
	exitCode int
	output   string
}

// World is the per-scenario state shared across step definitions. Environment is
// build identity (from env vars); Variables is runtime-provided (e.g. home). Both
// are the data templates render against. useSandbox selects the runtime: false
// (a @no-sandbox scenario) runs on the host, true runs in a compose sandbox.
type World struct {
	bin         string
	Environment map[string]any
	Variables   map[string]any
	useSandbox  bool
	specs       []docker.ContainerSpec
	rt          runtime.Runtime
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

// wantsSandbox reports whether a scenario's tags ask for a sandbox runtime: every
// scenario does unless it carries @no-sandbox.
func wantsSandbox(tagNames []string) bool {
	for _, name := range tagNames {
		if name == noSandboxTag {
			return false
		}
	}
	return true
}

// Declare appends a dependency container spec a Given step needs; nothing starts
// yet, and it is only realized when the scenario runs in a sandbox.
func (w *World) Declare(spec docker.ContainerSpec) {
	w.specs = append(w.specs, spec)
}

// buildRuntime selects the runtime for the scenario without starting it: the host
// OS for @no-sandbox scenarios, otherwise a compose sandbox with the declared
// dependency services.
func (w *World) buildRuntime() (runtime.Runtime, error) {
	if !w.useSandbox {
		return host.New(w.bin), nil
	}

	dir, err := os.MkdirTemp("", "sauron-e2e-")
	if err != nil {
		return nil, err
	}
	return docker.New(w.bin, dir, docker.WithContainer(w.specs...))
}

// Execute runs the binary under test through the lazily-built runtime, recording
// the result as the last invocation.
func (w *World) Execute(ctx context.Context, args ...string) error {
	if w.rt == nil {
		rt, err := w.buildRuntime()
		if err != nil {
			return err
		}
		if err := rt.Start(ctx); err != nil {
			return err
		}
		w.rt = rt
	}

	code, out, err := w.rt.Execute(ctx, args...)
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
	if w.rt != nil {
		err = w.rt.Stop(ctx)
		w.rt = nil
	}
	w.specs = nil
	w.last = nil
	return err
}

// Package gherkin holds the e2e harness's godog step definitions, the
// per-scenario World, and the expected-output templating. It is graybox: it runs
// the binary under test through an injected runtime.Runtime and asserts on its
// output via the public pkg/ surface, never importing the production internal/
// packages.
package gherkin

import (
	"context"

	"github.com/delfimarime/sauron/test/e2e/internal/runtime"
)

// result captures one Execute.
type result struct {
	exitCode int
	output   string
}

// World is the per-scenario state shared across step definitions. Environment is
// build identity (from env vars); Variables is runtime-provided (e.g. home). Both
// are the data templates render against. Commands run through rt, the runtime the
// suite selected for the scenario.
type World struct {
	Environment map[string]any
	Variables   map[string]any
	rt          runtime.Runtime
	last        *result
}

// newWorld builds a world bound to rt, reading build identity through getenv so
// tests need not touch the real process environment.
func newWorld(rt runtime.Runtime, getenv func(string) string) (*World, error) {
	env, err := newEnvironment(getenv)
	if err != nil {
		return nil, err
	}

	return &World{
		Environment: env,
		Variables:   newVariables(getenv),
		rt:          rt,
	}, nil
}

// Execute runs the binary under test through the runtime, recording the result as
// the last invocation.
func (w *World) Execute(ctx context.Context, args ...string) error {
	code, out, err := w.rt.Execute(ctx, args...)
	if err != nil {
		return err
	}

	w.last = &result{exitCode: code, output: out}
	return nil
}

// Last returns the most recent invocation, or nil when none has run.
func (w *World) Last() *result { return w.last }

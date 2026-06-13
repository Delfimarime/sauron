// Package internal holds the e2e harness internals: the per-scenario world,
// the built-binary runner, and the godog step definitions.
//
// This package is the test harness and is deliberately exempt from the
// Use Case/Action and ports-and-adapters rules that govern production code.
// It is the graybox external consumer of the public pkg/ surface: it execs the
// built binary located via SAURON_BIN and asserts on its output. It imports
// only pkg/, never internal/ — a rule Go's internal/ mechanism cannot enforce
// here (the import paths share the root module prefix), so it is enforced by a
// depguard rule in this module's .golangci.yml.
package internal

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
)

// envSauronBin is the environment variable that carries the absolute path to
// the built binary under test, injected by the gate-integration task.
const envSauronBin = "SAURON_BIN"

// CommandResult captures one invocation of the binary.
type CommandResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// World is the per-scenario state shared across step definitions. Repeated
// setup is factored here rather than copy-pasted across steps.
type World struct {
	// bin is the absolute path to the binary under test.
	bin string
	// last holds the result of the most recent invocation.
	last *CommandResult
}

// NewWorld resolves the binary under test from SAURON_BIN and returns a fresh
// world. It returns an error when SAURON_BIN is unset, so a misconfigured gate
// fails loudly rather than silently exercising nothing.
func NewWorld() (*World, error) {
	bin := os.Getenv(envSauronBin)
	if bin == "" {
		return nil, errors.New(envSauronBin + " is not set; the gate-integration task must point it at the built binary")
	}

	return &World{bin: bin}, nil
}

// Reset clears per-scenario state between scenarios.
func (w *World) Reset() {
	w.last = nil
}

// Run execs the binary under test with args, capturing stdout, stderr and the
// exit code, and records the result as the last invocation.
func (w *World) Run(ctx context.Context, args ...string) error {
	var stdout, stderr bytes.Buffer

	cmd := exec.CommandContext(ctx, w.bin, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	exitCode := 0
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			return err
		}
		exitCode = exitErr.ExitCode()
	}

	w.last = &CommandResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
	}

	return nil
}

// Last returns the most recent invocation result, or nil when no command has
// run in the current scenario.
func (w *World) Last() *CommandResult {
	return w.last
}

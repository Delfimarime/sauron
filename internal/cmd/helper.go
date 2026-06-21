package cmd

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"

	"github.com/delfimarime/sauron/internal/config"
	"github.com/delfimarime/sauron/internal/telemetry"
	"github.com/delfimarime/sauron/internal/usecase"
	"github.com/spf13/cobra"
)

// errInvalidFlag marks a malformed flag value the command layer rejects before
// the use case runs; it maps to the usage exit code.
var errInvalidFlag = errors.New("invalid flag")

// exit-code conventions shared by every command.
const (
	exitOK    = 0
	exitError = 1
	exitUsage = 2
)

// exitCode maps a command error to the process exit code: caller mistakes
// (usage errors and flag parse failures) yield exitUsage, every other failure
// yields exitError, and a nil error yields exitOK.
func exitCode(err error) int {
	if err == nil {
		return exitOK
	}

	var ucErr *usecase.Error
	if errors.As(err, &ucErr) {
		if ucErr.Type == usecase.TypeUsage {
			return exitUsage
		}
		return exitError
	}
	if errors.Is(err, errInvalidFlag) {
		return exitUsage
	}

	return exitError
}

// NewApp builds (but does not start) the fx app wired with the modules transversal
// to every command, then appends the caller's command-specific opts.
func NewApp(ctx context.Context, opts ...fx.Option) *fx.App {
	base := make([]fx.Option, 0, 8+len(opts))
	base = append(base,
		fx.Provide(func() context.Context { return ctx }),
		telemetry.NewFxOptions(),
		config.NewFxOptions(),
		fx.Provide(
			newPondPool,
		),
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
	)
	return fx.New(append(base, opts...)...)
}

// usageArgs wraps a cobra positional-args validator so a violation is classified
// as a usage error (exit code 2) rather than a generic failure.
func usageArgs(validate cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if err := validate(cmd, args); err != nil {
			return fmt.Errorf("%w: %w", errInvalidFlag, err)
		}
		return nil
	}
}

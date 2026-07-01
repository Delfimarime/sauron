package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"

	"github.com/delfimarime/sauron/internal/config"
	"github.com/delfimarime/sauron/internal/infrastructure/repository"
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

// ExitCode maps a command error to the process exit code: caller mistakes
// (usage errors and flag parse failures) yield exitUsage, every other failure
// yields exitError, and a nil error yields exitOK.
func ExitCode(err error) int {
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

// silenceFlagErrors installs the shared flag-error handler on cmd: a parse
// failure is wrapped with errInvalidFlag so it maps to exit code 2 (usage).
func silenceFlagErrors(cmd *cobra.Command) {
	cmd.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return fmt.Errorf("%w: %w", errInvalidFlag, err)
	})
}

// commandOption configures one piece of a cobra.Command built by newCommand.
type commandOption func(*cobra.Command)

// withLong sets the command's long description.
func withLong(long string) commandOption {
	return func(c *cobra.Command) { c.Long = long }
}

// withArgs sets the positional-args policy, wrapped so a violation is a usage
// error. A command that omits it carries no restriction — the shape a pure
// group needs, since cobra routes an unmatched subcommand argument to it.
func withArgs(args cobra.PositionalArgs) commandOption {
	return func(c *cobra.Command) { c.Args = usageArgs(args) }
}

// withFlags binds the command's own flags.
func withFlags(bind func(*cobra.Command)) commandOption {
	return func(c *cobra.Command) { bind(c) }
}

// withRunE wires the cobra-free handler as the command's RunE: run is invoked
// with the positional args, a live context, and the command's stdout.
func withRunE(run func(ctx context.Context, args []string, stdout io.Writer) error) commandOption {
	return func(c *cobra.Command) {
		c.RunE = func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), args, cmd.OutOrStdout())
		}
	}
}

// withSubcommands attaches subs as children — the shape a pure group command
// uses in place of withRunE; the two are never combined on one command.
func withSubcommands(subs ...*cobra.Command) commandOption {
	return func(c *cobra.Command) { c.AddCommand(subs...) }
}

// newCommand builds one cobra command — leaf or group — with the shared
// scaffold every command needs: silenced usage/errors and flag-error wrapping.
// use and short are positional since every command needs both; everything
// else — the long description, the args policy, flags, the handler, or
// subcommands — is an option, individually omittable, so this one constructor
// serves both a leaf command (withArgs/withFlags/withRunE) and a pure group
// (withSubcommands only).
func newCommand(use, short string, opts ...commandOption) *cobra.Command {
	cmd := &cobra.Command{
		Use:           use,
		Short:         short,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	silenceFlagErrors(cmd)
	for _, opt := range opts {
		opt(cmd)
	}

	return cmd
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

// runUseCase builds a minimal fx app on a cancellable run context, resolves the
// use case U, runs exec against it to produce a *P result, and tears the app down
// (cancel before Stop); exec receives the run context so it binds its work to the
// same lifecycle. The result is presentation-agnostic — the caller renders it.
func runUseCase[U, P any](ctx context.Context, exec func(context.Context, U) (*P, error), opts ...fx.Option) (*P, error) {
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var product *P
	var prob error
	app := NewApp(runCtx, append([]fx.Option{
		repository.NewFxOptions(),
		usecase.NewFxOptions(),
		fx.Invoke(func(uc U) { product, prob = exec(runCtx, uc) }),
	}, opts...)...)
	if err := app.Err(); err != nil {
		return nil, fmt.Errorf("build application: %w", err)
	}
	if err := app.Start(runCtx); err != nil {
		return nil, fmt.Errorf("start application: %w", err)
	}
	cancel()
	_ = app.Stop(context.WithoutCancel(ctx))

	return product, prob
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

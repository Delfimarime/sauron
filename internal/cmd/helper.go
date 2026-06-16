package cmd

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"

	"github.com/delfimarime/sauron/internal/config"
	"github.com/delfimarime/sauron/internal/infrastructure/repository"
	"github.com/delfimarime/sauron/internal/telemetry"
	"github.com/delfimarime/sauron/internal/usecase"
)

// NewApp builds (but does not start) the transversal fx app and appends the caller's opts.
func NewApp(ctx context.Context, opts ...fx.Option) *fx.App {
	base := make([]fx.Option, 0, 10+len(opts))
	base = append(base,
		fx.Supply(ctx),
		telemetry.NewFxOptions(),
		config.NewFxOptions(),
		repository.NewFxOptions(),
		usecase.NewFxOptions(),
		fx.Provide(
			newPondPool,
		),
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
	)
	return fx.New(append(base, opts...)...)
}
